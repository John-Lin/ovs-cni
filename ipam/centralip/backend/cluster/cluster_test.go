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
	"github.com/John-Lin/ovs-cni/ipam/centralip/backend/utils"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
	"time"
)

var node *NodeIPM
var err error

var validData = utils.IPMConfig{
	Network: "10.123.0.0/16",
	ETCDURL: "127.0.0.1:2379",
}

func TestNewNode(t *testing.T) {
	if _, defined := os.LookupEnv("TEST_ETCD"); !defined {
		t.SkipNow()
		return
	}
	node, err = New("pod1", &validData)
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
	assert.Equal(t, "", gwIP)
}

func TestGetAvailableIP(t *testing.T) {
	if _, defined := os.LookupEnv("TEST_ETCD"); !defined {
		t.SkipNow()
		return
	}
	t.Run("First IP", func(t *testing.T) {
		ip, ipNet, err := node.GetAvailableIP()
		assert.NoError(t, err)
		assert.Equal(t, "10.123.0.2/16", ipNet.String())
		assert.Equal(t, "10.123.0.2", ip)
	})
	time.Sleep(1 * time.Second)
	t.Run("Second IP", func(t *testing.T) {
		ip, ipNet, err := node.GetAvailableIP()
		assert.NoError(t, err)
		assert.Equal(t, "10.123.0.3/16", ipNet.String())
		assert.Equal(t, "10.123.0.3", ip)
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
		assert.Equal(t, "10.123.0.2/16", ipNet.String())
		assert.Equal(t, "10.123.0.2", ip)
	})
}

func TestSecondHost(t *testing.T) {
	if _, defined := os.LookupEnv("TEST_ETCD"); !defined {
		t.SkipNow()
		return
	}
	node2, err := New("pod2", &validData)
	assert.NoError(t, err)

	gwIP, err := node2.GetGateway()
	assert.NoError(t, err)
	assert.Equal(t, "", gwIP)
	ip, ipNet, err := node2.GetAvailableIP()
	assert.NoError(t, err)
	assert.Equal(t, "10.123.0.4/16", ipNet.String())
	assert.Equal(t, "10.123.0.4", ip)

}

func TestGenerateCentralIPMInvalid(t *testing.T) {
	if _, defined := os.LookupEnv("TEST_ETCD"); !defined {
		t.SkipNow()
		return
	}
	t.Run("invalid etcd", func(t *testing.T) {
		var InvalidETCD = utils.IPMConfig{
			Network: "10.123.0.0/16",
			ETCDURL: "127.0.0.1:23792",
		}
		var err error
		node, err = New("pod1", &InvalidETCD)
		assert.Error(t, err)
		assert.Nil(t, node)
	})
	t.Run("invalid network", func(t *testing.T) {
		var InvalidNetwork = utils.IPMConfig{
			Network: "10.23.0.0/16ds",
			ETCDURL: "127.0.0.1:2379",
		}

		var err error
		node, err = New("pod1", &InvalidNetwork)
		assert.Error(t, err)
		assert.Nil(t, node)
	})
}
