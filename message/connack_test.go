package message_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ysugimoto/gqtt/message"
)

func TestConnAckMessageEncodeDecode(t *testing.T) {
	c := message.NewConnAck(&message.Frame{
		Type: message.CONNACK,
	})
	buf, err := c.Encode()
	assert.NoError(t, err)

	f, p, err := message.ReceiveFrame(bytes.NewReader(buf))
	assert.NoError(t, err)
	assert.Exactly(t, f.Type, message.CONNACK)
	assert.Exactly(t, f.DUP, false)
	assert.Equal(t, f.QoS, uint8(0))
	assert.Exactly(t, f.RETAIN, false)
	assert.Equal(t, f.Size, uint64(2))
	assert.Equal(t, f.Size, uint64(len(p)))

	c, err = message.ParseConnAck(f, p)
	assert.NoError(t, err)
	assert.Equal(t, c.ReturnCode, message.ConnAckOK)
}
