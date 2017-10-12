package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"runtime"

	"github.com/containernetworking/cni/pkg/skel"
	"github.com/containernetworking/cni/pkg/types"
	"github.com/containernetworking/cni/pkg/types/current"
	"github.com/containernetworking/cni/pkg/version"
	"github.com/containernetworking/plugins/pkg/ip"
	"github.com/containernetworking/plugins/pkg/ipam"
	"github.com/containernetworking/plugins/pkg/ns"
	"github.com/j-keck/arping"
	log "github.com/sirupsen/logrus"
)

const defaultBrName = "br0"

type NetConf struct {
	types.NetConf
	IsMaster   bool     `json:"isMaster"`
	OVSBrName  string   `json:"ovsBridge"`
	VtepIPs    []string `json:"vtepIPs"`
	Controller string   `json:"controller,omitempty"`
}

func init() {
	// this ensures that main runs only on main thread (thread group leader).
	// since namespace ops (unshare, setns) are done for a single thread, we
	// must ensure that the goroutine does not jump from OS thread to thread
	runtime.LockOSThread()
}

func loadNetConf(bytes []byte) (*NetConf, string, error) {
	n := &NetConf{
		OVSBrName: defaultBrName,
	}
	if err := json.Unmarshal(bytes, n); err != nil {
		return nil, "", fmt.Errorf("failed to load netconf: %v", err)
	}
	return n, n.CNIVersion, nil
}

func ensureBridge(brName string) (*OVSSwitch, error) {
	ovsbr, err := NewOVSSwitch(brName)
	if err != nil {
		log.Fatal("failed to NewOVSSwitch: ", err)
		return nil, fmt.Errorf("failed to ensure bridge %q: %v", brName, err)
	}
	return ovsbr, nil
}

func setupVeth(netns ns.NetNS, br *OVSSwitch, ifName string, mtu int) (*current.Interface, *current.Interface, error) {
	contIface := &current.Interface{}
	hostIface := &current.Interface{}

	err := netns.Do(func(hostNS ns.NetNS) error {
		// create the veth pair in the container and move host end into host netns
		hostVeth, containerVeth, err := ip.SetupVeth(ifName, mtu, hostNS)
		if err != nil {
			return err
		}
		contIface.Name = containerVeth.Name
		contIface.Mac = containerVeth.HardwareAddr.String()
		contIface.Sandbox = netns.Path()
		hostIface.Name = hostVeth.Name

		// ip link set lo up
		_, err = ifaceUp("lo")
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, nil, err
	}

	err = br.addPort(hostIface.Name)
	if err != nil {
		log.Fatalf("failed to addPort switch - host: %v", err)
	}
	log.Infof("%s Adding a link:", br.BridgeName)

	return hostIface, contIface, nil
}

func setupBridge(n *NetConf) (*OVSSwitch, *current.Interface, error) {
	// create bridge if necessary
	ovsbr, err := ensureBridge(n.OVSBrName)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to setup bridge %q: %v", n.OVSBrName, err)
	}

	return ovsbr, &current.Interface{
		Name: ovsbr.BridgeName,
	}, nil
}

func cmdAdd(args *skel.CmdArgs) error {
	n, cniVersion, err := loadNetConf(args.StdinData)
	if err != nil {
		return err
	}

	// Create a Open vSwitch bridge
	br, brInterface, err := setupBridge(n)
	if err != nil {
		return err
	}

	netns, err := ns.GetNS(args.Netns)
	if err != nil {
		return fmt.Errorf("failed to open netns %q: %v", args.Netns, err)
	}
	defer netns.Close()

	hostInterface, containerInterface, err := setupVeth(netns, br, args.IfName, 1500)
	if err != nil {
		return err
	}

	// run the IPAM plugin and get back the config to apply
	r, err := ipam.ExecAdd(n.IPAM.Type, args.StdinData)
	if err != nil {
		return err
	}

	// Convert whatever the IPAM result was into the current Result type
	result, err := current.NewResultFromResult(r)
	if err != nil {
		return err
	}

	if len(result.IPs) == 0 {
		return errors.New("IPAM plugin returned missing IP config")
	}

	result.Interfaces = []*current.Interface{brInterface, hostInterface, containerInterface}

	// Configure the container hardware address and IP address(es)
	if err := netns.Do(func(_ ns.NetNS) error {
		contVeth, err := net.InterfaceByName(args.IfName)
		if err != nil {
			return err
		}

		// Add the IP to the interface
		for _, ipc := range result.IPs {
			// All IPs currently refer to the container interface
			// 0 -> bridge itself
			// 1 -> veth endpoint
			// 2 -> interface in Container.
			ipc.Interface = current.Int(2)
		}
		if err := ipam.ConfigureIface(args.IfName, result); err != nil {
			return err
		}

		// Send a gratuitous arp
		for _, ipc := range result.IPs {
			if ipc.Version == "4" {
				_ = arping.GratuitousArpOverIface(ipc.Address.IP, *contVeth)
			}
		}
		return nil
	}); err != nil {
		return err
	}

	return types.PrintResult(result, cniVersion)
}

func cmdDel(args *skel.CmdArgs) error {
	n, _, err := loadNetConf(args.StdinData)
	if err != nil {
		return err
	}

	if err := ipam.ExecDel(n.IPAM.Type, args.StdinData); err != nil {
		return err
	}

	if args.Netns == "" {
		return nil
	}

	// There is a netns so try to clean up. Delete can be called multiple times
	// so don't return an error if the device is already removed.
	// If the device isn't there then don't try to clean up IP masq either.
	// var ipnets []*net.IPNet
	// err = ns.WithNetNSPath(args.Netns, func(_ ns.NetNS) error {
	// 	var err error
	// 	ipnets, err = ip.DelLinkByNameAddr(args.IfName)
	// 	if err != nil && err == ip.ErrLinkNotFound {
	// 		return nil
	// 	}
	// 	return err
	// })

	// if err != nil {
	// 	return err
	// }

	return nil
}

func main() {
	skel.PluginMain(cmdAdd, cmdDel, version.All)
}
