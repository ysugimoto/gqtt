package main

import (
	"context"
	"github.com/ysugimoto/gqtt"
	"github.com/ysugimoto/gqtt/message"
	"log"
	"time"
)

func main() {
	client := gqtt.NewClient("mqtt://localhost:9999")
	ctx := context.Background()
	if err := client.Connect(ctx); err != nil {
		log.Fatal(err)
	}
	log.Println("client connected")

	if err := client.Subscribe("gqtt/example", message.QoS0); err != nil {
		log.Fatal(err)
	}
	log.Println("subscribed")

	ticker := time.NewTicker(10 * time.Second)

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
		case <-ticker.C:
			log.Printf("message publish")
			if err := client.Publish("gqtt/example", message.QoS0, []byte("Hello, MQTT5!")); err != nil {
				log.Fatal(err)
			}
		}
	}
}
