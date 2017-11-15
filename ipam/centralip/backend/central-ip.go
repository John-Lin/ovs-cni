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

package centralip

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/containernetworking/plugins/pkg/ip"
	"github.com/coreos/etcd/clientv3"
	"net"
	"time"
)

type CentralNet struct {
	Name       string      `json:"name"`
	CNIVersion string      `json:"cniVersion"`
	IPM        *CentralIPM `json:"ipam"`
}

type CentralIPM struct {
	cli      *clientv3.Client
	hostname string
	podname  string
	subnet   *net.IPNet
	IPMConfig
}

type IPMConfig struct {
	Type      string `json:"type"`
	Network   string `json:"network"`
	SubnetLen int    `json:"subnetLen"`
	SubnetMin string `json:"subnetMin"`
	SubnetMax string `json:"subnetMax"`
	ETCDURL   string `json:"etcdURL"`
}

const etcdPrefix string = "/ovs-cni/networks/"
const subnetPrefix string = etcdPrefix + "subnets/"

func GenerateCentralIPM(bytes []byte) (*CentralIPM, string, error) {
	n := &CentralNet{}
	if err := json.Unmarshal(bytes, n); err != nil {
		return nil, "", fmt.Errorf("failed to load netconf: %v", err)
	}
	return n.IPM, n.CNIVersion, nil
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

func (ipm *CentralIPM) deleteKey(prefix string) error {
	_, err := ipm.cli.Delete(context.TODO(), prefix)
	return err
}
func (ipm *CentralIPM) putValue(prefix, value string) error {
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

	results := make(map[string]string)
	for _, ev := range resp.Kvs {
		results[string(ev.Key)] = string(ev.Value)
	}

	return results, nil
}

func (ipm *CentralIPM) checkNodeIsRegisted() error {

	keyValues, err := ipm.getKeyValuesWithPrefix(etcdPrefix + ipm.hostname)
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

	nodeToSubnets, err := ipm.getKeyValuesWithPrefix(subnetPrefix)

	if err != nil {
		return err
	}

	for i := 1; ; i++ {
		cidr := fmt.Sprintf("%s%s/%d", subnetPrefix, nextSubnet.String(), ipm.SubnetLen)

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
	err = ipm.putValue(subnetPrefix+subnet.String(), ipm.hostname)
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

func (ipm *CentralIPM) Init(hostname, podname string) error {
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

func (ipm *CentralIPM) GetGateway() (string, error) {
	if ipm.subnet == nil {
		return "", fmt.Errorf("You should init IPM first")
	}

	gwPrefix := etcdPrefix + ipm.hostname + "/gateway"
	nodeValues, err := ipm.getKeyValuesWithPrefix(gwPrefix)
	if err != nil {
		return "", err
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

func (ipm *CentralIPM) GetAvailableIP() (string, *net.IPNet, error) {
	ipnet := &net.IPNet{}
	if ipm.subnet == nil {
		return "", ipnet, fmt.Errorf("You should init IPM first")
	}

	usedIPPrefix := etcdPrefix + ipm.hostname + "/used/"
	ipUsedToPod, err := ipm.getKeyValuesWithPrefix(usedIPPrefix)
	if err != nil {
		return "", ipnet, err
	}

	ipRange := powTwo(32 - (ipm.SubnetLen))
	//Since the first IP is gateway, we should skip it
	tmpIP := ip.NextIP(getNextIP(ipm.subnet))

	var availableIP string
	for i := 1; i < int(ipRange); i++ {
		//check.
		if _, ok := ipUsedToPod[usedIPPrefix+tmpIP.String()]; !ok {
			availableIP = tmpIP.String()
			ipm.putValue(usedIPPrefix+tmpIP.String(), ipm.podname)
			break
		}
		tmpIP = ip.NextIP(tmpIP)
	}

	//We need to generate a net.IPnet object which contains the IP and Mask.
	//We use ParseCIDR to create the net.IPnet object and assign IP back to it.
	cidr := fmt.Sprintf("%s/%d", availableIP, ipm.SubnetLen)
	var ip net.IP
	ip, ipnet, err = net.ParseCIDR(cidr)
	if err != nil {
		return "", ipnet, err
	}

	ipnet.IP = ip
	return availableIP, ipnet, nil
}

func (ipm *CentralIPM) DeleteIPByName(name string) error {
	//get all used ip address and try to matches it id.
	//ipm.deleteKey(podName)a
	usedIPPrefix := etcdPrefix + ipm.hostname + "/used/"
	ipUsedToPod, err := ipm.getKeyValuesWithPrefix(usedIPPrefix)
	if err != nil {
		return err
	}

	for k, v := range ipUsedToPod {
		if v == ipm.podname {
			err := ipm.deleteKey(k)
			return err
		}
	}
	return fmt.Errorf("There aren't any infomation about %s", ipm.hostname)
}
