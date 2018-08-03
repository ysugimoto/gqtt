package message_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ysugimoto/gqtt/message"
)

func TestPubAckMessageFailedIfMessageIDIsZero(t *testing.T) {
	a := message.NewPubAck(&message.Frame{
		Type: message.PUBACK,
	})
	buf, err := a.Encode()
	assert.Error(t, err)
	assert.Nil(t, buf)
}

func TestPubAckMessageEncodeDecodeOK(t *testing.T) {
	a := message.NewPubAck(&message.Frame{
		Type: message.PUBACK,
	})
	a.MessageID = uint16(1000)
	buf, err := a.Encode()
	assert.NoError(t, err)

	f, p, err := message.ReceiveFrame(bytes.NewReader(buf))
	assert.NoError(t, err)
	assert.Exactly(t, f.Type, message.PUBACK)
	assert.Exactly(t, f.DUP, false)
	assert.Equal(t, f.QoS, uint8(0))
	assert.Exactly(t, f.RETAIN, false)
	assert.Equal(t, f.Size, uint64(len(p)))

	a, err = message.ParsePubAck(f, p)
	assert.NoError(t, err)
	assert.Equal(t, a.MessageID, uint16(1000))
}
