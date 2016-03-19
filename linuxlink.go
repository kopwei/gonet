package gonet

import (
	"fmt"
	"net"
)

// LinuxLink ...
type LinuxLink struct {
	link *net.Interface
}

// LinuxLinkByName is used to get the link object
func LinuxLinkByName(name string) (*LinuxLink, error) {
	ifc, err := net.InterfaceByName(name)
	if err != nil {
		return nil, fmt.Errorf("Failed to retrieve interface via name %s due to %s",
			name, err.Error())

	}
	return &LinuxLink{link: ifc}, err
}

// DeleteLink is used to delete
func DeleteLink(name string) error {
	return nil
}
