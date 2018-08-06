package message_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ysugimoto/gqtt/message"
)

func TestSubAckMessageFailedIfPacketIdIsZero(t *testing.T) {
	s := message.NewSubAck()
	buf, err := s.Encode()
	assert.Error(t, err)
	assert.Nil(t, buf)
}

func TestSubAckMessageFailedIfReasonCodesIsEmpty(t *testing.T) {
	s := message.NewSubAck()
	s.PacketId = 1000
	buf, err := s.Encode()
	assert.Error(t, err)
	assert.Nil(t, buf)
}

func TestSubAckMessageEncodeDecodeOK(t *testing.T) {
	s := message.NewSubAck()
	s.PacketId = 1000
	s.AddReasonCode(message.GrantedQoS0, message.GrantedQoS1, message.GrantedQoS2)
	s.Property = &message.SubAckProperty{
		ReasonString: "suback ok",
		UserProperty: map[string]string{
			"foo": "bar",
		},
	}
	buf, err := s.Encode()
	assert.NoError(t, err)

	f, p, err := message.ReceiveFrame(bytes.NewReader(buf))
	assert.NoError(t, err)
	assert.Exactly(t, message.SUBACK, f.Type)
	assert.False(t, f.DUP)
	assert.Equal(t, message.QoS0, f.QoS)
	assert.False(t, f.RETAIN)
	assert.Equal(t, uint64(len(p)), f.Size)

	s, err = message.ParseSubAck(f, p)
	assert.NoError(t, err)
	assert.Equal(t, uint16(1000), s.PacketId)
	assert.Equal(t, 3, len(s.ReasonCodes))
	assert.Equal(t, message.GrantedQoS0, s.ReasonCodes[0])
	assert.Equal(t, message.GrantedQoS1, s.ReasonCodes[1])
	assert.Equal(t, message.GrantedQoS2, s.ReasonCodes[2])
	assert.NotNil(t, s.Property)
	assert.Equal(t, "suback ok", s.Property.ReasonString)
	assert.Contains(t, s.Property.UserProperty, "foo")
	assert.Equal(t, "bar", s.Property.UserProperty["foo"])
}
