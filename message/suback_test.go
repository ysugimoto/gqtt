package message_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ysugimoto/gqtt/message"
)

func TestSubAckMessageFailedIfMessageIDIsZero(t *testing.T) {
	s := message.NewSubAck(&message.Frame{
		Type: message.SUBACK,
	})
	buf, err := s.Encode()
	assert.Error(t, err)
	assert.Nil(t, buf)
}

func TestSubAckMessageFailedIfQoSsIsZero(t *testing.T) {
	s := message.NewSubAck(&message.Frame{
		Type: message.SUBACK,
	})
	s.MessageID = 1000
	buf, err := s.Encode()
	assert.Error(t, err)
	assert.Nil(t, buf)
}

func TestSubAckMessageEncodeDecodeOK(t *testing.T) {
	s := message.NewSubAck(&message.Frame{
		Type: message.SUBACK,
	})
	s.MessageID = 1000
	s.AddQoS(message.QoS0, message.QoS1)
	buf, err := s.Encode()
	assert.NoError(t, err)

	f, p, err := message.ReceiveFrame(bytes.NewReader(buf))
	assert.NoError(t, err)
	assert.Exactly(t, f.Type, message.SUBACK)
	assert.Exactly(t, f.DUP, false)
	assert.Equal(t, f.QoS, uint8(0))
	assert.Exactly(t, f.RETAIN, false)
	assert.Equal(t, f.Size, uint64(len(p)))

	s, err = message.ParseSubAck(f, p)
	assert.NoError(t, err)
	assert.Equal(t, s.MessageID, uint16(1000))
	assert.Equal(t, len(s.QoSs), 2)
	assert.Equal(t, s.QoSs[0], message.QoS0)
	assert.Equal(t, s.QoSs[1], message.QoS1)
}
