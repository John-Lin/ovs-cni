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

package utils

import (
	"github.com/stretchr/testify/assert"
	"net"
	"testing"
)

func TestPowOfTwo(t *testing.T) {
	assert.Equal(t, uint32(2), powTwo(1))
	assert.Equal(t, uint32(1), powTwo(0))
	assert.Equal(t, uint32(1024), powTwo(10))
	assert.Equal(t, uint32(2147483648), powTwo(31))
}

func TestIp2Int(t *testing.T) {
	v4Input := net.ParseIP("127.0.0.1")
	result, err := ipToInt(v4Input)
	assert.NoError(t, err)
	assert.Equal(t, uint32(2130706433), result)

	v6Input := net.ParseIP("2001:0DB8:02de:0000:0000:0000:0000:0e13")
	result, err = ipToInt(v6Input)
	assert.Error(t, err)
}

func TestInt2IP(t *testing.T) {
	input := intToIP(2130706433)
	assert.Equal(t, "127.0.0.1", input.String())

}

func TestGetGatewayFromIP(t *testing.T) {
	_, input, _ := net.ParseCIDR("192.168.194.0/22")

	gwIP := getNextIP(input)
	assert.Equal(t, gwIP.String(), "192.168.192.1")
}
