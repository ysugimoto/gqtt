package main

import (
	"context"
	"github.com/ysugimoto/gqtt"
	"github.com/ysugimoto/gqtt/message"
	"log"
	"os"
	"time"
)

func main() {
	var sig string = "default"
	if len(os.Args) > 1 {
		sig = os.Args[1]
	}

	client := gqtt.NewClient("mqtt://localhost:9999")
	defer client.Disconnect()

	ctx := context.Background()
	if err := client.Connect(ctx); err != nil {
		log.Fatal(err)
	}
	log.Println("client connected")

	if err := client.Subscribe("gqtt/example", message.QoS0); err != nil {
		log.Fatal(err)
	}
	log.Println("subscribed")

	ticker := time.NewTicker(3 * time.Second)

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
			if err := client.Publish("gqtt/example", message.QoS0, []byte("Hello, MQTT5! from "+sig)); err != nil {
				return
			}
		}
	}
}
