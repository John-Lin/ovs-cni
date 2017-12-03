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
	"github.com/containernetworking/cni/pkg/skel"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

var validNodeData = skel.CmdArgs{
	StdinData: []byte(`
	{
		"name":"mynet",
		"cniVersion":"0.3.1",
		"ipam":{
			"type":"central",
			"ipType": "node",
			"network":"10.245.0.0/16",
			"subnetLen": 24,
			"subnetMin": "10.245.5.0",
			"subnetMax": "10.245.6.0",
			"etcdURL": "127.0.0.1:2379"
		}
	}
	`),
}

var InvalidData = skel.CmdArgs{
	StdinData: []byte(`
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
	`),
}

var validClusterData = skel.CmdArgs{
	StdinData: []byte(`
	{
		"name":"mynet",
		"cniVersion":"0.3.1",
		"ipam":{
			"type":"central",
			"ipType": "cluster",
			"network":"10.245.0.0/16",
			"subnetLen": 24,
			"subnetMin": "10.245.5.0",
			"subnetMax": "10.245.6.0",
			"etcdURL": "127.0.0.1:2379"
		}

	}
	`),
}

func TestGenerateCentralIPM(t *testing.T) {
	if _, defined := os.LookupEnv("TEST_ETCD"); !defined {
		t.SkipNow()
		return
	}
	t.Run("Node instance", func(t *testing.T) {
		n, err, version := GenerateCentralIPM(&validNodeData)
		assert.NoError(t, err)
		assert.NotNil(t, n)
		assert.Equal(t, version, "0.3.1")
	})
	t.Run("Cluster instance", func(t *testing.T) {
		n, err, version := GenerateCentralIPM(&validClusterData)
		assert.NoError(t, err)
		assert.NotNil(t, n)
		assert.Equal(t, version, "0.3.1")
	})
}

func TestGenerateInvalidCentralIPM(t *testing.T) {
	if _, defined := os.LookupEnv("TEST_ETCD"); !defined {
		t.SkipNow()
		return
	}
	n, err, _ := GenerateCentralIPM(&InvalidData)
	assert.Error(t, err)
	assert.Nil(t, n)
}
