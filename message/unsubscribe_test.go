package message_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ysugimoto/gqtt/message"
)

func TestUnsubscribeNGWhenTopicIsEmpty(t *testing.T) {
	u := message.NewUnsubscribe()
	u.PacketId = 1000
	buf, err := u.Encode()
	assert.Error(t, err)
	assert.Nil(t, buf)
}

func TestUnsubscribeNGWhenPacketIdIsEmpty(t *testing.T) {
	u := message.NewUnsubscribe()
	u.AddTopic("foo/bar", "foo/bar/baz")
	buf, err := u.Encode()
	assert.Error(t, err)
	assert.Nil(t, buf)
}

func TestUnsubscribeMessageEncodeDecodeOK(t *testing.T) {
	u := message.NewUnsubscribe()
	u.PacketId = 1000
	u.AddTopic("foo/bar", "foo/bar/baz")
	u.Property = &message.UnsubscribeProperty{
		UserProperty: map[string]string{
			"foo": "bar",
		},
	}
	buf, err := u.Encode()
	assert.NoError(t, err)

	f, p, err := message.ReceiveFrame(bytes.NewReader(buf))
	assert.NoError(t, err)
	assert.Exactly(t, message.UNSUBSCRIBE, f.Type)
	assert.False(t, f.DUP)
	assert.Equal(t, message.QoS0, f.QoS)
	assert.False(t, f.RETAIN)
	assert.Equal(t, uint64(len(p)), f.Size)

	u, err = message.ParseUnsubscribe(f, p)
	assert.NoError(t, err)
	assert.Equal(t, uint16(1000), u.PacketId)
	assert.Equal(t, 2, len(u.Topics))
	assert.Equal(t, "foo/bar", u.Topics[0])
	assert.Equal(t, "foo/bar/baz", u.Topics[1])
	assert.NotNil(t, u.Property)
	assert.Contains(t, u.Property.UserProperty, "foo")
	assert.Equal(t, "bar", u.Property.UserProperty["foo"])
}
