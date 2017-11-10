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
	"time"
)

var ovsSwitch *OVSSwitch
var bridgeName string = "ovs_cni"

func TestNewOVSSwitch(t *testing.T) {
	var err error
	ovsSwitch, err = NewOVSSwitch(bridgeName)
	assert.NoError(t, err)
}

func TestAddPort(t *testing.T) {
	err := ovsSwitch.addPort("test")
	assert.NoError(t, err)
}

func TestAddVTEPs(t *testing.T) {
	err := ovsSwitch.AddVTEPs([]string{"10.16.1.1"})
	assert.NoError(t, err)
}

func TestAddPort_Invalid(t *testing.T) {
	err := ovsSwitch.addPort("")
	assert.Error(t, err)
}

func TestSetCtrl(t *testing.T) {
	err := ovsSwitch.SetCtrl("10.1.1.1:6653")
	assert.NoError(t, err)
}

func TestSetCtrl_Invalid(t *testing.T) {
	err := ovsSwitch.SetCtrl("abc")
	assert.Error(t, err)
	err = ovsSwitch.SetCtrl("10.1.1.1:abcde")
	assert.Error(t, err)
}

func TestDeleteOVSSwitch(t *testing.T) {
	err := ovsSwitch.Delete()
	assert.NoError(t, err)
}

func TestDeleteOVSSwitch_Invalid(t *testing.T) {
	//wait previous delete
	time.Sleep(1000 * time.Millisecond)
	err := ovsSwitch.Delete()
	assert.Error(t, err)
}

func TestCreateOVS(t *testing.T) {
	netConfig := NetConf{OVSBrName: "test0"}

	ovs, cT, err := createOVS(&netConfig)
	assert.NoError(t, err)
	assert.Equal(t, "test0", cT.Name)
	//wait previous delete
	time.Sleep(300 * time.Millisecond)
	ovs.Delete()
}
