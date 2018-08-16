package broker_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ysugimoto/gqtt/broker"
	"github.com/ysugimoto/gqtt/message"
)

func TestSubscribeQoS0(t *testing.T) {
	ss := broker.NewSubscription()
	reason, err := ss.Subscribe("1111-1111-1111-1111", message.SubscribeTopic{
		TopicName: "foo/bar",
		QoS:       message.QoS0,
	})
	assert.NoError(t, err)
	assert.Equal(t, message.GrantedQoS0, reason)
}

func TestSubscribeQoS1(t *testing.T) {
	ss := broker.NewSubscription()
	reason, err := ss.Subscribe("1111-1111-1111-1111", message.SubscribeTopic{
		TopicName: "foo/bar",
		QoS:       message.QoS1,
	})
	assert.NoError(t, err)
	assert.Equal(t, message.GrantedQoS1, reason)
}

func TestSubscribeQoS2(t *testing.T) {
	ss := broker.NewSubscription()
	reason, err := ss.Subscribe("1111-1111-1111-1111", message.SubscribeTopic{
		TopicName: "foo/bar",
		QoS:       message.QoS2,
	})
	assert.NoError(t, err)
	assert.Equal(t, message.GrantedQoS2, reason)
}

func TestSubscribeErrorIfInvalidMultiLevelWildcardContains(t *testing.T) {
	ss := broker.NewSubscription()
	reason, err := ss.Subscribe("1111-1111-1111-1111", message.SubscribeTopic{
		TopicName: "foo/bar#",
		QoS:       message.QoS0,
	})
	assert.Error(t, err)
	assert.Equal(t, message.UnspecifiedError, reason)

	reason, err = ss.Subscribe("1111-1111-1111-1111", message.SubscribeTopic{
		TopicName: "foo/bar#/baz",
		QoS:       message.QoS0,
	})
	assert.Error(t, err)
	assert.Equal(t, message.UnspecifiedError, reason)
}

func TestSubscribeWithMultiLevelWildcard(t *testing.T) {
	ss := broker.NewSubscription()
	reason, err := ss.Subscribe("1111-1111-1111-1111", message.SubscribeTopic{
		TopicName: "#",
		QoS:       message.QoS0,
	})
	assert.NoError(t, err)
	assert.Equal(t, message.NoSubscriptionExisted, reason)

	reason, err = ss.Subscribe("1111-1111-1111-1111", message.SubscribeTopic{
		TopicName: "foo/#",
		QoS:       message.QoS0,
	})
	assert.NoError(t, err)
	assert.Equal(t, message.NoSubscriptionExisted, reason)
}

func TestSubscribeWithSingleLevelWildcard(t *testing.T) {
	ss := broker.NewSubscription()
	reason, err := ss.Subscribe("1111-1111-1111-1111", message.SubscribeTopic{
		TopicName: "+",
		QoS:       message.QoS0,
	})
	assert.NoError(t, err)
	assert.Equal(t, message.NoSubscriptionExisted, reason)

	reason, err = ss.Subscribe("1111-1111-1111-1111", message.SubscribeTopic{
		TopicName: "foo/+",
		QoS:       message.QoS0,
	})
	assert.NoError(t, err)
	assert.Equal(t, message.NoSubscriptionExisted, reason)

	reason, err = ss.Subscribe("1111-1111-1111-1111", message.SubscribeTopic{
		TopicName: "foo/+/bar",
		QoS:       message.QoS0,
	})
	assert.NoError(t, err)
	assert.Equal(t, message.NoSubscriptionExisted, reason)
}

func TestSubscribeErrorIfInvalidSingleLevelWildcardContains(t *testing.T) {
	ss := broker.NewSubscription()
	reason, err := ss.Subscribe("1111-1111-1111-1111", message.SubscribeTopic{
		TopicName: "foo+",
		QoS:       message.QoS0,
	})
	assert.Error(t, err)
	assert.Equal(t, message.UnspecifiedError, reason)
}

func TestFindTopicsWithExactTopic(t *testing.T) {
	ss := broker.NewSubscription()
	ss.Subscribe("1111-1111-1111-1111", message.SubscribeTopic{
		TopicName: "foo/bar",
		QoS:       message.QoS0,
	})
	ss.Subscribe("1111-1111-1111-1111", message.SubscribeTopic{
		TopicName: "foo/baz",
		QoS:       message.QoS0,
	})
	topics, err := ss.FindTopics("foo/bar")
	assert.NoError(t, err)
	assert.Equal(t, 1, len(topics))
	assert.Equal(t, "foo/bar", topics[0])
}

func TestFindTopicsWithExactTopicMultiLevelWildcard(t *testing.T) {
	ss := broker.NewSubscription()
	ss.Subscribe("1111-1111-1111-1111", message.SubscribeTopic{
		TopicName: "hoge",
		QoS:       message.QoS0,
	})
	ss.Subscribe("1111-1111-1111-1111", message.SubscribeTopic{
		TopicName: "foo",
		QoS:       message.QoS0,
	})
	ss.Subscribe("1111-1111-1111-1111", message.SubscribeTopic{
		TopicName: "foo/bar",
		QoS:       message.QoS0,
	})
	ss.Subscribe("1111-1111-1111-1111", message.SubscribeTopic{
		TopicName: "foo/bar/baz",
		QoS:       message.QoS0,
	})
	topics, err := ss.FindTopics("foo/#")
	assert.NoError(t, err)
	assert.Equal(t, 3, len(topics))
	// unordered
}

func TestFindTopicsWithExactTopicSingleLevelWildcard(t *testing.T) {
	ss := broker.NewSubscription()
	ss.Subscribe("1111-1111-1111-1111", message.SubscribeTopic{
		TopicName: "hoge",
		QoS:       message.QoS0,
	})
	ss.Subscribe("1111-1111-1111-1111", message.SubscribeTopic{
		TopicName: "foo",
		QoS:       message.QoS0,
	})
	ss.Subscribe("1111-1111-1111-1111", message.SubscribeTopic{
		TopicName: "foo/bar",
		QoS:       message.QoS0,
	})
	ss.Subscribe("1111-1111-1111-1111", message.SubscribeTopic{
		TopicName: "foo/bar/baz",
		QoS:       message.QoS0,
	})
	ss.Subscribe("1111-1111-1111-1111", message.SubscribeTopic{
		TopicName: "foo/piyo/baz",
		QoS:       message.QoS0,
	})
	topics, err := ss.FindTopics("foo/+")
	assert.NoError(t, err)
	assert.Equal(t, 1, len(topics))
	assert.Equal(t, "foo/bar", topics[0])

	topics, err = ss.FindTopics("foo/+/baz")
	assert.NoError(t, err)
	assert.Equal(t, 2, len(topics))
	// unordered
}
