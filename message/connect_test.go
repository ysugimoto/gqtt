package message_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ysugimoto/gqtt/message"
)

func TestConnectEncodeErrorIfClientIDIsEmpty(t *testing.T) {
	c := message.NewConnect(&message.Frame{
		Type: message.CONNECT,
	})
	buf, err := c.Encode()
	assert.Error(t, err)
	assert.Nil(t, buf)
}

func TestConnectEncodeDecodeOK(t *testing.T) {
	c := message.NewConnect(&message.Frame{
		Type: message.CONNECT,
	})
	c.ClientID = "gqtt-example"
	c.ProtocolName = "MQIsdp"
	c.ProtocolVersion = 3
	c.CleanSession = true
	c.KeepAlive = 30
	buf, err := c.Encode()
	assert.NoError(t, err)

	f, p, err := message.ReceiveFrame(bytes.NewReader(buf))
	assert.NoError(t, err)
	assert.Exactly(t, f.Type, message.CONNECT)
	assert.Exactly(t, f.DUP, false)
	assert.Equal(t, f.QoS, uint8(0))
	assert.Exactly(t, f.RETAIN, false)
	assert.Equal(t, f.Size, uint64(len(p)))

	c, err = message.ParseConnect(f, p)
	assert.NoError(t, err)
	assert.Equal(t, c.ProtocolName, "MQIsdp")
	assert.Equal(t, c.ProtocolVersion, uint8(3))
	assert.Equal(t, c.FlagUsername, false)
	assert.Equal(t, c.FlagPassword, false)
	assert.Equal(t, c.WillRetain, false)
	assert.Equal(t, c.WillQoS, uint8(0))
	assert.Equal(t, c.FlagWill, false)
	assert.Equal(t, c.CleanSession, true)
	assert.Equal(t, c.KeepAlive, uint16(30))
	assert.Equal(t, c.ClientID, "gqtt-example")
	assert.Empty(t, c.WillTopic)
	assert.Empty(t, c.WillMessage)
	assert.Empty(t, c.Username)
	assert.Empty(t, c.Password)
}
