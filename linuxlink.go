package gonet

import (
	"bytes"
	"fmt"
	"net"
	"syscall"
	"time"

	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
)

// LinuxLink is the main interface towards the outside
// It describes the API of the link
type LinuxLink interface {
	Up() error
	Down() error
	SetName(name string) error
	Ifconfig(ip net.IP, netmask net.IPMask) error
	SetToNetNs(nspid int, newName string, ip net.IP, mask net.IPMask) error
	SetToDockerNs(containerID, newName string, ip net.IP, mask net.IPMask) error
}

// LinuxLink ...
type linuxLink struct {
	link netlink.Link
	//ifc  *net.Interface
}

// Link operations watchdog
func WaitForLink(ch <-chan netlink.LinkUpdate, ifaceName string, Flags uint32) int {
	for {
		timeout := time.After(time.Minute)
		select {
		case update := <-ch:
			if ifaceName == update.Link.Attrs().Name {
                if (Flags <= 0) || (Flags & update.IfInfomsg.Flags != 0) {
                    return update.Link.Attrs().Index
                }
			}
		case <-timeout:
			return 0
		}
	}
}

// Up is used to set the link to up state
func (lnk *linuxLink) Up() error {
	return netlink.LinkSetUp(lnk.link)
}

// Down is used to set the link to up state
func (lnk *linuxLink) Down() error {
	return netlink.LinkSetDown(lnk.link)
}

// SetName is used to set the link to up state
func (lnk *linuxLink) SetName(name string) error {
	if name == "" {
		return fmt.Errorf("The link name cannot be empty")
	}
	return netlink.LinkSetName(lnk.link, name)
}

// Ifconfig is used to configure the basic ip of the link
func (lnk *linuxLink) Ifconfig(ip net.IP, netmask net.IPMask) error {
	if ip == nil {
		return fmt.Errorf("Failed to configure the IP since the ip is not valid")
	}
	if netmask == nil {
		netmask = ip.DefaultMask()
	}
	ipNet := &net.IPNet{IP: ip, Mask: netmask}
	netAddr := &netlink.Addr{IPNet: ipNet}
	return netlink.AddrAdd(lnk.link, netAddr)
}

// LinuxLinkByName is used to get the link object
func LinuxLinkByName(name string) (LinuxLink, error) {
	link, err := netlink.LinkByName(name)
	if err != nil {
		return nil, fmt.Errorf("Failed to retrieve link via name %s due to %s",
			name, err.Error())

	}
	return &linuxLink{ /*ifc: ifc, */ link: link}, err
}

// DeleteLink is used to delete the link object
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

// SetToNetNs is used to put a network interface into netns
func (lnk *linuxLink) SetToNetNs(nspid int, newName string, ip net.IP, mask net.IPMask) error {
	if newName == "" {
		return fmt.Errorf("The new name cannot be empty")
	}

	newNsHandle, err := netns.GetFromPid(nspid)
	if err != nil {
		return fmt.Errorf("Failed to get the net ns for pid %d due to %s", nspid, err.Error())
	}

	return lnk.putLinkIntoNetNS(newNsHandle, newName, ip, mask)
}

// Warning! This method is not threadsafe!
func (lnk *linuxLink) putLinkIntoNetNS(nsHandle netns.NsHandle, newName string, ip net.IP, mask net.IPMask) error {
	err := lnk.Down()

	if err != nil {
		return fmt.Errorf("Failed to put link down due to %s", err.Error())
	}
	currentNetNs, err := netns.Get()
    defer netns.Setns(currentNetNs, syscall.CLONE_NEWNET)
	if err != nil {
		return fmt.Errorf("Failed to get current net ns due to %s", err.Error())
	}
	err = netlink.LinkSetNsFd(lnk.link, int(nsHandle))
	if err != nil {
		return fmt.Errorf("Failed to set net ns %d due to %s", nsHandle, err.Error())
	}
	err = netns.Set(nsHandle)
	if err != nil {
		return fmt.Errorf("Failed to switch to set net ns %d due to %s", nsHandle, err.Error())
	}

    chnl := make(chan netlink.LinkUpdate)
    done := make(chan struct{})
    defer close(done)
    defer close(chnl)

    if err := netlink.LinkSubscribe(chnl, done); err != nil {
        return fmt.Errorf("Failed to create ns watchdog channel for %s", newName)
    }
    if idx := WaitForLink(chnl, newName, 0); idx == 0 {
        return fmt.Errorf("Timeout waiting link %s", newName)
    }

	if newName != lnk.link.Attrs().Name {
        return fmt.Errorf("Error renaming link %s to %s",
                          lnk.link.Attrs().Name,
                          newName)
	}

	if ip != nil {
        if err := lnk.Ifconfig(ip, mask); err != nil {
			return fmt.Errorf("Failed to configure address on %s", newName)
		}
	}

    if idx := WaitForLink(chnl, newName, 1); idx == 0 {
        return fmt.Errorf("Timeout waiting link %s up", newName)
    }
	return nil
}
