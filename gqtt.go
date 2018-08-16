package gqtt

import (
	"github.com/ysugimoto/gqtt/broker"
	"github.com/ysugimoto/gqtt/client"
	"github.com/ysugimoto/gqtt/message"
)

type Topic = message.SubscribeTopic

func NewBroker(port int) *broker.Broker {
	return broker.NewBroker(port)
}

func NewClient(url string) *client.Client {
	return client.NewClient(url)
}
