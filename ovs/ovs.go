package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"runtime"
	"syscall"

	"github.com/containernetworking/cni/pkg/skel"
	"github.com/containernetworking/cni/pkg/types"
	"github.com/containernetworking/cni/pkg/types/current"
	"github.com/containernetworking/cni/pkg/version"
	"github.com/containernetworking/plugins/pkg/ip"
	"github.com/containernetworking/plugins/pkg/ipam"
	"github.com/containernetworking/plugins/pkg/ns"
	"github.com/containernetworking/plugins/pkg/utils"

	"github.com/j-keck/arping"
	log "github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
)

const defaultBrName = "br0"

type NetConf struct {
	types.NetConf
	OVSBrName   string   `json:"ovsBridge"`
	IsGW        bool     `json:"isGateway"`
	IsDefaultGW bool     `json:"isDefaultGateway"`
	IPMasq      bool     `json:"ipMasq"`
	VtepIPs     []string `json:"vtepIPs"`
	Controller  string   `json:"controller,omitempty"`
}

type gwInfo struct {
	gws               []net.IPNet
	family            int
	defaultRouteFound bool
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

// calcGateways processes the results from the IPAM plugin and does the
// following for each IP family:
//    - Calculates and compiles a list of gateway addresses
//    - Adds a default route if needed
func calcGateways(result *current.Result, n *NetConf) (*gwInfo, *gwInfo, error) {

	gwsV4 := &gwInfo{}
	gwsV6 := &gwInfo{}

	for _, ipc := range result.IPs {

		// Determine if this config is IPv4 or IPv6
		var gws *gwInfo
		defaultNet := &net.IPNet{}
		switch {
		case ipc.Address.IP.To4() != nil:
			gws = gwsV4
			gws.family = netlink.FAMILY_V4
			defaultNet.IP = net.IPv4zero
		case len(ipc.Address.IP) == net.IPv6len:
			gws = gwsV6
			gws.family = netlink.FAMILY_V6
			defaultNet.IP = net.IPv6zero
		default:
			return nil, nil, fmt.Errorf("Unknown IP object: %v", ipc)
		}
		defaultNet.Mask = net.IPMask(defaultNet.IP)

		// All IPs currently refer to the container interface
		// Add the IP to the interface
		// 0 -> bridge itself
		// 1 -> veth endpoint
		// 2 -> interface in container
		ipc.Interface = current.Int(2)

		// If not provided, calculate the gateway address
		// We use first address of specific subnet as its gateway
		if ipc.Gateway == nil && n.IsGW {
			ipc.Gateway = getNextIP(&ipc.Address)
		}

		// Add a default route for this family using the current
		// gateway address if necessary.
		if n.IsDefaultGW && !gws.defaultRouteFound {
			for _, route := range result.Routes {
				if route.GW != nil && defaultNet.String() == route.Dst.String() {
					gws.defaultRouteFound = true
					break
				}
			}
			if !gws.defaultRouteFound {
				result.Routes = append(
					result.Routes,
					&types.Route{Dst: *defaultNet, GW: ipc.Gateway},
				)
				gws.defaultRouteFound = true
			}
		}

		// Append this gateway address to the list of gateways
		if n.IsGW {
			gw := net.IPNet{
				IP:   ipc.Gateway,
				Mask: ipc.Address.Mask,
			}
			gws.gws = append(gws.gws, gw)
		}
	}
	return gwsV4, gwsV6, nil
}

func ensureBridgeAddr(br *OVSSwitch, family int, ipn *net.IPNet) error {
	ovsbrLink, err := netlink.LinkByName(br.BridgeName)
	if err != nil {
		return fmt.Errorf("could not get ovs bridge link: %v", err)
	}

	addrs, err := netlink.AddrList(ovsbrLink, family)
	if err != nil && err != syscall.ENOENT {
		return fmt.Errorf("could not get list of IP addresses: %v", err)
	}

	ipnStr := ipn.String()
	for _, a := range addrs {
		// string comp is actually easiest for doing IPNet comps
		if a.IPNet.String() == ipnStr {
			return nil
		}
	}

	addr := &netlink.Addr{IPNet: ipn, Label: ""}
	if err := netlink.AddrAdd(ovsbrLink, addr); err != nil {
		return fmt.Errorf("could not add IP address to %q: %v", br.BridgeName, err)
	}

	// set ovs link device up
	if err := setLinkUp(br.BridgeName); err != nil {
		return fmt.Errorf("Error setting link %s up. Err: %v", br.BridgeName, err)
	}
	return nil
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
		err = setLinkUp("lo")
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

func cmdAdd(args *skel.CmdArgs) error {
	n, cniVersion, err := loadNetConf(args.StdinData)
	if err != nil {
		return err
	}

	if n.IsDefaultGW {
		n.IsGW = true
	}

	// Create a Open vSwitch bridge
	br, brInterface, err := createOVS(n)
	if err != nil {
		return err
	}

	netns, err := ns.GetNS(args.Netns)
	if err != nil {
		return fmt.Errorf("failed to open netns %q: %v", args.Netns, err)
	}
	defer netns.Close()

	hostInterface, containerInterface, err := setupVeth(netns, br, args.IfName, 1400)
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

	// Gather gateway information for each IP family
	gwsV4, gwsV6, err := calcGateways(result, n)
	if err != nil {
		return err
	}

	// Configure the container hardware address and IP address(es)
	if err := netns.Do(func(_ ns.NetNS) error {
		contVeth, err := net.InterfaceByName(args.IfName)
		if err != nil {
			return err
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

	if n.IsGW {
		var firstV4Addr net.IP
		// Set the IP address(es) on the bridge and enable forwarding
		for _, gws := range []*gwInfo{gwsV4, gwsV6} {
			for _, gw := range gws.gws {
				if gw.IP.To4() != nil && firstV4Addr == nil {
					firstV4Addr = gw.IP
				}

				err = ensureBridgeAddr(br, gws.family, &gw)
				if err != nil {
					return fmt.Errorf("failed to set bridge addr: %v", err)
				}
			}

			if gws.gws != nil {
				if err = enableIPForward(gws.family); err != nil {
					return fmt.Errorf("failed to enable forwarding: %v", err)
				}
			}
		}
	}

	if len(n.VtepIPs) != 0 {
		// Create VxLAN tunnelings
		if err = br.AddVTEPs(n.VtepIPs); err != nil {
			return err
		}
	}

	if n.Controller != "" {
		// Set a SDN controller
		if err = br.SetCtrl(n.Controller); err != nil {
			return err
		}
	}

	if n.IPMasq {
		chain := utils.FormatChainName(n.Name, args.ContainerID)
		comment := utils.FormatComment(n.Name, args.ContainerID)
		for _, ipc := range result.IPs {
			if err = ip.SetupIPMasq(ip.Network(&ipc.Address), chain, comment); err != nil {
				return err
			}
		}
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
	var ipnets *net.IPNet
	err = ns.WithNetNSPath(args.Netns, func(_ ns.NetNS) error {
		var err error
		ipnets, err = ip.DelLinkByNameAddr(args.IfName, netlink.FAMILY_ALL)
		if err != nil && err == ip.ErrLinkNotFound {
			return nil
		}
		return err
	})

	if err != nil {
		return err
	}

	if n.IPMasq {
		chain := utils.FormatChainName(n.Name, args.ContainerID)
		comment := utils.FormatComment(n.Name, args.ContainerID)
		if err := ip.TeardownIPMasq(ipnets, chain, comment); err != nil {
			return err
		}
	}

	return err
}

func main() {
	skel.PluginMain(cmdAdd, cmdDel, version.All)
}
