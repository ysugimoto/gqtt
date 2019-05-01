package main

import (
	"context"
	"github.com/ysugimoto/gqtt"
	"github.com/ysugimoto/gqtt/message"
	"log"
)

func main() {
	server := gqtt.NewBroker(9999)
	ctx := context.Background()
	go server.ListenAndServe(ctx)
	for evt := range server.MessageEvent {
		switch e := evt.(type) {
		case *message.Subscribe:
			log.Println("Received SUBSCRIBE event: ", e.GetType())
		case *message.Connect:
			log.Println("Received CONNECT event", e.GetType())
		case *message.Publish:
			log.Println("Received PUBLISH event", e.GetType())
		}
	}
	<-ctx.Done()

}
