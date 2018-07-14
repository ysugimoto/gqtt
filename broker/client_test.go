package broker_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ysugimoto/gqtt/broker"
)

func TestClientSendable(t *testing.T) {
	c := broker.NewClient(nil, context.Background(), "1111-1111-1111-1111")
	c.Subscribe("/path/to/topic")

	assert.True(t, c.Sendable("/path/to/topic"))
	assert.False(t, c.Sendable("/path/to/topic/sub"))
	assert.False(t, c.Sendable("/path/to/topic/+"))
	assert.True(t, c.Sendable("/path/to/topic/#"))
	assert.True(t, c.Sendable("/path/to/#"))
	assert.True(t, c.Sendable("/path/+"))
	assert.True(t, c.Sendable("/path/+/topic"))

	c.Unsubscribe("/path/to/topic")
	assert.False(t, c.Sendable("/path/to/topic"))
	assert.False(t, c.Sendable("/path/to/topic/sub"))
	assert.False(t, c.Sendable("/path/to/topic/+"))
	assert.False(t, c.Sendable("/path/to/topic/#"))
	assert.False(t, c.Sendable("/path/to/#"))
	assert.False(t, c.Sendable("/path/+"))
	assert.False(t, c.Sendable("/path/+/topic"))
}
