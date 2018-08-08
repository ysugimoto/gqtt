package gqtt

import (
	"github.com/ysugimoto/gqtt/broker"
	"github.com/ysugimoto/gqtt/client"
)

func NewBroker(port int) *broker.Broker {
	return broker.NewBroker(port)
}

func NewClient(url string) *client.Client {
	return client.NewClient(url)
}
