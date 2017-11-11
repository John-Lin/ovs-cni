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

package centralip

import (
	"encoding/binary"
	"fmt"
	"github.com/containernetworking/plugins/pkg/ip"
	"net"
)

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

func ipToInt(ip net.IP) (uint32, error) {
	if v4 := ip.To4(); v4 != nil {
		return binary.BigEndian.Uint32(ip[12:16]), nil
	}
	return 0, fmt.Errorf("IP should be ipv4\n")
}

func intToIP(nn uint32) net.IP {
	ip := make(net.IP, 4)
	binary.BigEndian.PutUint32(ip, nn)
	return ip
}

//We use the first IP as gateway address
func getNextIP(ipn *net.IPNet) net.IP {
	nid := ipn.IP.Mask(ipn.Mask)
	return ip.NextIP(nid)
}
