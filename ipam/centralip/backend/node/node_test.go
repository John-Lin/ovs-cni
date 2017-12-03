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
//	"github.com/stretchr/testify/assert"
//	"testing"
)

/*
const validData string = `
	{
		"name":"mynet",
		"cniVersion":"0.3.1",
		"ipam":{
			"type":"central",
			"network":"10.245.0.0/16",
			"subnetLen": 24,
			"subnetMin": "10.245.5.0",
			"subnetMax": "10.245.6.0",
			"etcdURL": "127.0.0.1:2379"
		}
	}
	`

func TestGenerateCentralIPM(t *testing.T) {
	var err error
	var version string
	n, version, err = GenerateCentralIPM([]byte(validData))
	assert.NoError(t, err)
	assert.Equal(t, n.ETCDURL, "127.0.0.1:2379")
	assert.Equal(t, version, "0.3.1")

	err = n.Init("test0", "pod1")
	assert.NoError(t, err)
	err = n.Init("test0", "pod1")
	assert.NoError(t, err)
}

func TestGetGateway(t *testing.T) {
	gwIP, err := n.GetGateway()
	assert.NoError(t, err)
	assert.Equal(t, "10.245.5.1", gwIP)
	gwIP, err = n.GetGateway()
	assert.NoError(t, err)
	assert.Equal(t, "10.245.5.1", gwIP)
}

func TestGetAvailableIP(t *testing.T) {
	t.Run("First IP", func(t *testing.T) {
		ip, ipNet, err := n.GetAvailableIP()
		assert.NoError(t, err)
		assert.Equal(t, "10.245.5.2/24", ipNet.String())
		assert.Equal(t, "10.245.5.2", ip)
	})
	t.Run("Second IP", func(t *testing.T) {
		ip, ipNet, err := n.GetAvailableIP()
		assert.NoError(t, err)
		assert.Equal(t, "10.245.5.3/24", ipNet.String())
		assert.Equal(t, "10.245.5.3", ip)
	})
	t.Run("remove first IP", func(t *testing.T) {
		err := n.DeleteIPByName("pod1")
		assert.NoError(t, err)
	})
	t.Run("Fetch IP again", func(t *testing.T) {
		ip, ipNet, err := n.GetAvailableIP()
		assert.NoError(t, err)
		assert.Equal(t, "10.245.5.2/24", ipNet.String())
		assert.Equal(t, "10.245.5.2", ip)
	})
}

func TestSecondSubnet(t *testing.T) {
	ipm, _, err := GenerateCentralIPM([]byte(validData))
	assert.NoError(t, err)

	err = ipm.Init("test1", "pod2")
	assert.NoError(t, err)

	gwIP, err := ipm.GetGateway()
	assert.NoError(t, err)
	assert.Equal(t, "10.245.6.1", gwIP)
}

func TestGenerateCentralIPMInvalid(t *testing.T) {
	t.Run("invalid config", func(t *testing.T) {
		const inValidData string = `
		{
			asdd
		}
		`
		var err error
		n, _, err = GenerateCentralIPM([]byte(inValidData))
		assert.Error(t, err)
	})
	t.Run("invalid etcd", func(t *testing.T) {
		const inValidData string = `
		{
			"ipam":{
				"type":"central",
				"network":"10.245.0.0/16",
				"subnetLen": 24,
				"subnetMin": "10.245.5.0",
				"subnetMax": "10.245.6.0",
				"etcdURL": "127.0.0.1:23791"
			}
		}
		`
		var err error
		n, _, err = GenerateCentralIPM([]byte(inValidData))
		assert.NoError(t, err)
		err = n.Init("test_invalid", "pod0")
		assert.Error(t, err)
	})
	t.Run("invalid call(init first)", func(t *testing.T) {
		const inValidData string = `

		{
			"ipam":{
				"type":"central",
				"network":"10.245.0.0/16",
				"subnetLen": 24,
				"subnetMin": "10.245.5.0",
				"subnetMax": "10.245.50.0",
				"etcdURL": "127.0.0.1:2379"
			}
		}
		`
		var err error
		n, _, err = GenerateCentralIPM([]byte(inValidData))
		assert.NoError(t, err)
		_, err = n.GetGateway()
		assert.Error(t, err)
		_, _, err = n.GetAvailableIP()
		assert.Error(t, err)

	})
	t.Run("no available subnet", func(t *testing.T) {
		var err error
		ipm, _, err := GenerateCentralIPM([]byte(validData))
		assert.NoError(t, err)

		err = ipm.Init("test3", "pod2")
		assert.Error(t, err)
	})
}
*/
