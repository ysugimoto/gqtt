package message_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ysugimoto/gqtt/message"
)

func TestSubscribeNGWhenTopicIsEmpty(t *testing.T) {
	s := message.NewSubscribe()
	s.PacketId = 1000
	buf, err := s.Encode()
	assert.Error(t, err)
	assert.Nil(t, buf)
}

func TestSubscribeNGWhenPackageIdIsEmpty(t *testing.T) {
	s := message.NewSubscribe()
	s.AddTopic(message.SubscribeTopic{
		TopicName: "foo/bar",
		QoS:       message.QoS1,
	})
	buf, err := s.Encode()
	assert.Error(t, err)
	assert.Nil(t, buf)
}

func TestSubscribeMessageEncodeDecodeOK(t *testing.T) {
	t.Run("Single subscribe", func(t *testing.T) {
		s := message.NewSubscribe()
		s.PacketId = 1000
		s.AddTopic(message.SubscribeTopic{
			TopicName: "foo/bar",
			QoS:       message.QoS1,
		})
		buf, err := s.Encode()
		assert.NoError(t, err)

		f, p, err := message.ReceiveFrame(bytes.NewReader(buf))
		assert.NoError(t, err)
		assert.Exactly(t, message.SUBSCRIBE, f.Type)
		assert.False(t, f.DUP)
		assert.Equal(t, message.QoS0, f.QoS)
		assert.False(t, f.RETAIN)
		assert.Equal(t, f.Size, uint64(len(p)))

		s, err = message.ParseSubscribe(f, p)
		assert.NoError(t, err)
		assert.Equal(t, uint16(1000), s.PacketId)
		assert.Nil(t, s.Property)
		assert.Equal(t, 1, len(s.Subscriptions))
		st := s.Subscriptions[0]
		assert.Equal(t, "foo/bar", st.TopicName)
		assert.Equal(t, message.QoS1, st.QoS)
		assert.Equal(t, uint8(0), st.RetainHandling)
		assert.False(t, st.RAP)
		assert.False(t, st.NoLocal)
	})
	t.Run("Multiple subscribe", func(t *testing.T) {
		s := message.NewSubscribe()
		s.PacketId = 1000
		s.AddTopic(
			message.SubscribeTopic{
				TopicName: "foo/bar",
				QoS:       message.QoS1,
			},
			message.SubscribeTopic{
				TopicName:      "lorem/ipsum",
				RetainHandling: 2,
				NoLocal:        true,
				QoS:            message.QoS0,
			},
		)
		buf, err := s.Encode()
		assert.NoError(t, err)

		f, p, err := message.ReceiveFrame(bytes.NewReader(buf))
		assert.NoError(t, err)
		assert.Exactly(t, message.SUBSCRIBE, f.Type)
		assert.False(t, f.DUP)
		assert.Equal(t, message.QoS0, f.QoS)
		assert.False(t, f.RETAIN)
		assert.Equal(t, f.Size, uint64(len(p)))

		s, err = message.ParseSubscribe(f, p)
		assert.NoError(t, err)
		assert.Equal(t, uint16(1000), s.PacketId)
		assert.Nil(t, s.Property)
		assert.Equal(t, 2, len(s.Subscriptions))
		st1 := s.Subscriptions[0]
		assert.Equal(t, "foo/bar", st1.TopicName)
		assert.Equal(t, message.QoS1, st1.QoS)
		assert.Equal(t, uint8(0), st1.RetainHandling)
		assert.False(t, st1.RAP)
		assert.False(t, st1.NoLocal)

		st2 := s.Subscriptions[1]
		assert.Equal(t, "lorem/ipsum", st2.TopicName)
		assert.Equal(t, message.QoS0, st2.QoS)
		assert.Equal(t, uint8(2), st2.RetainHandling)
		assert.False(t, st2.RAP)
		assert.True(t, st2.NoLocal)
	})
	t.Run("Full case", func(t *testing.T) {
		s := message.NewSubscribe()
		s.PacketId = 1000
		s.AddTopic(
			message.SubscribeTopic{
				TopicName: "foo/bar",
				QoS:       message.QoS1,
			},
			message.SubscribeTopic{
				TopicName:      "lorem/ipsum",
				RetainHandling: 2,
				NoLocal:        true,
				QoS:            message.QoS0,
			},
		)
		s.Property = &message.SubscribeProperty{
			SubscriptionIdentifier: 1234567890,
			UserProperty: map[string]string{
				"foo": "bar",
			},
		}
		buf, err := s.Encode()
		assert.NoError(t, err)

		f, p, err := message.ReceiveFrame(bytes.NewReader(buf))
		assert.NoError(t, err)
		assert.Exactly(t, message.SUBSCRIBE, f.Type)
		assert.False(t, f.DUP)
		assert.Equal(t, message.QoS0, f.QoS)
		assert.False(t, f.RETAIN)
		assert.Equal(t, f.Size, uint64(len(p)))

		s, err = message.ParseSubscribe(f, p)
		assert.NoError(t, err)
		assert.Equal(t, uint16(1000), s.PacketId)
		assert.Equal(t, 2, len(s.Subscriptions))
		st1 := s.Subscriptions[0]
		assert.Equal(t, "foo/bar", st1.TopicName)
		assert.Equal(t, message.QoS1, st1.QoS)
		assert.Equal(t, uint8(0), st1.RetainHandling)
		assert.False(t, st1.RAP)
		assert.False(t, st1.NoLocal)

		st2 := s.Subscriptions[1]
		assert.Equal(t, "lorem/ipsum", st2.TopicName)
		assert.Equal(t, message.QoS0, st2.QoS)
		assert.Equal(t, uint8(2), st2.RetainHandling)
		assert.False(t, st2.RAP)
		assert.True(t, st2.NoLocal)

		assert.NotNil(t, s.Property)
		assert.Equal(t, uint64(1234567890), s.Property.SubscriptionIdentifier)
		assert.NotNil(t, s.Property.UserProperty)
		assert.Contains(t, s.Property.UserProperty, "foo")
		assert.Equal(t, "bar", s.Property.UserProperty["foo"])
	})
}
