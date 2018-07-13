package message_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ysugimoto/gqtt/message"
)

func TestUnsubscribeNGWhenTopicIsEmpty(t *testing.T) {
	s := message.NewUnsubscribe(&message.Frame{
		Type: message.UNSUBSCRIBE,
	})
	s.MessageID = 1000
	buf, err := s.Encode()
	assert.Error(t, err)
	assert.Nil(t, buf)
}

func TestUnsubscribeNGWhenMessageIDIsEmpty(t *testing.T) {
	s := message.NewUnsubscribe(&message.Frame{
		Type: message.UNSUBSCRIBE,
	})
	s.AddTopic("foo/bar", "foo/bar/baz")
	buf, err := s.Encode()
	assert.Error(t, err)
	assert.Nil(t, buf)
}

func TestUnsubscribeMessageEncodeDecodeOK(t *testing.T) {
	s := message.NewUnsubscribe(&message.Frame{
		Type: message.UNSUBSCRIBE,
	})
	s.MessageID = 1000
	s.AddTopic("foo/bar", "foo/bar/baz")
	buf, err := s.Encode()
	assert.NoError(t, err)

	f, p, err := message.ReceiveFrame(bytes.NewReader(buf))
	assert.NoError(t, err)
	assert.Exactly(t, f.Type, message.UNSUBSCRIBE)
	assert.Exactly(t, f.DUP, false)
	assert.Equal(t, f.QoS, uint8(0))
	assert.Exactly(t, f.RETAIN, false)
	assert.Equal(t, f.Size, uint64(len(p)))

	s, err = message.ParseUnsubscribe(f, p)
	assert.NoError(t, err)
	assert.Equal(t, s.MessageID, uint16(1000))
	assert.Equal(t, len(s.Topics), 2)
	assert.Equal(t, s.Topics[0], "foo/bar")
	assert.Equal(t, s.Topics[1], "foo/bar/baz")
}
