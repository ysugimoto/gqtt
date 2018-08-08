package message_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ysugimoto/gqtt/message"
)

/*
func TestPublishQoS0MessageEncodeNGWhenTopicNameIsEmpty(t *testing.T) {
	pb := message.NewPublish(0)
	pb.Body = []byte("gqtt-body")
	_, err := pb.Encode()
	assert.Error(t, err)
}

func TestPublishQoS0MessageEncodeDecodeOK(t *testing.T) {
	pb := message.NewPublish(0)
	pb.TopicName = "foo/bar"
	pb.Body = []byte("gqtt-body")
	pb.Property = &message.PublishProperty{
		PayloadFormatIndicator: 1,
		MessageExpiryInterval:  100,
		ContentType:            "text/plain",
		ResponseTopic:          "/resp/topic",
		CorrelationData:        []byte("somedata"),
		SubscriptionIdentifier: 1234567890,
		TopicAlias:             10,
		UserProperty: map[string]string{
			"foo": "bar",
		},
	}
	buf, err := pb.Encode()
	assert.NoError(t, err)

	f, p, err := message.ReceiveFrame(bytes.NewReader(buf))
	assert.NoError(t, err)
	assert.Exactly(t, f.Type, message.PUBLISH)
	assert.Exactly(t, f.DUP, false)
	assert.Equal(t, f.QoS, message.QoS0)
	assert.Exactly(t, f.RETAIN, false)
	assert.Equal(t, f.Size, uint64(len(p)))

	pb, err = message.ParsePublish(f, p)
	assert.NoError(t, err)
	assert.Equal(t, "foo/bar", pb.TopicName)
	assert.Equal(t, []byte("gqtt-body"), pb.Body)
	assert.Equal(t, uint16(0), pb.PacketId)
	assert.NotNil(t, pb.Property)
	prop := pb.Property
	assert.Equal(t, uint8(1), prop.PayloadFormatIndicator)
	assert.Equal(t, uint32(100), prop.MessageExpiryInterval)
	assert.Equal(t, "text/plain", prop.ContentType)
	assert.Equal(t, "/resp/topic", prop.ResponseTopic)
	assert.Equal(t, []byte("somedata"), prop.CorrelationData)
	assert.Equal(t, uint64(1234567890), prop.SubscriptionIdentifier)
	assert.Equal(t, uint16(10), prop.TopicAlias)
	assert.Equal(t, uint8(1), prop.PayloadFormatIndicator)
	assert.NotNil(t, prop.UserProperty)
	assert.Contains(t, prop.UserProperty, "foo")
	assert.Equal(t, prop.UserProperty["foo"], "bar")
}

func TestPublishQoS1MessageErrorIfPacketIdIsZero(t *testing.T) {
	pb := message.NewPublish(0)
	pb.TopicName = "foo/bar"
	pb.Body = []byte("gqtt-body")
	pb.QoS = message.QoS1
	buf, err := pb.Encode()
	assert.Error(t, err)
	assert.Nil(t, buf)
}

*/
func TestDecodePublishFromRawBytes(t *testing.T) {
	pb := message.NewPublish(0, message.WithQoS(message.QoS0))
	pb.TopicName = "gqtt/example"
	buf, err := pb.Encode()
	assert.NoError(t, err)
	// raw := []byte{48, 15, 0, 2, 103, 113, 116, 116, 47, 101, 120, 97, 109, 112, 108, 101, 0}
	// assert.Equal(t, raw, buf)
	f, p, err := message.ReceiveFrame(bytes.NewReader(buf))
	assert.NoError(t, err)
	_, err = message.ParsePublish(f, p)
	assert.NoError(t, err)
}
