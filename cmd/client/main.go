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
	auth := gqtt.WithBasicAuth("admin", "admin")
	if err := client.Connect(ctx, auth); err != nil {
		log.Fatal(err)
	}
	log.Println("client connected")

	if err := client.Subscribe("gqtt/example", message.QoS2); err != nil {
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
			ticker.Stop()
			if err := client.Publish("gqtt/example", message.QoS2, []byte("Hello, MQTT5! from "+sig)); err != nil {
				return
			}
		}
	}
}
