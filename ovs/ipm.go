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
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"os"
	"time"
)

type IPMConfig struct {
	Network   string
	SubnetLen int
	SubnetMin string
	SubnetMax string
}

const etcdPrefix string = "/ovs-cni/networks/"

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

func checkSubNetRegister(subnet String) error {

	return nil
}

func registerSubnet(nodeName string, ipmconfig IPMConfig, cli clientv3.Client) error {
	fmt.Println("Try to register")
	_, err := cli.Put(context.TODO(), etcdPrefix+nodeName, "QQQ")
	if err != nil {
		return fmt.Errorf("Put data into etcd fail: %v", err)
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

	fmt.Println(name)

	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"10.240.0.26:2379"},
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

	if err := GetSubnet(IPMConfig{Network: "10.16.0.0", SubnetLen: 24, SubnetMin: "10.16.4.0", SubnetMax: "10.16.2.0"}); err != nil {
		fmt.Println(err)
	}
}
