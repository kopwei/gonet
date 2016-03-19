package main

import (
	"log"

	"github.com/kopwei/gonet"
)

func main() {
	_, err := gonet.NewVethLinkPair("endpointA", "endpointB")
	if err != nil {
		log.Fatalf("Fail to create veth link due to %s", err.Error())
	}
	log.Printf("Successfully created veth pair %s and %s", "endpointA", "endpointB")

}
