package main

import (
	"errors"
	"fmt"
	"github.com/John-Lin/ovsdb"
	log "github.com/sirupsen/logrus"
	"net"
	"strconv"
	"time"
)

// OVSSwitch is a bridge instance
type OVSSwitch struct {
	NodeType     string
	BridgeName   string
	CtrlHostPort string
	ovsdb        *ovsdb.OvsDriver
}

// NewOVSSwitch for creating a ovs bridge
func NewOVSSwitch(bridgeName string) (*OVSSwitch, error) {
	sw := new(OVSSwitch)
	sw.NodeType = "OVSSwitch"
	sw.BridgeName = bridgeName

	sw.ovsdb = ovsdb.NewOvsDriverWithUnix(bridgeName)
	log.Infoln("Adding a switch:", sw.BridgeName)

	// Check if port is already part of the OVS and add it
	if !sw.ovsdb.IsPortNamePresent(bridgeName) {
		// Create an internal port in OVS
		err := sw.ovsdb.CreatePort(bridgeName, "internal", 0)
		if err != nil {
			return nil, err
		}
	}

	time.Sleep(300 * time.Millisecond)
	// log.Infof("Waiting for OVS bridge %s setup", bridgeName)

	// ip link set ovs up
	_, err := ifaceUp(bridgeName)
	if err != nil {
		return nil, err
	}

	return sw, nil
}

// addPort for asking OVSDB driver to add the port
func (sw *OVSSwitch) addPort(ifName string) error {
	if !sw.ovsdb.IsPortNamePresent(ifName) {
		err := sw.ovsdb.CreatePort(ifName, "", 0)
		if err != nil {
			log.Fatalf("Error creating the port. Err: %v", err)
			return err
		}
	}
	return nil
}

// SetCtrl for seting up OpenFlow controller for ovs bridge
func (sw *OVSSwitch) SetCtrl(hostport string) error {
	host, port, err := net.SplitHostPort(hostport)
	if err != nil {
		log.Fatalf("Invalid controller IP and port. Err: %v", err)
		return err
	}
	uPort, err := strconv.ParseUint(port, 10, 32)
	if err != nil {
		log.Fatalf("Invalid controller port number. Err: %v", err)
		return err
	}
	err = sw.ovsdb.AddController(host, uint16(uPort))
	if err != nil {
		log.Fatalf("Error adding controller to OVS. Err: %v", err)
		return err
	}
	sw.CtrlHostPort = hostport
	return nil
}

func (sw *OVSSwitch) Delete() error {
	if exist := sw.ovsdb.IsBridgePresent(sw.BridgeName); exist != true {
		return errors.New(sw.BridgeName + " doesn't exist, we can delete")
	}

	return sw.ovsdb.DeleteBridge(sw.BridgeName)
}

func (sw *OVSSwitch) AddVETPs(VtepIPs []string) error {
	for _, v := range VtepIPs {
		intfName := vxlanIfName(v)
		isPresent, vsifName := sw.ovsdb.IsVtepPresent(v)

		if !isPresent || (vsifName != intfName) {
			//create VTEP
			err := sw.ovsdb.CreateVtep(intfName, v)
			if err != nil {
				return fmt.Errorf("Error creating VTEP port %s. Err: %v", intfName, err)
			}

		}
	}
	return nil
}
