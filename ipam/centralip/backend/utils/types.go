package utils

import (
	"net"
)

type IPMConfig struct {
	Type      string `json:"type"`
	IPType    string `json:"ipType"`
	Network   string `json:"network"`
	SubnetLen int    `json:"subnetLen"`
	SubnetMin string `json:"subnetMin"`
	SubnetMax string `json:"subnetMax"`
	ETCDURL   string `json:"etcdURL"`
}

type CentralIPM interface {
	GetGateway() (string, error)
	GetAvailableIP() (string, *net.IPNet, error)
	Delete() error
}

const ETCDPrefix string = "/ovs-cni/networks/"
