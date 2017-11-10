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
	"encoding/json"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"net"
	"os"
	"time"
)

type CentralNet struct {
	IPM *IPMConfig `json:"ipam"`
}

type IPMConfig struct {
	Network   string `json:"network"`
	SubnetLen int    `json:"subnetLen"`
	SubnetMin string `json:"subnetMin"`
	SubnetMax string `json:"subnetMax"`
	ETCDURL   string `json:"etcdURL"`
}

const etcdPrefix string = "/ovs-cni/networks/"

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
	ipStart, err := ipToInt(ipnet)
	if err != nil {
		return nil, err
	}

	//Since the subnet len is 24, we need to add 2^(32-24) for each subnet.
	//(168822528 + 2^8) == 10.16.8.0
	//(168822528 + 2* 2 ^8 ) == 10.16.9.0
	ipNextSubnet := powTwo(32 - ipmconfig.SubnetLen)
	ipEnd := net.ParseIP(ipmconfig.SubnetMax)

	nextSubnet := intToIP(ipStart)

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
		nextSubnet = intToIP(ipStart + ipNextSubnet*uint32(i))
	}

	subnet := &net.IPNet{IP: nextSubnet, Mask: net.CIDRMask(ipmconfig.SubnetLen, 32)}
	_, err = cli.Put(context.TODO(), etcdPrefix+nodeName, subnet.String())
	return subnet, err
}

func GetSubnet(IPMConfig IPMConfig, name string) (*net.IPNet, error) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{IPMConfig.ETCDURL},
		DialTimeout: 5 * time.Second,
	})

	if err != nil {
		return nil, err
	}

	subnet, err := checkNodeRegister(name, *cli)
	if err != nil {
		return subnet, nil
	}

	//Use subnet we register befored.
	if subnet != nil {
		return subnet, nil
	}

	//Register new subnet
	subnet, err = registerSubnet(name, IPMConfig, *cli)
	return subnet, err
}

func GenerateHostLocalConfig(input []byte) []byte {
	n := CentralNet{}
	if err := json.Unmarshal(input, &n); err != nil {
		return []byte{}
	}

	name, err := os.Hostname()
	if err != nil {
		return []byte{}
	}

	subnet, err := GetSubnet(*n.IPM, name)
	if err != nil {
		return []byte{}
	}
	//Generate data to localHost
	newConfig := string(`
{
"ipam":{
"type":"host-local",
"subnet":"` + subnet.String() + `"
}
}
`)
	return []byte(newConfig)
}
