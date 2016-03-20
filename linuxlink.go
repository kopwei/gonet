// +build linux

package gonet

import (
	"fmt"
	"net"
	"syscall"

	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
)

// LinuxLink ...
type LinuxLink struct {
	ifc *net.Interface
}

func (linuxLink *LinuxLink) getNetLink() (netlink.Link, error) {
	netlk, err := netlink.LinkByName(linuxLink.ifc.Name)
	if err != nil {
		return nil, fmt.Errorf("Failed to get the netlink with name %s due to %s",
			linuxLink.ifc.Name, err.Error())
	}
	return netlk, err
}

// Up is used to set the link to up state
func (linuxLink *LinuxLink) Up() error {
	netlk, err := linuxLink.getNetLink()
	if err != nil {
		return fmt.Errorf("Error occurred in fetching net link due to %s", err.Error())
	}
	return netlink.LinkSetUp(netlk)
}

// SetName is used to set the link to up state
func (linuxLink *LinuxLink) SetName(name string) error {
	if name == "" {
		return fmt.Errorf("The link name cannot be empty")
	}
	netlk, err := linuxLink.getNetLink()
	if err != nil {
		return fmt.Errorf("Error occurred in fetching net link due to %s", err.Error())
	}
	return netlink.LinkSetName(netlk, name)
}

// LinuxLinkByName is used to get the link object
func LinuxLinkByName(name string) (*LinuxLink, error) {
	ifc, err := net.InterfaceByName(name)
	if err != nil {
		return nil, fmt.Errorf("Failed to retrieve interface via name %s due to %s",
			name, err.Error())

	}
	return &LinuxLink{ifc: ifc}, err
}

// DeleteLink is used to delete
func DeleteLink(name string) error {
	if name == "" {
		return fmt.Errorf("The name of the link is not valid")
	}
	netLnk, err := netlink.LinkByName(name)
	if err != nil {
		return fmt.Errorf("Failed to find the link with name %s due to %s", name, err.Error())
	}
	return netlink.LinkDel(netLnk)
}

// PutLinkIntoNetNs is used to put a network interface into netns
func PutLinkIntoNetNs(link *LinuxLink, nspid int, newName string) error {
	if newName == "" {
		return fmt.Errorf("The new name cannot be empty")
	}
	currentNetNs, err := netns.Get()
	defer netns.Setns(currentNetNs, syscall.CLONE_NEWNET)
	if err != nil {
		return fmt.Errorf("Failed to get current net ns due to %s", err.Error())
	}
	newNetNs, err := netns.GetFromPid(nspid)
	if err != nil {
		return fmt.Errorf("Failed to get the net ns for pid %d due to %s", nspid, err.Error())
	}

	return putLinkIntoNetNS(link, newNetNs, newName)
}

func putLinkIntoNetNS(link *LinuxLink, nsHandle netns.NsHandle, newName string) error {
	err := netns.Set(nsHandle)
	if err != nil {
		return fmt.Errorf("Failed to set net ns %d due to %s", nsHandle, err.Error())
	}
	err = link.Up()
	if err != nil {
		return fmt.Errorf("Failed to set the link up due to %s", err.Error())
	}
	if newName != link.ifc.Name {
		err = link.SetName(newName)
		if err != nil {
			return fmt.Errorf("Failed to set the link to new name %s due to %s",
				newName, err.Error())
		}
	}
	return nil

}
