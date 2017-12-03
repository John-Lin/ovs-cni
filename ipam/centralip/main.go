package main

import (
	"net"
	"os"
	//	"strings"
	"github.com/John-Lin/ovs-cni/ipam/centralip/backend"
	"github.com/containernetworking/cni/pkg/skel"
	"github.com/containernetworking/cni/pkg/types"
	"github.com/containernetworking/cni/pkg/types/current"
	"github.com/containernetworking/cni/pkg/version"
)

func main() {
	skel.PluginMain(cmdAdd, cmdDel, version.All)
}

/*

type CmdArgs struct {
	ContainerID string
	Netns       string
	IfName      string
	Args        string
	Path        string
	StdinData   []byte
}

*/
func cmdAdd(args *skel.CmdArgs) error {
	n, err, cniversion := centralip.GenerateCentralIPM(args)
	if err != nil {
		return err
	}

	gwIP, err := n.GetGateway()
	_, IP, err := n.GetAvailableIP()
	if err != nil {
		return err
	}

	i := net.ParseIP(gwIP)

	version := "4"
	if IP.IP.To4() == nil {
		version = "6"
	}
	ipconfig := &current.IPConfig{
		Version: version,
		Address: *IP,
		Gateway: i,
	}

	result := &current.Result{}
	result.IPs = append(result.IPs, ipconfig)
	result.Routes = []*types.Route{}
	return types.PrintResult(result, cniversion)
}

func cmdDel(args *skel.CmdArgs) error {
	n, _, err := centralip.GenerateCentralIPM(args)
	if err != nil {
		return err
	}

	err = n.Delete()
	if err != nil {
		return err
	}

	return nil
}
