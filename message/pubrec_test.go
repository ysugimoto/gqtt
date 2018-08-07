package message_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ysugimoto/gqtt/message"
)

func TestPubRecMessageFailedIfPacketIdIsZero(t *testing.T) {
	rec := message.NewPubRec()
	buf, err := rec.Encode()
	assert.Error(t, err)
	assert.Nil(t, buf)
}

func TestPubRecEncodeDecodeOK(t *testing.T) {
	t.Run("Omit return code", func(t *testing.T) {
		rec := message.NewPubRec()
		rec.PacketId = 60000
		buf, err := rec.Encode()
		assert.NoError(t, err)

		f, p, err := message.ReceiveFrame(bytes.NewReader(buf))
		assert.NoError(t, err)
		assert.Exactly(t, message.PUBREC, f.Type)
		assert.False(t, f.DUP)
		assert.Equal(t, message.QoS0, f.QoS)
		assert.False(t, f.RETAIN)
		assert.Equal(t, uint64(len(p)), f.Size)

		rec, err = message.ParsePubRec(f, p)
		assert.NoError(t, err)
		assert.Equal(t, uint16(60000), rec.PacketId)
		assert.Equal(t, message.Success, rec.ReasonCode)
		assert.Nil(t, rec.Property)
	})
	t.Run("Omit variable property", func(t *testing.T) {
		rec := message.NewPubRec()
		rec.PacketId = 60000
		rec.ReasonCode = message.NoSubscriptionExisted
		buf, err := rec.Encode()
		assert.NoError(t, err)

		f, p, err := message.ReceiveFrame(bytes.NewReader(buf))
		assert.NoError(t, err)
		assert.Exactly(t, f.Type, message.PUBREC)
		assert.Exactly(t, f.DUP, false)
		assert.Equal(t, f.QoS, message.QoS0)
		assert.Exactly(t, f.RETAIN, false)
		assert.Equal(t, f.Size, uint64(len(p)))

		rec, err = message.ParsePubRec(f, p)
		assert.NoError(t, err)
		assert.Equal(t, uint16(60000), rec.PacketId)
		assert.Equal(t, message.NoSubscriptionExisted, rec.ReasonCode)
		assert.Nil(t, rec.Property)
	})
	t.Run("Full case", func(t *testing.T) {
		rec := message.NewPubRec()
		rec.PacketId = 60000
		rec.ReasonCode = message.Success
		rec.Property = &message.PubRecProperty{
			ReasonString: "well done",
			UserProperty: map[string]string{
				"foo": "bar",
			},
		}
		buf, err := rec.Encode()
		assert.NoError(t, err)

		f, p, err := message.ReceiveFrame(bytes.NewReader(buf))
		assert.NoError(t, err)
		assert.Exactly(t, f.Type, message.PUBREC)
		assert.Exactly(t, f.DUP, false)
		assert.Equal(t, f.QoS, message.QoS0)
		assert.Exactly(t, f.RETAIN, false)
		assert.Equal(t, f.Size, uint64(len(p)))

		rec, err = message.ParsePubRec(f, p)
		assert.NoError(t, err)
		assert.Equal(t, uint16(60000), rec.PacketId)
		assert.Equal(t, message.Success, rec.ReasonCode)
		assert.NotNil(t, rec.Property)
		assert.Equal(t, "well done", rec.Property.ReasonString)
		assert.NotNil(t, rec.Property.UserProperty)
		u := rec.Property.UserProperty
		assert.Contains(t, u, "foo")
		assert.Equal(t, "bar", u["foo"])
	})
}
