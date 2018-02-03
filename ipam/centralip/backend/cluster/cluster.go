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

package cluster

import (
	"fmt"
	"github.com/John-Lin/ovs-cni/ipam/centralip/backend/utils"
	"github.com/coreos/etcd/clientv3"
	"math/rand"
	"net"
	"strings"
)

type NodeIPM struct {
	cli     *clientv3.Client
	podname string
	subnet  *net.IPNet
	config  *utils.IPMConfig
}

const clusterPrefix string = utils.ETCDPrefix + "cluster/"

func New(podName string, config *utils.IPMConfig) (*NodeIPM, error) {
	node := &NodeIPM{}
	node.config = config
	var err error

	node.podname = podName
	if strings.HasPrefix(config.ETCDURL, "https") {
		node.cli, err = utils.ConnectETCDWithTLS(config.ETCDURL, config.ETCDCertFile, config.ETCDKeyFile, config.ETCDTrustedCAFileFile)
	} else {
		node.cli, err = utils.ConnectETCD(config.ETCDURL)
	}
	if err != nil {
		return nil, err
	}

	_, node.subnet, err = net.ParseCIDR(config.Network)
	if err != nil {
		return nil, err
	}
	return node, nil
}

func (node *NodeIPM) GetGateway() (string, error) {
	return "", nil
}

func (node *NodeIPM) GetAvailableIP() (string, *net.IPNet, error) {
	ipnet := &net.IPNet{}
	if node.subnet == nil {
		return "", ipnet, fmt.Errorf("You should init IPM first")
	}

	usedIPPrefix := clusterPrefix + "used/"

	length, _ := node.subnet.Mask.Size()
	ipRange := utils.PowTwo(32 - (length))
	//Since the first IP is gateway, we should skip it

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
			utils.PutValue(node.cli, usedIPPrefix+tryIP.String(), node.podname)
			break
		}
	}

	var err error
	//We need to generate a net.IPnet object which contains the IP and Mask.
	//We use ParseCIDR to create the net.IPnet object and assign IP back to it.
	cidr := fmt.Sprintf("%s/%d", availableIP, length)
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
	usedIPPrefix := clusterPrefix + "used/"
	ipUsedToPod, err := utils.GetKeyValuesWithPrefix(node.cli,usedIPPrefix)
	if err != nil {
		return err
	}

	for k, v := range ipUsedToPod {
		fmt.Println(k, v)
		if v == node.podname {
			err := utils.DeleteKey(node.cli, k)
			return err
		}
	}
	return fmt.Errorf("There aren't any infomation about pod %s", node.podname)
}
