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
	"github.com/stretchr/testify/assert"
	"github.com/vishvananda/netlink"
	"io/ioutil"
	"math/rand"
	"net"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestVxlanIfName(t *testing.T) {
	// Test to returns formatted vxlan interface name
	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)

	var reg []string
	for i := 0; i <= 3; i++ {
		reg = append(reg, strconv.Itoa(r1.Intn(256)))
	}

	intfName := vxlanIfName(strings.Join(reg[:], "."))

	checked := "vxif" + strings.Join(reg[:], "_")
	assert.Equal(t, intfName, checked, "Those two names should be the same")
}

func TestSetLinkUp(t *testing.T) {
	err := setLinkUp("lo")
	assert.NoError(t, err)
}

func TestSetLinkUp_Invalid(t *testing.T) {
	err := setLinkUp("unknown")
	assert.Error(t, err)
}

func TestIPForward(t *testing.T) {
	v4Path := "/proc/sys/net/ipv4/ip_forward"
	v6Path := "/proc/sys/net/ipv6/conf/all/forwarding"

	if _, err := os.Stat(v4Path); !os.IsNotExist(err) {
		enableIPForward(netlink.FAMILY_V4)
		content, _ := ioutil.ReadFile(v4Path)
		assert.Equal(t, "1\n", string(content))
	}

	if _, err := os.Stat(v6Path); !os.IsNotExist(err) {
		enableIPForward(netlink.FAMILY_V6)
		content, _ := ioutil.ReadFile(v6Path)
		assert.Equal(t, "1\n", string(content))
	}
}

func TestGetGatewayFromIP(t *testing.T) {
	_, input, _ := net.ParseCIDR("192.168.194.0/22")

	gwIP := getNextIP(input)
	assert.Equal(t, gwIP.String(), "192.168.192.1")
}
