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

func TestIfaceUp_Success(t *testing.T) {
	loopback := "lo"
	link, err := ifaceUp(loopback)
	assert.NotNilf(t, link, "error: %s should exist", loopback)
	assert.NoError(t, err)
}

func TestIfaceUp_Invalid(t *testing.T) {
	loopback := "unknown"
	link, err := ifaceUp(loopback)
	assert.Nilf(t, link, "error: %s should exist", loopback)
	assert.Error(t, err)
}
