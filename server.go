package main

import (
	"github.com/ysugimoto/gqtt/broker"
	"log"
)

func main() {
	server := &broker.Broker{}
	log.Fatal(server.Run(9999))
}
