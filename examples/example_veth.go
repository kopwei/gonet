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

	// Delete new link
	err = gonet.DeleteLink("endpointA")
	if err != nil {
		log.Fatalf("Fail to delete link %s due to %s", "endpointA", err.Error())
	}
	log.Printf("Successfully delete link with name %s", "endpointA")
}
