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
	auth := gqtt.WithLoginAuth("admin", "admin")
	will := gqtt.WithWill(message.QoS0, false, "some/will", "send will", nil)
	if err := client.Connect(ctx, auth, will); err != nil {
		log.Fatal(err)
	}
	log.Println("client connected")

	if err := client.Subscribe("gqtt/example", message.QoS2); err != nil {
		log.Fatal(err)
	}
	log.Println("subscribed")

	ticker := time.NewTicker(3 * time.Second)
	times := 0

	for {
		select {
		case <-client.Closed:
			log.Println("connection closed")
			return
		case <-ctx.Done():
			log.Println("context canceled")
			return
		case msg := <-client.Message:
			if msg.RETAIN {
				log.Printf("received retain message")
			}
			log.Printf("published message received: %s\n", string(msg.Body))
		case <-ticker.C:
			// ticker.Stop()

			times++
			var err error
			if times%2 > 0 {
				log.Printf("message publish")
				err = client.Publish(
					"gqtt/example",
					[]byte("Hello, MQTT5! from "+sig),
					gqtt.WithRetain(),
				)
			} else {
				log.Printf("message publish with clear retain")
				err = client.Publish(
					"gqtt/example",
					[]byte{},
					gqtt.WithRetain(),
				)
			}
			if err != nil {
				return
			}
		}
	}
}
