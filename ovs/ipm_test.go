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

package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

var n *CentralIPM
const validData string = `
	{
	"network":"10.245.0.0/16",
	"subnetLen": 24,
	"subnetMin": "10.245.5.0",
	"subnetMax": "10.245.50.0",
	"etcdURL": "127.0.0.1:2379"
	}
	`

func TestGenerateCentralIPM(t *testing.T) {
	var err error
	n, err = generateCentralIPM([]byte(validData))
	assert.NoError(t, err)
	assert.Equal(t, n.ETCDURL, "127.0.0.1:2379")

	err = n.Init("test0", "pod1")
	assert.NoError(t, err)
}

func TestGetGateway(t *testing.T) {
	gwIP, err := n.GetGateway()
	assert.NoError(t, err)
	assert.Equal(t, "10.245.5.1", gwIP)
}

func TestGetAvailableIP(t *testing.T) {
	t.Run("First IP", func(t *testing.T) {
		IP, err := n.GetAvailableIP()
		assert.NoError(t, err)
		assert.Equal(t, "10.245.5.2", IP)
	})
	t.Run("Second IP", func(t *testing.T) {
		IP, err := n.GetAvailableIP()
		assert.NoError(t, err)
		assert.Equal(t, "10.245.5.3", IP)
	})
}

func TestSecondSubnet(t *testing.T) {
	ipm, err := generateCentralIPM([]byte(validData))
	assert.NoError(t, err)

	err = ipm.Init("test1", "pod2")
	assert.NoError(t, err)

	gwIP, err := ipm.GetGateway()
	assert.NoError(t, err)
	assert.Equal(t, "10.245.6.1", gwIP)
}
