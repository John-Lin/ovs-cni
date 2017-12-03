// Copyright (c) 2017
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

package node

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/John-Lin/ovs-cni/ipam/centralip/backend/utils"
	"github.com/containernetworking/plugins/pkg/ip"
	"github.com/coreos/etcd/clientv3"
	"net"
	"time"
)

type NodeIPM struct {
	cli      *clientv3.Client
	hostname string
	podname  string
	subnet   *net.IPNet
	IPMConfig
}

const subnetPrefix string = etcdPrefix + "subnets/"

func New(podName string, config IPMConfig) (*NodeIPM, error) {
	node := &NodeIPM{}
	var err error

	node.hostname, err = os.Hostname()
	if err != nil {
		return nil, err
	}

	node.podname = podName
	err = node.connect(config.ETCDURL)
	if err != nil {
		return nil, err
	}

	err = node.registerNode()
	if err != nil {
		return nil, err
	}
	return node, nil
}

/*
	ETCD Related
*/
func (node *NodeIPM) connect(etcdUrl string) error {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{etcdUrl},
		DialTimeout: 5 * time.Second,
	})

	node.cli = cli
	return err
}

func (node *NodeIPM) deleteKey(prefix string) error {
	_, err := node.cli.Delete(context.TODO(), prefix)
	return err
}
func (node *NodeIPM) putValue(prefix, value string) error {
	_, err := node.cli.Put(context.TODO(), prefix, value)
	return err
}

func (node *NodeIPM) getKeyValuesWithPrefix(key string) (map[string]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	resp, err := node.cli.Get(ctx, key, clientv3.WithPrefix())
	cancel()
	if err != nil {
		return nil, fmt.Errorf("Fetch etcd prefix error:%v", err)
	}

	results := make(map[string]string)
	for _, ev := range resp.Kvs {
		results[string(ev.Key)] = string(ev.Value)
	}

	return results, nil
}

func (node *NodeIPM) checkNodeIsRegisted() error {

	keyValues, err := node.getKeyValuesWithPrefix(etcdPrefix + node.hostname)
	if err != nil {
		return err
	}

	if 0 == len(keyValues) {
		return nil
	}

	_, node.subnet, err = net.ParseCIDR(keyValues[etcdPrefix+node.hostname])
	return err
}

func (node *NodeIPM) registerSubnet() error {
	//Convert the subnet to int. for example.
	//string(10.16.7.0) -> net.IP(10.16.7.0) -> int(168822528)
	ipnet := net.ParseIP(node.SubnetMin)
	ipStart, err := ipToInt(ipnet)
	if err != nil {
		return err
	}

	//Since the subnet len is 24, we need to add 2^(32-24) for each subnet.
	//(168822528 + 2^8) == 10.16.8.0
	//(168822528 + 2* 2 ^8 ) == 10.16.9.0
	ipNextSubnet := powTwo(32 - node.SubnetLen)
	ipEnd := net.ParseIP(node.SubnetMax)

	nextSubnet := intToIP(ipStart)

	nodeToSubnets, err := node.getKeyValuesWithPrefix(subnetPrefix)

	if err != nil {
		return err
	}

	for i := 1; ; i++ {
		cidr := fmt.Sprintf("%s%s/%d", subnetPrefix, nextSubnet.String(), node.SubnetLen)

		if _, ok := nodeToSubnets[cidr]; !ok {
			break
		}
		if ipEnd.String() == nextSubnet.String() {
			return fmt.Errorf("No available subnet for registering")
		}
		nextSubnet = intToIP(ipStart + ipNextSubnet*uint32(i))
	}

	subnet := &net.IPNet{IP: nextSubnet, Mask: net.CIDRMask(node.SubnetLen, 32)}
	node.subnet = subnet

	//store the $etcdPrefix/hostname -> subnet
	err = node.putValue(etcdPrefix+node.hostname, subnet.String())
	if err != nil {
		return err
	}

	//store the $etcdPrefix/subnets/$subnet -> hostname  for fast lookup for existing subnet
	err = node.putValue(subnetPrefix+subnet.String(), node.hostname)
	return err
}

func (node *NodeIPM) registerNode() error {
	//Check Node Exist
	err := node.checkNodeIsRegisted()
	if err != nil {
		return err
	}

	if node.subnet == nil {
		err := node.registerSubnet()
		if err != nil {
			return err
		}
	}
	return nil
}

func (node *NodeIPM) Init(hostname, podname string) error {
	node.hostname = hostname
	node.podname = podname

	err := node.connect(node.ETCDURL)
	if err != nil {
		return err
	}

	err = node.registerNode()
	if err != nil {
		return err
	}
	return nil
}

func (node *NodeIPM) GetGateway() (string, error) {
	if node.subnet == nil {
		return "", fmt.Errorf("You should init IPM first")
	}

	gwPrefix := etcdPrefix + node.hostname + "/gateway"
	nodeValues, err := node.getKeyValuesWithPrefix(gwPrefix)
	if err != nil {
		return "", err
	}

	var gwIP string
	if len(nodeValues) == 0 {
		gwIP = getNextIP(node.subnet).String()
		node.putValue(gwPrefix, gwIP)
	} else {
		gwIP = nodeValues[gwPrefix]
	}
	return gwIP, nil
}

func (node *NodeIPM) GetAvailableIP() (string, *net.IPNet, error) {
	ipnet := &net.IPNet{}
	if node.subnet == nil {
		return "", ipnet, fmt.Errorf("You should init IPM first")
	}

	usedIPPrefix := etcdPrefix + node.hostname + "/used/"
	ipUsedToPod, err := node.getKeyValuesWithPrefix(usedIPPrefix)
	if err != nil {
		return "", ipnet, err
	}

	ipRange := powTwo(32 - (node.SubnetLen))
	//Since the first IP is gateway, we should skip it
	tmpIP := ip.NextIP(getNextIP(node.subnet))

	var availableIP string
	for i := 1; i < int(ipRange); i++ {
		//check.
		if _, ok := ipUsedToPod[usedIPPrefix+tmpIP.String()]; !ok {
			availableIP = tmpIP.String()
			node.putValue(usedIPPrefix+tmpIP.String(), node.podname)
			break
		}
		tmpIP = ip.NextIP(tmpIP)
	}

	//We need to generate a net.IPnet object which contains the IP and Mask.
	//We use ParseCIDR to create the net.IPnet object and assign IP back to it.
	cidr := fmt.Sprintf("%s/%d", availableIP, node.SubnetLen)
	var ip net.IP
	ip, ipnet, err = net.ParseCIDR(cidr)
	if err != nil {
		return "", ipnet, err
	}

	ipnet.IP = ip
	return availableIP, ipnet, nil
}

func (node *NodeIPM) Delete() error {
	//get all used ip address and try to matches it id.
	usedIPPrefix := etcdPrefix + node.hostname + "/used/"
	ipUsedToPod, err := node.getKeyValuesWithPrefix(usedIPPrefix)
	if err != nil {
		return err
	}

	for k, v := range ipUsedToPod {
		if v == node.podname {
			err := node.deleteKey(k)
			return err
		}
	}
	return fmt.Errorf("There aren't any infomation about %s", node.hostname)
}
