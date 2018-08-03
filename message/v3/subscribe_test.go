package message_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ysugimoto/gqtt/message"
)

func TestSubscribeNGWhenTopicIsEmpty(t *testing.T) {
	s := message.NewSubscribe(&message.Frame{
		Type: message.SUBSCRIBE,
	})
	s.MessageID = 1000
	buf, err := s.Encode()
	assert.Error(t, err)
	assert.Nil(t, buf)
}

func TestSubscribeNGWhenMessageIDIsEmpty(t *testing.T) {
	s := message.NewSubscribe(&message.Frame{
		Type: message.SUBSCRIBE,
	})
	s.AddTopic("foo/bar", message.QoS1)
	buf, err := s.Encode()
	assert.Error(t, err)
	assert.Nil(t, buf)
}

func TestSubscribeMessageEncodeDecodeOK(t *testing.T) {
	s := message.NewSubscribe(&message.Frame{
		Type: message.SUBSCRIBE,
	})
	s.MessageID = 1000
	s.AddTopic("foo/bar", message.QoS1)
	buf, err := s.Encode()
	assert.NoError(t, err)

	f, p, err := message.ReceiveFrame(bytes.NewReader(buf))
	assert.NoError(t, err)
	assert.Exactly(t, f.Type, message.SUBSCRIBE)
	assert.Exactly(t, f.DUP, false)
	assert.Equal(t, f.QoS, uint8(0))
	assert.Exactly(t, f.RETAIN, false)
	assert.Equal(t, f.Size, uint64(len(p)))

	s, err = message.ParseSubscribe(f, p)
	assert.NoError(t, err)
	assert.Equal(t, len(s.Topics), 1)
	topic := s.Topics[0]
	assert.Equal(t, topic.Name, "foo/bar")
	assert.Equal(t, topic.QoS, message.QoS1)
	assert.Equal(t, s.MessageID, uint16(1000))
}

func TestSubscribeMessageEncodeDecodeOKWithMultipleTopics(t *testing.T) {
	s := message.NewSubscribe(&message.Frame{
		Type: message.SUBSCRIBE,
		QoS:  message.QoS1,
	})
	s.MessageID = 1000

	s.AddTopic("foo/bar", message.QoS1)
	s.AddTopic("foo/bar/baz", message.QoS0)
	s.AddTopic("foo/bar/baz/qux", message.QoS2)
	buf, err := s.Encode()
	assert.NoError(t, err)

	f, p, err := message.ReceiveFrame(bytes.NewReader(buf))
	assert.NoError(t, err)
	assert.Exactly(t, f.Type, message.SUBSCRIBE)
	assert.Exactly(t, f.DUP, false)
	assert.Equal(t, f.QoS, uint8(1))
	assert.Exactly(t, f.RETAIN, false)
	assert.Equal(t, f.Size, uint64(len(p)))

	s, err = message.ParseSubscribe(f, p)
	assert.NoError(t, err)
	assert.Equal(t, len(s.Topics), 3)
	assert.Equal(t, s.MessageID, uint16(1000))

	assert.Equal(t, s.Topics[0].Name, "foo/bar")
	assert.Equal(t, s.Topics[0].QoS, message.QoS1)
	assert.Equal(t, s.Topics[1].Name, "foo/bar/baz")
	assert.Equal(t, s.Topics[1].QoS, message.QoS0)
	assert.Equal(t, s.Topics[2].Name, "foo/bar/baz/qux")
	assert.Equal(t, s.Topics[2].QoS, message.QoS2)

}
