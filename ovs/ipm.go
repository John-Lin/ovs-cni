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
	"github.com/containernetworking/plugins/pkg/ip"
	"net"
	"time"
)


type CentralIPM struct {
	cli	*clientv3.Client
	hostname string
	podname string
	subnet	*net.IPNet
	IPMConfig
}

type IPMConfig struct {
	Network   string `json:"network"`
	SubnetLen int    `json:"subnetLen"`
	SubnetMin string `json:"subnetMin"`
	SubnetMax string `json:"subnetMax"`
	ETCDURL   string `json:"etcdURL"`
}

const etcdPrefix string = "/ovs-cni/networks/"

func generateCentralIPM(bytes[]byte) (*CentralIPM, error) {
	n := &CentralIPM{}
	if err := json.Unmarshal(bytes, n); err != nil {
		return nil, fmt.Errorf("failed to load netconf: %v", err)
	}
	return n, nil
}


/*
	ETCD Related
*/
func (ipm *CentralIPM) connect(etcdUrl string) error {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{etcdUrl},
		DialTimeout: 5 * time.Second,
	})

	ipm.cli = cli
	return err
}

func (ipm *CentralIPM) putValue(prefix, value string) (error) {
	_, err := ipm.cli.Put(context.TODO(), prefix, value)
	return err
}

func (ipm *CentralIPM) getKeyValuesWithPrefix(key string) (map[string]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	resp, err := ipm.cli.Get(ctx, key, clientv3.WithPrefix())
	cancel()
	if err != nil {
		return nil, fmt.Errorf("Fetch etcd prefix error:%v", err)
	}

	results :=  make(map[string]string)
	for _, ev := range resp.Kvs {
		results[string(ev.Key)] = string(ev.Value)
	}

	return results, nil
}

func (ipm *CentralIPM) checkNodeIsRegisted() error {

	keyValues, err := ipm.getKeyValuesWithPrefix(etcdPrefix+ipm.hostname)
	if err != nil {
		return err
	}

	if 0 == len(keyValues) {
		return nil
	}

	_, ipm.subnet, err = net.ParseCIDR(keyValues[etcdPrefix+ipm.hostname])
	return err
}

func (ipm *CentralIPM) registerSubnet() error {
	//Convert the subnet to int. for example.
	//string(10.16.7.0) -> net.IP(10.16.7.0) -> int(168822528)
	ipnet := net.ParseIP(ipm.SubnetMin)
	ipStart, err := ipToInt(ipnet)
	if err != nil {
		return err
	}

	//Since the subnet len is 24, we need to add 2^(32-24) for each subnet.
	//(168822528 + 2^8) == 10.16.8.0
	//(168822528 + 2* 2 ^8 ) == 10.16.9.0
	ipNextSubnet := powTwo(32 - ipm.SubnetLen)
	ipEnd := net.ParseIP(ipm.SubnetMax)

	nextSubnet := intToIP(ipStart)

	nodeToSubnets, err := ipm.getKeyValuesWithPrefix(etcdPrefix+"subnets/")

	if err != nil {
		return fmt.Errorf("Check Subnet Exist: %v", err)
	}

	for i := 1; ; i++ {
		cidr := fmt.Sprintf("%s/%d", nextSubnet.String(), ipm.SubnetLen)

		if _, ok := nodeToSubnets[cidr]; !ok {
			break
		}
		if ipEnd.String() == nextSubnet.String() {
			return fmt.Errorf("No available subnet for registering")
		}
		nextSubnet = intToIP(ipStart + ipNextSubnet*uint32(i))
	}

	subnet := &net.IPNet{IP: nextSubnet, Mask: net.CIDRMask(ipm.SubnetLen, 32)}
	ipm.subnet = subnet

	//store the $etcdPrefix/hostname -> subnet
	err = ipm.putValue(etcdPrefix+ipm.hostname, subnet.String())
	if err != nil {
		return err
	}

	//store the $etcdPrefix/subnets/$subnet -> hostname  for fast lookup for existing subnet
	err = ipm.putValue(etcdPrefix + "subnets/" +subnet.String(), ipm.hostname)
	return err
}

func (ipm *CentralIPM) registerNode() error {
	//Check Node Exist
	err := ipm.checkNodeIsRegisted()
	if err != nil {
		return err
	}

	if ipm.subnet == nil {
		err := ipm.registerSubnet()
		if err != nil {
			return err
		}
	}
	return nil
}

func (ipm *CentralIPM) Init(hostname, podname string)  error {
	ipm.hostname = hostname
	ipm.podname = podname

	err := ipm.connect(ipm.ETCDURL)
	if err != nil {
		return err
	}

	err = ipm.registerNode()
	if err != nil {
		return err
	}
	return nil
}

func (ipm *CentralIPM) GetGateway() (string,error) {
	if ipm.subnet == nil {
		return "", fmt.Errorf("You should init IPM first")
	}
	
	gwPrefix := etcdPrefix + ipm.hostname + "/gateway"
	nodeValues, err := ipm.getKeyValuesWithPrefix(gwPrefix)
	if err != nil { 
		return "",err
	}

	var gwIP string
	if len(nodeValues) == 0 {
		gwIP = getNextIP(ipm.subnet).String()
		ipm.putValue(gwPrefix, gwIP)
	} else {
		gwIP = nodeValues[gwPrefix]
	}
	return gwIP, nil
}

func (ipm *CentralIPM) GetAvailableIP() (string,error) {
	if ipm.subnet == nil {
		return "", fmt.Errorf("You should init IPM first")
	}

	ipPrefix := etcdPrefix + ipm.hostname + "/"
	ipUsedToPod, err := ipm.getKeyValuesWithPrefix(ipPrefix)
	if err != nil { 
		return "",err
	}

	ipRange := powTwo(32-(ipm.SubnetLen))
	//Since the first IP is gateway, we should skip it
	tmpIP:= ip.NextIP(getNextIP(ipm.subnet))

	usedIPPrefix := ipPrefix + "used/"
	var availableIP string
	for i:=1;i<int(ipRange);i++ {
		//check.
		if _, ok := ipUsedToPod[usedIPPrefix + tmpIP.String()]; !ok {
			availableIP = tmpIP.String()
			ipm.putValue(usedIPPrefix+ tmpIP.String(), ipm.podname)
			break
		}
		tmpIP = ip.NextIP(tmpIP) 
	}

	return availableIP, nil
}
