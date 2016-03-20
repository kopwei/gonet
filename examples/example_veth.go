package main

import (
	"log"

	"github.com/kopwei/gonet"
)

func main() {
	// Create new link
	_, err := gonet.NewVethLinkPair("endpointA", "endpointB")
	if err != nil {
		log.Fatalf("Fail to create veth link due to %s", err.Error())
	}
	log.Printf("Successfully created veth pair %s and %s", "endpointA", "endpointB")
	/*
		ipNet := &net.IPNet{IP: net.ParseIP("192.168.0.3"), Mask: nil}
		err = gonet.SetPeerLinkToDockerNs(vpair.PeerLink, "", "eth0", ipNet)
		if err != nil {
			log.Fatalf("Failed to put peer link to container due to %s", err.Error())
		}
		log.Printf("Successfully add endpoint B for container\n")


		err = vpair.PeerLink.Ifconfig(net.ParseIP("192.168.0.3"), nil)
		if err != nil {
			log.Fatalf("Failed to configure the net ip due to %s", err.Error())
		}
		log.Printf("Successfully configured the %s's ip to %s", "endpointB", "192.168.0.3")
	*/

	// Delete new link
	err = gonet.DeleteLink("endpointA")
	if err != nil {
		log.Fatalf("Fail to delete link %s due to %s", "endpointA", err.Error())
	}
	log.Printf("Successfully delete link with name %s", "endpointA")

}
