// Copyright (c) 2017 Che Wei, Lin
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"encoding/binary"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"net"
	"os"
	"time"
)

type IPMConfig struct {
	Network   string
	SubnetLen int
	SubnetMin string
	SubnetMax string
	ETCDUrl   string
}

const etcdPrefix string = "/ovs-cni/networks/"

func powTwo(times int) uint32 {
	if times == 0 {
		return uint32(1)
	}

	var ans uint32
	ans = 1
	for i := 0; i < times; i++ {
		ans *= 2
	}

	return ans
}

func ip2int(ip net.IP) uint32 {
	if len(ip) == 16 {
		return binary.BigEndian.Uint32(ip[12:16])
	}
	return binary.BigEndian.Uint32(ip)
}

func int2ip(nn uint32) net.IP {
	ip := make(net.IP, 4)
	binary.BigEndian.PutUint32(ip, nn)
	return ip
}

func checkNodeRegister(nodeName string, cli clientv3.Client) (*net.IPNet, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	resp, err := cli.Get(ctx, etcdPrefix+nodeName, clientv3.WithPrefix())
	cancel()
	if err != nil {
		return nil, fmt.Errorf("Fetch etcd prefix error:%v", err)
	}

	if 0 == len(resp.Kvs) {
		return nil, nil
	}

	_, net, err := net.ParseCIDR(string(resp.Kvs[0].Value))
	return net, err
}

func getCurrentSubNets(cli clientv3.Client) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	resp, err := cli.Get(ctx, etcdPrefix, clientv3.WithPrefix())
	cancel()
	if err != nil {
		return nil, fmt.Errorf("Fetch etcd prefix error:%v", err)
	}

	subnets := []string{}
	for _, ev := range resp.Kvs {
		subnets = append(subnets, string(ev.Value))
	}
	return subnets, nil
}

func checkSubNetRegistered(subnet string, subsets []string) bool {
	for _, ev := range subsets {
		if ev == subnet {
			return true
		}
	}
	return false
}

func registerSubnet(nodeName string, ipmconfig IPMConfig, cli clientv3.Client) (*net.IPNet, error) {
	//Convert the subnet to int. for example.
	//string(10.16.7.0) -> net.IP(10.16.7.0) -> int(168822528)
	ipnet := net.ParseIP(ipmconfig.SubnetMin)
	ipStart := ip2int(ipnet)
	//Since the subnet len is 24, we need to add 2^(32-24) for each subnet.
	//(168822528 + 2^8) == 10.16.8.0
	//(168822528 + 2* 2 ^8 ) == 10.16.9.0
	ipNextSubnet := powTwo(32 - ipmconfig.SubnetLen)
	ipEnd := net.ParseIP(ipmconfig.SubnetMax)

	nextSubnet := int2ip(ipStart)

	subnets, err := getCurrentSubNets(cli)

	if err != nil {
		return nil, fmt.Errorf("Check Subnet Exist: %v", err)
	}

	for i := 1; ; i++ {
		cidr := fmt.Sprintf("%s/%d", nextSubnet.String(), ipmconfig.SubnetLen)
		exist := checkSubNetRegistered(cidr, subnets)
		//we can use this subnet if no one uses it
		if !exist {
			break
		}
		if ipEnd.String() == nextSubnet.String() {
			return nil, fmt.Errorf("No available subnet for registering")
		}
		nextSubnet = int2ip(ipStart + ipNextSubnet*uint32(i))
	}

	subnet := &net.IPNet{IP: nextSubnet, Mask: net.CIDRMask(ipmconfig.SubnetLen, 32)}
	_, err = cli.Put(context.TODO(), etcdPrefix+nodeName, subnet.String())
	return subnet, err
}

func GetSubnet(ipconfig IPMConfig) (*net.IPNet, error) {
	name, err := os.Hostname()
	if err != nil {
		return nil, fmt.Errorf("Failed to get NodeName: %v", err)
	}

	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{ipconfig.ETCDUrl},
		DialTimeout: 5 * time.Second,
	})

	subnet, err := checkNodeRegister(name, *cli)
	if err != nil {
		fmt.Println(err)
	}

	//Use subnet we register befored.
	if subnet != nil {
		return subnet, nil
	}

	//Register new subnet
	subnet, err = registerSubnet(name, ipconfig, *cli)
	return subnet, err

}

func main() {

	subnet, err := GetSubnet(IPMConfig{Network: "10.16.0.0", SubnetLen: 24, SubnetMin: "10.16.4.0", SubnetMax: "10.16.10.0", ETCDUrl: "10.240.0.26:2379"})
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(subnet)
	}
}
