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

func checkNodeRegister(nodeName string, cli clientv3.Client) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	resp, err := cli.Get(ctx, etcdPrefix+nodeName, clientv3.WithPrefix())
	cancel()
	if err != nil {
		return false, fmt.Errorf("Fetch etcd prefix error:%v", err)
	}

	if 0 == len(resp.Kvs) {
		return false, nil
	}
	/*
		for _, ev := range resp.Kvs {
			fmt.Printf("%s : %s\n", ev.Key, ev.Value)
		}
	*/
	return true, nil
}

func checkSubNetRegistered(subnet string, cli clientv3.Client) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	resp, err := cli.Get(ctx, etcdPrefix, clientv3.WithPrefix())
	cancel()
	if err != nil {
		return false, fmt.Errorf("Fetch etcd prefix error:%v", err)
	}

	for _, ev := range resp.Kvs {
		if string(ev.Value) == subnet {
			return true, nil
		}
	}
	return false, nil
}

func registerSubnet(nodeName string, ipmconfig IPMConfig, cli clientv3.Client) error {
	fmt.Println("Try to register")
	_, err := cli.Put(context.TODO(), etcdPrefix+nodeName, "QQQ")
	if err != nil {
		return fmt.Errorf("Put data into etcd fail: %v", err)
	}

	ipnet := net.ParseIP(ipmconfig.SubnetMin)
	ipStart := ip2int(ipnet)
	ipNextSubnet := powTwo(32 - ipmconfig.SubnetLen)
	ipEnd := net.ParseIP(ipmconfig.SubnetMax)

	fmt.Println(ipEnd)
	for i := 0; ; i++ {
		nextSubnet := int2ip(ipStart + ipNextSubnet*uint32(i))
		success, err := checkSubNetRegistered(nextSubnet.String(), cli)
		if err != nil {
			return fmt.Errorf("Check Subnet Exist: %v", err)
		}
		if success {
			return nil
		}
		if ipEnd.String() == nextSubnet.String() {
			return nil
		}
	}
	return nil
}

func GetSubnet(ipconfig IPMConfig) error {
	fmt.Println(ipconfig.Network)
	fmt.Println(ipconfig.SubnetLen)
	name, err := os.Hostname()
	if err != nil {
		return fmt.Errorf("Failed to get NodeName: %v", err)
	}

	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{ipconfig.ETCDUrl},
		DialTimeout: 5 * time.Second,
	})

	exist, err := checkNodeRegister(name, *cli)
	if err != nil {
		fmt.Println(err)
	}

	if !exist {
		registerSubnet(name, ipconfig, *cli)
	}

	return nil
}

func main() {

	if err := GetSubnet(IPMConfig{Network: "10.16.0.0", SubnetLen: 24, SubnetMin: "10.16.4.0", SubnetMax: "10.16.10.0", ETCDUrl: "10.240.0.26:2379"}); err != nil {
		fmt.Println(err)
	}
}
