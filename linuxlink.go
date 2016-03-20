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
	link netlink.Link
	//ifc  *net.Interface
}

// Up is used to set the link to up state
func (linuxLink *LinuxLink) Up() error {
	return netlink.LinkSetUp(linuxLink.link)
}

// Down is used to set the link to up state
func (linuxLink *LinuxLink) Down() error {
	return netlink.LinkSetDown(linuxLink.link)
}

// SetName is used to set the link to up state
func (linuxLink *LinuxLink) SetName(name string) error {
	if name == "" {
		return fmt.Errorf("The link name cannot be empty")
	}
	return netlink.LinkSetName(linuxLink.link, name)
}

// Ifconfig is used to configure the basic ip of the link
func (linuxLink *LinuxLink) Ifconfig(ip net.IP, netmask net.IPMask) error {
	if ip == nil {
		return fmt.Errorf("Failed to configure the IP since the ip is not valid")
	}
	if netmask == nil {
		netmask = ip.DefaultMask()
	}
	ipNet := &net.IPNet{IP: ip, Mask: netmask}
	netAddr := &netlink.Addr{IPNet: ipNet}
	return netlink.AddrAdd(linuxLink.link, netAddr)
}

// LinuxLinkByName is used to get the link object
func LinuxLinkByName(name string) (*LinuxLink, error) {
	link, err := netlink.LinkByName(name)
	if err != nil {
		return nil, fmt.Errorf("Failed to retrieve link via name %s due to %s",
			name, err.Error())

	}
	return &LinuxLink{ /*ifc: ifc, */ link: link}, err
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
func PutLinkIntoNetNs(link *LinuxLink, nspid int, newName string, ipaddr *net.IPNet) error {
	if newName == "" {
		return fmt.Errorf("The new name cannot be empty")
	}

	newNsHandle, err := netns.GetFromPid(nspid)
	if err != nil {
		return fmt.Errorf("Failed to get the net ns for pid %d due to %s", nspid, err.Error())
	}

	return putLinkIntoNetNS(link, newNsHandle, newName, ipaddr)
}

func putLinkIntoNetNS(link *LinuxLink, nsHandle netns.NsHandle, newName string, ipaddr *net.IPNet) error {
	err := link.Down()
	if err != nil {
		return fmt.Errorf("Failed to put link down due to %s", err.Error())
	}
	currentNetNs, err := netns.Get()
	defer netns.Setns(currentNetNs, syscall.CLONE_NEWNET)
	if err != nil {
		return fmt.Errorf("Failed to get current net ns due to %s", err.Error())
	}
	err = netlink.LinkSetNsFd(link.link, int(nsHandle))
	if err != nil {
		return fmt.Errorf("Failed to set net ns %d due to %s", nsHandle, err.Error())
	}
	err = netns.Set(nsHandle)
	if err != nil {
		return fmt.Errorf("Failed to switch to set net ns %d due to %s", nsHandle, err.Error())
	}
	if newName != link.link.Attrs().Name {
		err = link.SetName(newName)
		if err != nil {
			return fmt.Errorf("Failed to set the link to new name %s due to %s",
				newName, err.Error())
		}
	}

	if ipaddr != nil {
		err = link.Ifconfig(ipaddr.IP, ipaddr.Mask)
		if err != nil {
			return fmt.Errorf("Failed to configure the links ip due to %s", err.Error())
		}
	}
	return link.Up()
}
