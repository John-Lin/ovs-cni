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
	"fmt"
	"github.com/John-Lin/ovs-cni/ipam/centralip/backend/utils"
	"github.com/coreos/etcd/clientv3"
	"math/rand"
	"net"
	"strings"
)

type NodeIPM struct {
	cli      *clientv3.Client
	hostname string
	podname  string
	subnet   *net.IPNet
	config   *utils.IPMConfig
}

const nodePrefix string = utils.ETCDPrefix + "node/"
const subnetPrefix string = nodePrefix + "subnets/"

func New(podName, hostname string, config *utils.IPMConfig) (*NodeIPM, error) {
	node := &NodeIPM{}
	node.config = config
	var err error

	node.hostname = hostname
	node.podname = podName

	if strings.HasPrefix(config.ETCDURL, "https") {
		node.cli, err = utils.ConnectETCDWithTLS(config.ETCDURL, config.ETCDCertFile, config.ETCDKeyFile, config.ETCDTrustedCAFileFile)
	} else {
		node.cli, err = utils.ConnectETCD(config.ETCDURL)
	}

	if err != nil {
		return nil, err
	}

	err = node.registerNode()
	if err != nil {
		return nil, err
	}
	return node, nil
}

func (node *NodeIPM) checkNodeIsRegisted() error {

	keyValues, err := utils.GetKeyValuesWithPrefix(node.cli,nodePrefix + node.hostname)
	if err != nil {
		return err
	}

	if 0 == len(keyValues) {
		return nil
	}

	_, node.subnet, err = net.ParseCIDR(keyValues[nodePrefix+node.hostname])
	return err
}

func (node *NodeIPM) registerSubnet() error {
	//Convert the subnet to int. for example.
	//string(10.16.7.0) -> net.IP(10.16.7.0) -> int(168822528)
	ipnet := net.ParseIP(node.config.SubnetMin)
	ipStart, err := utils.IpToInt(ipnet)
	if err != nil {
		fmt.Println(node)
		return err
	}

	//Since the subnet len is 24, we need to add 2^(32-24) for each subnet.
	//(168822528 + 2^8) == 10.16.8.0
	//(168822528 + 2* 2 ^8 ) == 10.16.9.0
	ipNextSubnet := utils.PowTwo(32 - node.config.SubnetLen)
	ipEnd := net.ParseIP(node.config.SubnetMax)

	nextSubnet := utils.IntToIP(ipStart)

	nodeToSubnets, err :=utils.GetKeyValuesWithPrefix(node.cli,subnetPrefix)

	if err != nil {
		return err
	}

	for i := 1; ; i++ {
		cidr := fmt.Sprintf("%s%s/%d", subnetPrefix, nextSubnet.String(), node.config.SubnetLen)

		if _, ok := nodeToSubnets[cidr]; !ok {
			break
		}
		if ipEnd.String() == nextSubnet.String() {
			return fmt.Errorf("No available subnet for registering")
		}
		nextSubnet = utils.IntToIP(ipStart + ipNextSubnet*uint32(i))
	}

	subnet := &net.IPNet{IP: nextSubnet, Mask: net.CIDRMask(node.config.SubnetLen, 32)}
	node.subnet = subnet

	//store the $nodePrefix/hostname -> subnet
	err = utils.PutValue(node.cli,nodePrefix+node.hostname, subnet.String())
	if err != nil {
		return err
	}

	//store the $nodePrefix/subnets/$subnet -> hostname  for fast lookup for existing subnet
	err = utils.PutValue(node.cli,subnetPrefix+subnet.String(), node.hostname)
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

func (node *NodeIPM) GetGateway() (string, error) {
	if node.subnet == nil {
		return "", fmt.Errorf("You should init IPM first")
	}

	gwPrefix := nodePrefix + node.hostname + "/gateway"
	nodeValues, err := utils.GetKeyValuesWithPrefix(node.cli,gwPrefix)
	if err != nil {
		return "", err
	}

	var gwIP string
	if len(nodeValues) == 0 {
		gwIP = utils.GetNextIP(node.subnet).String()
		utils.PutValue(node.cli,gwPrefix, gwIP)
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

	usedIPPrefix := nodePrefix + node.hostname + "/used/"
	ipRange := utils.PowTwo(32 - (node.config.SubnetLen))

	//change to random a ip (must not be gateway) and try to check the etcd:
	start := utils.GetNextIP(node.subnet)

	//If ip Range = 256, our target it 2~255
	//rand.Intn( range - 2 ) will return 0<=n<254,
	//+2 will cause 2<=n<256
	retryTimes := 20
	var availableIP string
	for i := 0; i < retryTimes; i++ {
		tryIP := utils.GetIPByInt(start, uint32(rand.Intn(int(ipRange-2))+1))

		ipUsedToPod, err := utils.GetKeyValuesWithPrefix(node.cli,usedIPPrefix)
		if err != nil {
			return "", ipnet, err
		}

		//check.
		if _, ok := ipUsedToPod[usedIPPrefix+tryIP.String()]; !ok {
			availableIP = tryIP.String()
			utils.PutValue(node.cli,usedIPPrefix+tryIP.String(), node.podname)
			break
		}
	}

	var err error
	//We need to generate a net.IPnet object which contains the IP and Mask.
	//We use ParseCIDR to create the net.IPnet object and assign IP back to it.
	cidr := fmt.Sprintf("%s/%d", availableIP, node.config.SubnetLen)
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
	usedIPPrefix := nodePrefix + node.hostname + "/used/"
	ipUsedToPod, err := utils.GetKeyValuesWithPrefix(node.cli,usedIPPrefix)
	if err != nil {
		return err
	}

	for k, v := range ipUsedToPod {
		if v == node.podname {
			err := utils.DeleteKey(node.cli, k)
			return err
		}
	}
	return fmt.Errorf("There aren't any infomation about %s", node.hostname)
}
