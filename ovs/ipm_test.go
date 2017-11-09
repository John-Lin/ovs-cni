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
	"testing"
)

func TestGenerateHostLocalConfig(t *testing.T) {
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
}
