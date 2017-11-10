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
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFailGenerateHostLocalConfig(t *testing.T) {
	t.Run("syntex Error", func(t *testing.T) {
		newConfig := string(`
{syntex error
}
`)
		result := GenerateHostLocalConfig([]byte(newConfig))
		assert.Equal(t, "", string(result))
	})
	t.Run("etcd connection Error", func(t *testing.T) {
		newConfig := string(`
{
"ipam":{
"etcdURL": "127.0.0.1:9999"
}
}
`)
		result := GenerateHostLocalConfig([]byte(newConfig))
		assert.Equal(t, "", string(result))
	})

	t.Run("CIRD Error", func(t *testing.T) {
		newConfig := string(`
{
"ipam":{
"type":"central-ipm",
"subnetMin": "10.245.5.0.1",
"subnetMax": "10.245.5.0.1",
"etcdURL": "127.0.0.1:2379"
}
}
`)
		result := GenerateHostLocalConfig([]byte(newConfig))
		assert.Equal(t, "", string(result))
	})

	t.Run("Success Regisger", func(t *testing.T) {
		newConfig := string(`
{
"ipam":{
"type":"central-ipm",
"network":"10.245.0.0/16",
"subnetLen": 24,
"subnetMin": "10.245.5.0",
"subnetMax": "10.245.50.0",
"etcdURL": "127.0.0.1:2379"
}
}
`)

		expected := string(`
{
"ipam":{
"type":"host-local",
"subnet":"10.245.5.0/24"
}
}
`)

		result := GenerateHostLocalConfig([]byte(newConfig))
		assert.Equal(t, expected, string(result))
		//Make sure we can get the same result from same node
		result = GenerateHostLocalConfig([]byte(newConfig))
		assert.Equal(t, expected, string(result))
	})
}

func TestGetSubnet(t *testing.T) {
	t.Run("Next Subnet", func(t *testing.T) {
		newConfig := string(`
{
"ipam":{
"type":"central-ipm",
"network":"10.245.0.0/16",
"subnetLen": 24,
"subnetMin": "10.245.5.0",
"subnetMax": "10.245.6.0",
"etcdURL": "127.0.0.1:2379"
}
}
`)

		n := CentralNet{}
		err := json.Unmarshal([]byte(newConfig), &n)
		assert.NoError(t, err)
		expected := "10.245.6.0/24"

		subnet, _ := GetSubnet(*n.IPM, "ovs-cni-test")
		assert.Equal(t, expected, subnet.String())
	})
	t.Run("Full Subnet", func(t *testing.T) {
		newConfig := string(`
{
"ipam":{
"type":"central-ipm",
"network":"10.245.0.0/16",
"subnetLen": 24,
"subnetMin": "10.245.5.0",
"subnetMax": "10.245.6.0",
"etcdURL": "127.0.0.1:2379"
}
}
		`)

		n := CentralNet{}
		err := json.Unmarshal([]byte(newConfig), &n)
		assert.NoError(t, err)

		_, err = GetSubnet(*n.IPM, "ovs-cni-test-full")
		assert.Error(t, err)
	})
}