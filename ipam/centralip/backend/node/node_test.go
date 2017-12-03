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
	"github.com/John-Lin/ovs-cni/ipam/centralip/backend/utils"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
	"time"
)

var node *NodeIPM
var err error

var validData = utils.IPMConfig{
	Network:   "10.123.0.0/16",
	SubnetLen: 24,
	SubnetMin: "10.123.5.0",
	SubnetMax: "10.123.6.0",
	ETCDURL:   "127.0.0.1:2379",
}

func TestNewNode(t *testing.T) {
	if _, defined := os.LookupEnv("TEST_ETCD"); !defined {
		t.SkipNow()
		return
	}
	node, err = New("pod1", "host1", &validData)
	assert.NoError(t, err)
	assert.NotNil(t, node)
	assert.Equal(t, node.config.ETCDURL, "127.0.0.1:2379")
}

func TestGetGateway(t *testing.T) {
	if _, defined := os.LookupEnv("TEST_ETCD"); !defined {
		t.SkipNow()
		return
	}
	gwIP, err := node.GetGateway()
	assert.NoError(t, err)
	assert.Equal(t, "10.123.5.1", gwIP)
	gwIP, err = node.GetGateway()
	assert.NoError(t, err)
	assert.Equal(t, "10.123.5.1", gwIP)
}

func TestGetAvailableIP(t *testing.T) {
	if _, defined := os.LookupEnv("TEST_ETCD"); !defined {
		t.SkipNow()
		return
	}
	t.Run("First IP", func(t *testing.T) {
		ip, ipNet, err := node.GetAvailableIP()
		assert.NoError(t, err)
		assert.Equal(t, "10.123.5.2/24", ipNet.String())
		assert.Equal(t, "10.123.5.2", ip)
	})
	time.Sleep(1 * time.Second)
	t.Run("Second IP", func(t *testing.T) {
		ip, ipNet, err := node.GetAvailableIP()
		assert.NoError(t, err)
		assert.Equal(t, "10.123.5.3/24", ipNet.String())
		assert.Equal(t, "10.123.5.3", ip)
	})
	time.Sleep(1 * time.Second)
	t.Run("remove first IP", func(t *testing.T) {
		err := node.Delete()
		assert.NoError(t, err)
	})
	time.Sleep(1 * time.Second)
	t.Run("Fetch IP again", func(t *testing.T) {
		ip, ipNet, err := node.GetAvailableIP()
		assert.NoError(t, err)
		assert.Equal(t, "10.123.5.2/24", ipNet.String())
		assert.Equal(t, "10.123.5.2", ip)
	})
}

func TestSecondHost(t *testing.T) {
	if _, defined := os.LookupEnv("TEST_ETCD"); !defined {
		t.SkipNow()
		return
	}
	node2, err := New("pod1", "host2", &validData)
	assert.NoError(t, err)

	gwIP, err := node2.GetGateway()
	assert.NoError(t, err)
	assert.Equal(t, "10.123.6.1", gwIP)
	ip, ipNet, err := node2.GetAvailableIP()
	assert.NoError(t, err)
	assert.Equal(t, "10.123.6.2/24", ipNet.String())
	assert.Equal(t, "10.123.6.2", ip)

}

func TestGenerateCentralIPMInvalid(t *testing.T) {
	if _, defined := os.LookupEnv("TEST_ETCD"); !defined {
		t.SkipNow()
		return
	}
	var InvalidData = utils.IPMConfig{
		Network:   "10.123.0.0/16",
		SubnetLen: 24,
		SubnetMin: "10.123.5.0",
		SubnetMax: "10.123.6.0",
		ETCDURL:   "127.0.0.1:23792",
	}

	t.Run("invalid etcd", func(t *testing.T) {
		var err error
		node, err = New("pod1", "host1", &InvalidData)
		assert.Error(t, err)
		assert.Nil(t, node)
	})
	t.Run("no available subnet", func(t *testing.T) {
		var err error
		node, err = New("pod1", "host3", &validData)
		assert.Error(t, err)
		assert.Nil(t, node)
	})

}
