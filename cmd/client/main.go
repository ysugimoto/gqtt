package main

import (
	"context"
	"github.com/ysugimoto/gqtt"
	"github.com/ysugimoto/gqtt/message"
	"log"
)

func main() {
	client := gqtt.NewClient("mqtt://localhost:9999")
	if err := client.Connect(context.Background()); err != nil {
		log.Fatal(err)
	}
	log.Println("client connected")

	if err := client.Subscribe("gqtt/example", message.QoS0); err != nil {
		log.Fatal(err)
	}
	log.Println("subscribed")

	if err := client.Publish("gqtt/example", message.QoS0, []byte("Hello, MQTT5!")); err != nil {
		log.Fatal(err)
	}
	log.Println("published")

	<-client.Closed
}
