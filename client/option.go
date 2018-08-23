package client

import (
	"github.com/ysugimoto/gqtt/message"
)

type optionName string

const (
	nameBasicAuth optionName = "basic"
	nameLoginAuth optionName = "login"
	nameWill      optionName = "will"
)

type ClientOption struct {
	name  optionName
	value interface{}
}

func WithBasicAuth(user, password string) ClientOption {
	return ClientOption{
		name: nameBasicAuth,
		value: map[string]string{
			"user": user,
			"pass": password,
		},
	}
}

func WithLoginAuth(user, password string) ClientOption {
	return ClientOption{
		name: nameLoginAuth,
		value: map[string]string{
			"user": user,
			"pass": password,
		},
	}
}

func WithWill(qos message.QoSLevel, retain bool, topic, payload string, property *message.WillProperty) ClientOption {
	return ClientOption{
		name: nameWill,
		value: map[string]interface{}{
			"qos":      qos,
			"retain":   retain,
			"topic":    topic,
			"payload":  payload,
			"property": property,
		},
	}
}
