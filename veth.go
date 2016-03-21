package gonet

import (
	"fmt"
	"net"

	"github.com/vishvananda/netlink"
)

// VethLinkPair is the interface of linux veth link pair
type VethLinkPair interface {
	SetPeerIntoNetNS(netnspid int, newName string, ipaddr *net.IPNet) error
}

type vethLinkPair struct {
	IfcLink  LinuxLink
	PeerLink LinuxLink
}

// NewVethLinkPair ...
func NewVethLinkPair(ifcName, peerName string) (VethLinkPair, error) {
	linkAttr := netlink.LinkAttrs{Name: ifcName}
	vethLink := netlink.Veth{LinkAttrs: linkAttr, PeerName: peerName}
	err := netlink.LinkAdd(&vethLink)
	if err != nil {
		return nil, fmt.Errorf("Failed to create veth link %s and %s due to %s",
			ifcName, peerName, err.Error())
	}
	ifcLink, err := LinuxLinkByName(ifcName)
	if err != nil {
		return nil, fmt.Errorf("Failed to get veth endpoint %s link due to %s",
			ifcName, err.Error())
	}
	peerLink, err := LinuxLinkByName(peerName)
	if err != nil {
		return nil, fmt.Errorf("Failed to get veth endpoint %s link due to %s",
			peerName, err.Error())
	}
	return &vethLinkPair{IfcLink: ifcLink, PeerLink: peerLink}, nil
}

// SetPeerIntoNetNS is used to put the peer into a specific netns
func (veth *vethLinkPair) SetPeerIntoNetNS(netnspid int, newName string, ipaddr *net.IPNet) error {
	return veth.PeerLink.SetToNetNs(netnspid, newName, ipaddr)
}
