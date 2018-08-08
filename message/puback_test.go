package message_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ysugimoto/gqtt/message"
)

func TestPubAckMessageFailedIfPacketIdIsZero(t *testing.T) {
	ack := message.NewPubAck(10)
	_, err := ack.Encode()
	assert.NoError(t, err)
}

func TestPubAckEncodeDecodeOK(t *testing.T) {
	t.Run("Omit return code", func(t *testing.T) {
		ack := message.NewPubAck(60000)
		buf, err := ack.Encode()
		assert.NoError(t, err)

		f, p, err := message.ReceiveFrame(bytes.NewReader(buf))
		assert.NoError(t, err)
		assert.Exactly(t, message.PUBACK, f.Type)
		assert.False(t, f.DUP)
		assert.Equal(t, message.QoS0, f.QoS)
		assert.False(t, f.RETAIN)
		assert.Equal(t, uint64(len(p)), f.Size)

		ack, err = message.ParsePubAck(f, p)
		assert.NoError(t, err)
		assert.Equal(t, uint16(60000), ack.PacketId)
		assert.Equal(t, message.Success, ack.ReasonCode)
		assert.Nil(t, ack.Property)
	})
	t.Run("Omit variable property", func(t *testing.T) {
		ack := message.NewPubAck(60000)
		ack.ReasonCode = message.NoSubscriptionExisted
		buf, err := ack.Encode()
		assert.NoError(t, err)

		f, p, err := message.ReceiveFrame(bytes.NewReader(buf))
		assert.NoError(t, err)
		assert.Exactly(t, f.Type, message.PUBACK)
		assert.Exactly(t, f.DUP, false)
		assert.Equal(t, f.QoS, message.QoS0)
		assert.Exactly(t, f.RETAIN, false)
		assert.Equal(t, f.Size, uint64(len(p)))

		ack, err = message.ParsePubAck(f, p)
		assert.NoError(t, err)
		assert.Equal(t, uint16(60000), ack.PacketId)
		assert.Equal(t, message.NoSubscriptionExisted, ack.ReasonCode)
		assert.Nil(t, ack.Property)
	})
	t.Run("Full case", func(t *testing.T) {
		ack := message.NewPubAck(60000)
		ack.ReasonCode = message.Success
		ack.Property = &message.PubAckProperty{
			ReasonString: "well done",
			UserProperty: map[string]string{
				"foo": "bar",
			},
		}
		buf, err := ack.Encode()
		assert.NoError(t, err)

		f, p, err := message.ReceiveFrame(bytes.NewReader(buf))
		assert.NoError(t, err)
		assert.Exactly(t, f.Type, message.PUBACK)
		assert.Exactly(t, f.DUP, false)
		assert.Equal(t, f.QoS, message.QoS0)
		assert.Exactly(t, f.RETAIN, false)
		assert.Equal(t, f.Size, uint64(len(p)))

		ack, err = message.ParsePubAck(f, p)
		assert.NoError(t, err)
		assert.Equal(t, uint16(60000), ack.PacketId)
		assert.Equal(t, message.Success, ack.ReasonCode)
		assert.NotNil(t, ack.Property)
		assert.Equal(t, "well done", ack.Property.ReasonString)
		assert.NotNil(t, ack.Property.UserProperty)
		u := ack.Property.UserProperty
		assert.Contains(t, u, "foo")
		assert.Equal(t, "bar", u["foo"])
	})
}
