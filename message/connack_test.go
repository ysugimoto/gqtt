package message_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ysugimoto/gqtt/message"
)

func TestConnAckMessageEncodeDecode(t *testing.T) {
	c := message.NewConnAck(message.Success)
	c.ConnAckProperty = &message.ConnAckProperty{
		SessionExpiryInterval:    1000,
		AssignedClientIdentifier: "gqtt-test",
		ServerKeepAlive:          10000,
		AuthenticationMethod:     "BASIC",
		AuthenticationData:       []byte("gqttauth"),
		ResponseInformation:      "thisisresponse",
		ServerReference:          "mqtt://example.com",
		ReasonString:             "some extra reason",
		ReceiveMaximum:           100,
		TopicAliasMaximum:        100,
		MaximumQoS:               1,
		RetainAvalilable:         true,
		UserProperty: map[string]string{
			"foo": "bar",
		},
		MaximumPacketSize:              65535,
		WildcardSubscriptionAvailable:  true,
		SubscrptionIdentifierAvailable: false,
		SharedSubscriptionsAvaliable:   true,
	}
	buf, err := c.Encode()
	assert.NoError(t, err)

	f, p, err := message.ReceiveFrame(bytes.NewReader(buf))
	assert.NoError(t, err)
	assert.Exactly(t, f.Type, message.CONNACK)
	assert.Exactly(t, f.DUP, false)
	assert.Equal(t, f.QoS, uint8(0))
	assert.Exactly(t, f.RETAIN, false)

	ca, err := message.ParseConnAck(f, p)
	assert.NoError(t, err)
	assert.Equal(t, ca.ReasonCode, message.Success)
	assert.NotNil(t, ca.ConnAckProperty)
	prop := ca.ConnAckProperty
	assert.Equal(t, uint32(1000), prop.SessionExpiryInterval)
	assert.Equal(t, "gqtt-test", prop.AssignedClientIdentifier)
	assert.Equal(t, uint16(10000), prop.ServerKeepAlive)
	assert.Equal(t, "BASIC", prop.AuthenticationMethod)
	assert.Equal(t, []byte("gqttauth"), prop.AuthenticationData)
	assert.Equal(t, "thisisresponse", prop.ResponseInformation)
	assert.Equal(t, "mqtt://example.com", prop.ServerReference)
	assert.Equal(t, uint16(100), prop.ReceiveMaximum)
	assert.Equal(t, uint16(100), prop.TopicAliasMaximum)
	assert.Equal(t, uint8(1), prop.MaximumQoS)
	assert.True(t, prop.RetainAvalilable)
	assert.NotNil(t, prop.UserProperty)
	assert.Contains(t, prop.UserProperty, "foo")
	assert.Equal(t, prop.UserProperty["foo"], "bar")
	assert.Equal(t, uint32(65535), prop.MaximumPacketSize)
	assert.True(t, prop.WildcardSubscriptionAvailable)
	assert.False(t, prop.SubscrptionIdentifierAvailable)
	assert.True(t, prop.SharedSubscriptionsAvaliable)
}
