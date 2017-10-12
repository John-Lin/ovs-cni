package main

import (
	"encoding/json"
	"fmt"
	"runtime"

	"github.com/containernetworking/cni/pkg/skel"
	"github.com/containernetworking/cni/pkg/types"
	"github.com/containernetworking/cni/pkg/types/current"
	"github.com/containernetworking/cni/pkg/version"
	"github.com/containernetworking/plugins/pkg/ip"
	"github.com/containernetworking/plugins/pkg/ipam"
	"github.com/containernetworking/plugins/pkg/ns"
	log "github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
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

func bridgeByName(name string) (*netlink.Bridge, error) {
	l, err := netlink.LinkByName(name)
	if err != nil {
		return nil, fmt.Errorf("could not lookup %q: %v", name, err)
	}
	br, ok := l.(*netlink.Bridge)
	if !ok {
		return nil, fmt.Errorf("%q already exists but is not a bridge", name)
	}
	return br, nil
}

func ensureBridge(brName string) (*netlink.Bridge, error) {
	ovsbr, err := NewOVSSwitch(brName)
	if err != nil {
		log.Fatal("failed to NewOVSSwitch: ", err)
		return nil, fmt.Errorf("failed to create bridge %q: %v", brName, err)
	}

	// Re-fetch link to read all attributes and if it already existed,
	// ensure it's really a bridge with similar configuration
	br, err := bridgeByName(brName)
	if err != nil {
		return nil, err
	}

	return br, nil
}

func setupVeth(netns ns.NetNS, br *netlink.Bridge, ifName string, mtu int, hairpinMode bool) (*current.Interface, *current.Interface, error) {
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
		return nil
	})
	if err != nil {
		return nil, nil, err
	}

	// need to lookup hostVeth again as its index has changed during ns move
	hostVeth, err := netlink.LinkByName(hostIface.Name)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to lookup %q: %v", hostIface.Name, err)
	}
	hostIface.Mac = hostVeth.Attrs().HardwareAddr.String()

	// connect host veth end to the bridge
	if err := netlink.LinkSetMaster(hostVeth, br); err != nil {
		return nil, nil, fmt.Errorf("failed to connect %q to bridge %v: %v", hostVeth.Attrs().Name, br.Attrs().Name, err)
	}

	return hostIface, contIface, nil
}

func setupBridge(n *NetConf) (*netlink.Bridge, *current.Interface, error) {
	// create bridge if necessary
	br, err := ensureBridge(n.OVSBrName)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create bridge %q: %v", n.OVSBrName, err)
	}

	return br, &current.Interface{
		Name: br.Attrs().Name,
	}, nil
}

func cmdAdd(args *skel.CmdArgs) error {
	n, cniVersion, err := loadNetConf(args.StdinData)
	if err != nil {
		return err
	}

	br, brInterface, err := setupBridge(n)
	if err != nil {
		return err
	}

	netns, err := ns.GetNS(args.Netns)
	if err != nil {
		return fmt.Errorf("failed to open netns %q: %v", args.Netns, err)
	}
	defer netns.Close()

	_ = netns
	_ = br
	_ = brInterface

	// hostInterface, containerInterface, err := setupVeth(netns, br, args.IfName, n.MTU, n.HairpinMode)
	// if err != nil {
	// 	return err
	// }

	// run the IPAM plugin and get back the config to apply
	// r, err := ipam.ExecAdd(n.IPAM.Type, args.StdinData)
	// if err != nil {
	// 	return err
	// }

	// Convert whatever the IPAM result was into the current Result type
	// result, err := current.NewResultFromResult(r)
	// if err != nil {
	// 	return err
	// }

	// if len(result.IPs) == 0 {
	// 	return errors.New("IPAM plugin returned missing IP config")
	// }

	// result.Interfaces = []*current.Interface{brInterface, hostInterface, containerInterface}

	// // Configure the container hardware address and IP address(es)
	// if err := netns.Do(func(_ ns.NetNS) error {
	// 	contVeth, err := net.InterfaceByName(args.IfName)
	// 	if err != nil {
	// 		return err
	// 	}

	// 	// Add the IP to the interface
	// 	if err := ipam.ConfigureIface(args.IfName, result); err != nil {
	// 		return err
	// 	}

	// 	// Send a gratuitous arp
	// 	for _, ipc := range result.IPs {
	// 		if ipc.Version == "4" {
	// 			_ = arping.GratuitousArpOverIface(ipc.Address.IP, *contVeth)
	// 		}
	// 	}
	// 	return nil
	// }); err != nil {
	// 	return err
	// }

	return nil
	// return types.PrintResult(result, cniVersion)
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

	// return err
}

func main() {
	skel.PluginMain(cmdAdd, cmdDel, version.All)
}
