package message_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ysugimoto/gqtt/message"
)

func TestConnectEncodeErrorIfClientIdIsEmpty(t *testing.T) {
	buf, err := message.NewConnect().Encode()
	assert.Error(t, err)
	assert.Nil(t, buf)
}

func TestConnectEncodeDecodeOK(t *testing.T) {
	c := message.NewConnect()
	c.ClientId = "gqtt-example"
	c.ProtocolName = "MQTT"
	c.ProtocolVersion = 5
	c.CleanStart = true
	c.KeepAlive = 30
	buf, err := c.Encode()
	assert.NoError(t, err)

	f, p, err := message.ReceiveFrame(bytes.NewReader(buf))
	assert.NoError(t, err)
	assert.Exactly(t, f.Type, message.CONNECT)
	assert.Exactly(t, f.DUP, false)
	assert.Equal(t, f.QoS, message.QoS0)
	assert.Exactly(t, f.RETAIN, false)
	assert.Equal(t, f.Size, uint64(len(p)))

	c, err = message.ParseConnect(f, p)
	assert.NoError(t, err)
	assert.Equal(t, c.ProtocolName, "MQTT")
	assert.Equal(t, c.ProtocolVersion, uint8(5))
	assert.Equal(t, c.FlagUsername, false)
	assert.Equal(t, c.FlagPassword, false)
	assert.Equal(t, c.WillRetain, false)
	assert.Equal(t, c.WillQoS, message.QoS0)
	assert.Equal(t, c.FlagWill, false)
	assert.Equal(t, c.CleanStart, true)
	assert.Equal(t, c.KeepAlive, uint16(30))
	assert.Equal(t, c.ClientId, "gqtt-example")
	assert.Nil(t, c.Property)
	assert.Nil(t, c.WillProperty)
	assert.Empty(t, c.WillTopic)
	assert.Empty(t, c.WillPayload)
	assert.Empty(t, c.Username)
	assert.Empty(t, c.Password)
}

func TestUsingPropertyOnHeader(t *testing.T) {
	c := message.NewConnect()
	c.ClientId = "gqtt-example"
	c.ProtocolName = "MQTT"
	c.ProtocolVersion = 5
	c.CleanStart = true
	c.KeepAlive = 30

	c.Property = &message.ConnectProperty{
		SessionExpiryInterval:      1000,
		AuthenticationMethod:       "BASIC",
		AuthenticationData:         []byte("user:password"),
		RequestProblemInformation:  true,
		RequestResponseInformation: true,
		ReceiveMaximum:             100,
		TopicAliasMaximum:          100,
		UserProperty: map[string]string{
			"foo": "bar",
		},
		MaximumPacketSize: 65535,
	}
	buf, err := c.Encode()
	assert.NoError(t, err)

	f, p, err := message.ReceiveFrame(bytes.NewReader(buf))
	assert.NoError(t, err)
	assert.Exactly(t, f.Type, message.CONNECT)
	assert.Exactly(t, f.DUP, false)
	assert.Equal(t, f.QoS, message.QoS0)
	assert.Exactly(t, f.RETAIN, false)
	assert.Equal(t, f.Size, uint64(len(p)))

	c, err = message.ParseConnect(f, p)
	assert.NoError(t, err)
	assert.Equal(t, c.ProtocolName, "MQTT")
	assert.Equal(t, c.ProtocolVersion, uint8(5))
	assert.Equal(t, c.FlagUsername, false)
	assert.Equal(t, c.FlagPassword, false)
	assert.Equal(t, c.WillRetain, false)
	assert.Equal(t, c.WillQoS, message.QoS0)
	assert.Equal(t, c.FlagWill, false)
	assert.Equal(t, c.CleanStart, true)
	assert.Equal(t, c.KeepAlive, uint16(30))
	assert.Equal(t, c.ClientId, "gqtt-example")
	assert.NotNil(t, c.Property)
	assert.Equal(t, uint32(1000), c.Property.SessionExpiryInterval)
	assert.Equal(t, "BASIC", c.Property.AuthenticationMethod)
	assert.Equal(t, []byte("user:password"), c.Property.AuthenticationData)
	assert.True(t, c.Property.RequestProblemInformation)
	assert.True(t, c.Property.RequestResponseInformation)
	assert.Equal(t, uint16(100), c.Property.ReceiveMaximum)
	assert.Equal(t, uint16(100), c.Property.TopicAliasMaximum)
	assert.NotNil(t, c.Property.UserProperty)
	u := c.Property.UserProperty
	assert.Contains(t, u, "foo")
	assert.Equal(t, u["foo"], "bar")
	assert.Equal(t, uint32(65535), c.Property.MaximumPacketSize)
}

func TestUsingWillPropertyOnHeader(t *testing.T) {
	c := message.NewConnect()
	c.ClientId = "gqtt-example"
	c.ProtocolName = "MQTT"
	c.ProtocolVersion = 5
	c.CleanStart = true
	c.KeepAlive = 30

	c.WillProperty = &message.WillProperty{
		PayloadFormatIndicator: 1,
		MessageExpiryInterval:  1000,
		ContentType:            "text/plain",
		ResponseTopic:          "/response/will",
		CorrelationData:        []byte("correlationdata"),
		WillDelayInterval:      100,
		UserProperty: map[string]string{
			"will": "be",
		},
	}
	buf, err := c.Encode()
	assert.NoError(t, err)

	f, p, err := message.ReceiveFrame(bytes.NewReader(buf))
	assert.NoError(t, err)
	assert.Exactly(t, f.Type, message.CONNECT)
	assert.Exactly(t, f.DUP, false)
	assert.Equal(t, f.QoS, message.QoS0)
	assert.Exactly(t, f.RETAIN, false)
	assert.Equal(t, f.Size, uint64(len(p)))

	c, err = message.ParseConnect(f, p)
	assert.NoError(t, err)
	assert.Equal(t, c.ProtocolName, "MQTT")
	assert.Equal(t, c.ProtocolVersion, uint8(5))
	assert.Equal(t, c.FlagUsername, false)
	assert.Equal(t, c.FlagPassword, false)
	assert.Equal(t, c.WillRetain, false)
	assert.Equal(t, c.WillQoS, message.QoS0)
	assert.Equal(t, c.FlagWill, false)
	assert.Equal(t, c.CleanStart, true)
	assert.Equal(t, c.KeepAlive, uint16(30))
	assert.Equal(t, c.ClientId, "gqtt-example")
	assert.NotNil(t, c.WillProperty)
	assert.Equal(t, uint8(1), c.WillProperty.PayloadFormatIndicator)
	assert.Equal(t, uint32(1000), c.WillProperty.MessageExpiryInterval)
	assert.Equal(t, "text/plain", c.WillProperty.ContentType)
	assert.Equal(t, "/response/will", c.WillProperty.ResponseTopic)
	assert.Equal(t, []byte("correlationdata"), c.WillProperty.CorrelationData)
	assert.Equal(t, uint32(100), c.WillProperty.WillDelayInterval)
	assert.NotNil(t, c.WillProperty.UserProperty)
	u := c.WillProperty.UserProperty
	assert.Contains(t, u, "will")
	assert.Equal(t, u["will"], "be")
}

func TestForAuth(t *testing.T) {
	c := message.NewConnect()
	c.ClientId = "1111"
	c.Property = &message.ConnectProperty{
		AuthenticationMethod: "basic",
		AuthenticationData:   []byte("foo:bar"),
	}
	buf, err := c.Encode()
	assert.NoError(t, err)

	f, p, err := message.ReceiveFrame(bytes.NewReader(buf))
	assert.NoError(t, err)
	_, err = message.ParseConnect(f, p)
	assert.NoError(t, err)

}
