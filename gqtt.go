package gqtt

import (
	"github.com/ysugimoto/gqtt/broker"
	"github.com/ysugimoto/gqtt/client"
	"github.com/ysugimoto/gqtt/message"
)

type Topic = message.SubscribeTopic
type Option = client.ClientOption

func NewBroker(port int) *broker.Broker {
	return broker.NewBroker(port)
}

func NewClient(url string) *client.Client {
	return client.NewClient(url)
}

func WithBasicAuth(user, password string) Option {
	return client.WithBasicAuth(user, password)
}

func WithLoginAuth(user, password string) Option {
	return client.WithLoginAuth(user, password)
}

func WithWill(qos message.QoSLevel, retain bool, topic, payload string, property *message.WillProperty) Option {
	return client.WithWill(qos, retain, topic, payload, property)
}

func WithRetain() Option {
	return client.WithRetain()
}

func WithQoS(qos message.QoSLevel) Option {
	return client.WithQoS(qos)
}
