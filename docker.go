package gonet

import (
	"fmt"
	"net"

	"github.com/vishvananda/netns"
)

// SetPeerLinkToDockerNs is used to put the link into containers namespace with specified
// name
func (lnk *linuxLink) SetToDockerNs(containerID, newName string, ipaddr *net.IPNet) error {
	if containerID == "" {
		return fmt.Errorf("The container id cannot be empty")
	}
	nsHandle, err := netns.GetFromDocker(containerID)
	if err != nil {
		return fmt.Errorf("Failed to get container's network namespace due to %s", err.Error())
	}
	return lnk.putLinkIntoNetNS(nsHandle, newName, ipaddr)
}
