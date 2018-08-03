package message_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ysugimoto/gqtt/message"
)

func TestDisconnectMessageEncodeDecode(t *testing.T) {
	c := message.NewDisconnect(&message.Frame{
		Type: message.DISCONNECT,
	})
	buf, err := c.Encode()
	assert.NoError(t, err)

	f, p, err := message.ReceiveFrame(bytes.NewReader(buf))
	assert.NoError(t, err)
	assert.Exactly(t, f.Type, message.DISCONNECT)
	assert.Exactly(t, f.DUP, false)
	assert.Equal(t, f.QoS, uint8(0))
	assert.Exactly(t, f.RETAIN, false)
	assert.Equal(t, f.Size, uint64(0))
	assert.Equal(t, f.Size, uint64(len(p)))
}
