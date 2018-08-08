package main

import (
	"context"
	"github.com/ysugimoto/gqtt"
	"log"
)

func main() {
	server := gqtt.NewBroker(9999)
	log.Fatal(server.ListenAndServe(context.Background()))
}
