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
	"encoding/json"
	"fmt"
	"os"
	// "github.com/John-Lin/ovs-cni/ipam/centralip/backend/cluster"
	"github.com/John-Lin/ovs-cni/ipam/centralip/backend/node"
	"github.com/John-Lin/ovs-cni/ipam/centralip/backend/utils"
	"github.com/containernetworking/cni/pkg/skel"
)

type CentralNet struct {
	Name       string           `json:"name"`
	CNIVersion string           `json:"cniVersion"`
	IPM        *utils.IPMConfig `json:"ipam"`
}

func GenerateCentralIPM(args *skel.CmdArgs) (utils.CentralIPM, error, string) {
	n := &CentralNet{}
	if err := json.Unmarshal(args.StdinData, n); err != nil {
		return nil, fmt.Errorf("failed to load netconf: %v", err), ""
	}

	switch n.IPM.IPType {
	case "node":
		hostname, _ := os.Hostname()
		node, err := node.New(args.ContainerID, hostname, n.IPM)

		return node, err, n.CNIVersion
	case "cluster":
		return nil, nil, ""
	default:
		return nil, fmt.Errorf("Unsupport IPM type %s", n.IPM.Type), ""
	}
}
