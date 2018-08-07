package message_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ysugimoto/gqtt/message"
)

func TestUnsubAckMessageFailedIfPacketIdIsZero(t *testing.T) {
	u := message.NewUnsubAck(message.ServerBusy)
	buf, err := u.Encode()
	assert.Error(t, err)
	assert.Nil(t, buf)
}

func TestUnsubAckMessageFailedIfReasonCodesIsEmpty(t *testing.T) {
	u := message.NewUnsubAck()
	u.PacketId = 1000
	buf, err := u.Encode()
	assert.Error(t, err)
	assert.Nil(t, buf)
}

func TestUnsubAckMessageEncodeDecodeOK(t *testing.T) {
	u := message.NewUnsubAck()
	u.PacketId = 1000
	u.AddReasonCode(message.GrantedQoS0, message.GrantedQoS1, message.GrantedQoS2)
	u.Property = &message.UnsubAckProperty{
		ReasonString: "unsuback ok",
		UserProperty: map[string]string{
			"foo": "bar",
		},
	}
	buf, err := u.Encode()
	assert.NoError(t, err)

	f, p, err := message.ReceiveFrame(bytes.NewReader(buf))
	assert.NoError(t, err)
	assert.Exactly(t, message.UNSUBACK, f.Type)
	assert.False(t, f.DUP)
	assert.Equal(t, message.QoS0, f.QoS)
	assert.False(t, f.RETAIN)
	assert.Equal(t, uint64(len(p)), f.Size)

	u, err = message.ParseUnsubAck(f, p)
	assert.NoError(t, err)
	assert.Equal(t, uint16(1000), u.PacketId)
	assert.Equal(t, 3, len(u.ReasonCodes))
	assert.Equal(t, message.GrantedQoS0, u.ReasonCodes[0])
	assert.Equal(t, message.GrantedQoS1, u.ReasonCodes[1])
	assert.Equal(t, message.GrantedQoS2, u.ReasonCodes[2])
	assert.NotNil(t, u.Property)
	assert.Equal(t, "unsuback ok", u.Property.ReasonString)
	assert.Contains(t, u.Property.UserProperty, "foo")
	assert.Equal(t, "bar", u.Property.UserProperty["foo"])
}
