package message_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ysugimoto/gqtt/message"
)

func TestPubCompMessageFailedIfPacketIdIsZero(t *testing.T) {
	comp := message.NewPubComp(0)
	buf, err := comp.Encode()
	assert.Error(t, err)
	assert.Nil(t, buf)
}

func TestPubCompEncodeDecodeOK(t *testing.T) {
	t.Run("Omit return code", func(t *testing.T) {
		comp := message.NewPubComp(60000)
		buf, err := comp.Encode()
		assert.NoError(t, err)

		f, p, err := message.ReceiveFrame(bytes.NewReader(buf))
		assert.NoError(t, err)
		assert.Exactly(t, message.PUBCOMP, f.Type)
		assert.False(t, f.DUP)
		assert.Equal(t, message.QoS0, f.QoS)
		assert.False(t, f.RETAIN)
		assert.Equal(t, uint64(len(p)), f.Size)

		comp, err = message.ParsePubComp(f, p)
		assert.NoError(t, err)
		assert.Equal(t, uint16(60000), comp.PacketId)
		assert.Equal(t, message.Success, comp.ReasonCode)
		assert.Nil(t, comp.Property)
	})
	t.Run("Omit variable property", func(t *testing.T) {
		comp := message.NewPubComp(60000)
		comp.ReasonCode = message.NoSubscriptionExisted
		buf, err := comp.Encode()
		assert.NoError(t, err)

		f, p, err := message.ReceiveFrame(bytes.NewReader(buf))
		assert.NoError(t, err)
		assert.Exactly(t, f.Type, message.PUBCOMP)
		assert.Exactly(t, f.DUP, false)
		assert.Equal(t, f.QoS, message.QoS0)
		assert.Exactly(t, f.RETAIN, false)
		assert.Equal(t, f.Size, uint64(len(p)))

		comp, err = message.ParsePubComp(f, p)
		assert.NoError(t, err)
		assert.Equal(t, uint16(60000), comp.PacketId)
		assert.Equal(t, message.NoSubscriptionExisted, comp.ReasonCode)
		assert.Nil(t, comp.Property)
	})
	t.Run("Full case", func(t *testing.T) {
		comp := message.NewPubComp(60000)
		comp.ReasonCode = message.Success
		comp.Property = &message.PubCompProperty{
			ReasonString: "well done",
			UserProperty: map[string]string{
				"foo": "bar",
			},
		}
		buf, err := comp.Encode()
		assert.NoError(t, err)

		f, p, err := message.ReceiveFrame(bytes.NewReader(buf))
		assert.NoError(t, err)
		assert.Exactly(t, f.Type, message.PUBCOMP)
		assert.Exactly(t, f.DUP, false)
		assert.Equal(t, f.QoS, message.QoS0)
		assert.Exactly(t, f.RETAIN, false)
		assert.Equal(t, f.Size, uint64(len(p)))

		comp, err = message.ParsePubComp(f, p)
		assert.NoError(t, err)
		assert.Equal(t, uint16(60000), comp.PacketId)
		assert.Equal(t, message.Success, comp.ReasonCode)
		assert.NotNil(t, comp.Property)
		assert.Equal(t, "well done", comp.Property.ReasonString)
		assert.NotNil(t, comp.Property.UserProperty)
		u := comp.Property.UserProperty
		assert.Contains(t, u, "foo")
		assert.Equal(t, "bar", u["foo"])
	})
}
