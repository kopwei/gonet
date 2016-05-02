package gonet

import (
	"fmt"
	"net"

	"github.com/vishvananda/netlink"
)

// VethLinkPair is the interface of linux veth link pair
type VethLinkPair interface {
	SetPeerIntoNetNS(netnspid int, newName string, ip net.IP, mask net.IPMask) error
}

type vethLinkPair struct {
	IfcLink  LinuxLink
	PeerLink LinuxLink
}

// NewVethLinkPair ...
func NewVethLinkPair(ifcName, peerName string) (VethLinkPair, error) {
    chnl := make(chan netlink.LinkUpdate)
    done := make(chan struct{})
	linkAttr := netlink.LinkAttrs{Name: ifcName}
	vethLink := netlink.Veth{LinkAttrs: linkAttr, PeerName: peerName}
    defer close(done)
    defer close(chnl)
    if err := netlink.LinkSubscribe(chnl, done); err != nil {
        return nil, fmt.Errorf("Failed to create watchdog channel for %s", ifcName)
    }
	err := netlink.LinkAdd(&vethLink)
    if idx := WaitForLink(chnl, ifcName, 0); idx == 0 {
        return nil, fmt.Errorf("Timeout waiting link %s", ifcName)
    }
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
func (veth *vethLinkPair) SetPeerIntoNetNS(netnspid int, newName string, ip net.IP, mask net.IPMask) error {
	return veth.PeerLink.SetToNetNs(netnspid, newName, ip, mask)
}
