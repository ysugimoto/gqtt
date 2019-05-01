package main

import (
	"context"
	"github.com/ysugimoto/gqtt"
	"github.com/ysugimoto/gqtt/message"
	"log"
)

func main() {
	client := gqtt.NewClient("mqtt://localhost:9999")
	defer client.Disconnect()

	ctx := context.Background()
	auth := gqtt.WithLoginAuth("admin", "admin")
	if err := client.Connect(ctx, auth); err != nil {
		log.Fatal(err)
	}
	log.Println("client connected")

	if err := client.Subscribe("some/will", message.QoS2); err != nil {
		log.Fatal(err)
	}
	log.Println("subscribed")

	for {
		select {
		case <-client.Closed:
			log.Println("connection closed")
			return
		case <-ctx.Done():
			log.Println("context canceled")
			return
		case msg := <-client.Message:
			log.Printf("published message received: %s\n", string(msg.Body))
		}
	}
}
