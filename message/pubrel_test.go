package message_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ysugimoto/gqtt/message"
)

func TestPubRelMessageFailedIfPacketIdIsZero(t *testing.T) {
	rel := message.NewPubRel()
	buf, err := rel.Encode()
	assert.Error(t, err)
	assert.Nil(t, buf)
}

func TestPubRelEncodeDecodeOK(t *testing.T) {
	t.Run("Omit return code", func(t *testing.T) {
		rel := message.NewPubRel()
		rel.PacketId = 60000
		buf, err := rel.Encode()
		assert.NoError(t, err)

		f, p, err := message.ReceiveFrame(bytes.NewReader(buf))
		assert.NoError(t, err)
		assert.Exactly(t, message.PUBREL, f.Type)
		assert.False(t, f.DUP)
		assert.Equal(t, message.QoS0, f.QoS)
		assert.False(t, f.RETAIN)
		assert.Equal(t, uint64(len(p)), f.Size)

		rel, err = message.ParsePubRel(f, p)
		assert.NoError(t, err)
		assert.Equal(t, uint16(60000), rel.PacketId)
		assert.Equal(t, message.Success, rel.ReasonCode)
		assert.Nil(t, rel.Property)
	})
	t.Run("Omit variable property", func(t *testing.T) {
		rel := message.NewPubRel()
		rel.PacketId = 60000
		rel.ReasonCode = message.NoSubscriptionExisted
		buf, err := rel.Encode()
		assert.NoError(t, err)

		f, p, err := message.ReceiveFrame(bytes.NewReader(buf))
		assert.NoError(t, err)
		assert.Exactly(t, f.Type, message.PUBREL)
		assert.Exactly(t, f.DUP, false)
		assert.Equal(t, f.QoS, message.QoS0)
		assert.Exactly(t, f.RETAIN, false)
		assert.Equal(t, f.Size, uint64(len(p)))

		rel, err = message.ParsePubRel(f, p)
		assert.NoError(t, err)
		assert.Equal(t, uint16(60000), rel.PacketId)
		assert.Equal(t, message.NoSubscriptionExisted, rel.ReasonCode)
		assert.Nil(t, rel.Property)
	})
	t.Run("Full case", func(t *testing.T) {
		rel := message.NewPubRel()
		rel.PacketId = 60000
		rel.ReasonCode = message.Success
		rel.Property = &message.PubRelProperty{
			ReasonString: "well done",
			UserProperty: map[string]string{
				"foo": "bar",
			},
		}
		buf, err := rel.Encode()
		assert.NoError(t, err)

		f, p, err := message.ReceiveFrame(bytes.NewReader(buf))
		assert.NoError(t, err)
		assert.Exactly(t, f.Type, message.PUBREL)
		assert.Exactly(t, f.DUP, false)
		assert.Equal(t, f.QoS, message.QoS0)
		assert.Exactly(t, f.RETAIN, false)
		assert.Equal(t, f.Size, uint64(len(p)))

		rel, err = message.ParsePubRel(f, p)
		assert.NoError(t, err)
		assert.Equal(t, uint16(60000), rel.PacketId)
		assert.Equal(t, message.Success, rel.ReasonCode)
		assert.NotNil(t, rel.Property)
		assert.Equal(t, "well done", rel.Property.ReasonString)
		assert.NotNil(t, rel.Property.UserProperty)
		u := rel.Property.UserProperty
		assert.Contains(t, u, "foo")
		assert.Equal(t, "bar", u["foo"])
	})
}
