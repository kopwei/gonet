package gonet

import (
	"fmt"

	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
)

// VethLinkPair is the structure of linux veth link pair
type VethLinkPair struct {
	IfcLink  *LinuxLink
	PeerLink *LinuxLink
}

// NewVethLinkPair ...
func NewVethLinkPair(ifcName, peerName string) (*VethLinkPair, error) {
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
	return &VethLinkPair{IfcLink: ifcLink, PeerLink: peerLink}, nil
}

// SetPeerIntoNetNS is used to put the peer into a specific netns
func (veth *VethLinkPair) SetPeerIntoNetNS(netnspid int, newName string) error {
	return PutLinkIntoNetNs(veth.PeerLink, netnspid, newName)
}

// SetPeerLinkForDocker is used to put the peer link into containers namespace with specified
// name
func (veth *VethLinkPair) SetPeerLinkForDocker(containerID string, newName string) error {
	if containerID == "" {
		return fmt.Errorf("The container id cannot be empty")
	}
	nsHandle, err := netns.GetFromDocker(containerID)
	if err != nil {
		return fmt.Errorf("Failed to get container's network namespace due to %s", err.Error())
	}
	return putLinkIntoNetNS(veth.PeerLink, nsHandle, newName)
}
