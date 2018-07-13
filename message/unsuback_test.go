package message_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ysugimoto/gqtt/message"
)

func TestUnsubAckMessageFailedIfMessageIDIsZero(t *testing.T) {
	s := message.NewUnsubAck(&message.Frame{
		Type: message.UNSUBACK,
	})
	buf, err := s.Encode()
	assert.Error(t, err)
	assert.Nil(t, buf)
}

func TestUnsubAckMessageEncodeDecodeOK(t *testing.T) {
	s := message.NewUnsubAck(&message.Frame{
		Type: message.UNSUBACK,
	})
	s.MessageID = 1000
	buf, err := s.Encode()
	assert.NoError(t, err)

	f, p, err := message.ReceiveFrame(bytes.NewReader(buf))
	assert.NoError(t, err)
	assert.Exactly(t, f.Type, message.UNSUBACK)
	assert.Exactly(t, f.DUP, false)
	assert.Equal(t, f.QoS, uint8(0))
	assert.Exactly(t, f.RETAIN, false)
	assert.Equal(t, f.Size, uint64(len(p)))

	s, err = message.ParseUnsubAck(f, p)
	assert.NoError(t, err)
	assert.Equal(t, s.MessageID, uint16(1000))
}
