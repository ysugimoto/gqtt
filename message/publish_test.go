package message_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ysugimoto/gqtt/message"
)

func TestPublishQoS0MessageEncodeNGWhenTopicNameIsEmpty(t *testing.T) {
	pb := message.NewPublish(&message.Frame{
		Type: message.PUBLISH,
	})
	pb.Body = "gqtt-body"
	buf, err := pb.Encode()
	assert.Error(t, err)
	assert.Nil(t, buf)
}

func TestPublishQoS0MessageEncodeDecodeOK(t *testing.T) {
	pb := message.NewPublish(&message.Frame{
		Type: message.PUBLISH,
	})
	pb.TopicName = "foo/bar"
	pb.Body = "gqtt-body"
	buf, err := pb.Encode()
	assert.NoError(t, err)

	f, p, err := message.ReceiveFrame(bytes.NewReader(buf))
	assert.NoError(t, err)
	assert.Exactly(t, f.Type, message.PUBLISH)
	assert.Exactly(t, f.DUP, false)
	assert.Equal(t, f.QoS, uint8(0))
	assert.Exactly(t, f.RETAIN, false)
	assert.Equal(t, f.Size, uint64(len(p)))

	pb, err = message.ParsePublish(f, p)
	assert.NoError(t, err)
	assert.Equal(t, pb.TopicName, "foo/bar")
	assert.Equal(t, pb.Body, "gqtt-body")
	assert.Empty(t, pb.MessageID)
}

func TestPublishQoS1MessageErrorIfMessaggeIDIsEmpty(t *testing.T) {
	pb := message.NewPublish(&message.Frame{
		Type: message.PUBLISH,
		QoS:  1,
	})
	pb.TopicName = "foo/bar"
	pb.Body = "gqtt-body"
	buf, err := pb.Encode()
	assert.Error(t, err)
	assert.Nil(t, buf)
}

func TestPublishQoS1MessageEncodeDecodeOK(t *testing.T) {
	pb := message.NewPublish(&message.Frame{
		Type: message.PUBLISH,
		QoS:  1,
	})
	pb.MessageID = 11111
	pb.TopicName = "foo/bar"
	pb.Body = "gqtt-body"
	buf, err := pb.Encode()
	assert.NoError(t, err)

	f, p, err := message.ReceiveFrame(bytes.NewReader(buf))
	assert.NoError(t, err)
	assert.Exactly(t, f.Type, message.PUBLISH)
	assert.Exactly(t, f.DUP, false)
	assert.Equal(t, f.QoS, uint8(1))
	assert.Exactly(t, f.RETAIN, false)
	assert.Equal(t, f.Size, uint64(len(p)))

	pb, err = message.ParsePublish(f, p)
	assert.NoError(t, err)
	assert.Equal(t, pb.TopicName, "foo/bar")
	assert.Equal(t, pb.Body, "gqtt-body")
	assert.Equal(t, pb.MessageID, uint16(11111))
}
