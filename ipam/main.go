package main

import (
	"net"
	"os"
	//	"strings"
	"fmt"
	"github.com/John-Lin/ovs-cni/ipam/centralip"
	"github.com/containernetworking/cni/pkg/skel"
	"github.com/containernetworking/cni/pkg/types"
	"github.com/containernetworking/cni/pkg/types/current"
	"github.com/containernetworking/cni/pkg/version"
)

func main() {
	skel.PluginMain(cmdAdd, cmdDel, version.All)
}

func cmdAdd(args *skel.CmdArgs) error {
	n, cniversion, err := centralip.GenerateCentralIPM(args.StdinData)
	if err != nil {
		return err
	}

	hostname, err := os.Hostname()
	if err != nil {
		return err
	}

	podName := args.ContainerID
	err = n.Init(hostname, podName)

	gwIP, err := n.GetGateway()
	_, IP, err := n.GetAvailableIP()
	if err != nil {
		fmt.Println(err)
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
	return nil
}
