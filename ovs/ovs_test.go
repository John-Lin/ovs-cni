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

func TestLoadNetConf(t *testing.T) {

	t.Run("Valid", func(t *testing.T) {
		config := string(`
		{
			"name":"mynet",
			"cniVersion":"0.3.1",
			"type":"ovs",
			"ovsBridge":"br0",
			"isDefaultGateway": true,
			"ipMasq": true
		}
		`)

		n, s, err := loadNetConf([]byte(config))
		assert.NoError(t, err)
		assert.Equal(t, s, "0.3.1")
		assert.Equal(t, "br0", n.OVSBrName)
	})
	t.Run("InValid", func(t *testing.T) {
		config := string(`
		{
			asddad
		}
		`)

		n, s, err := loadNetConf([]byte(config))
		assert.Error(t, err)
		assert.Nil(t, n)
		assert.Equal(t, s, "")
	})

}
