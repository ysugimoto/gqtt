package message_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ysugimoto/gqtt/message"
)

func TestDisconnectMessageEncodeDecode(t *testing.T) {
	t.Run("Omit property should be success", func(t *testing.T) {
		d := message.NewDisconnect(message.NormalDisconnection)
		buf, err := d.Encode()
		assert.NoError(t, err)

		f, p, err := message.ReceiveFrame(bytes.NewReader(buf))
		assert.NoError(t, err)
		assert.Exactly(t, message.DISCONNECT, f.Type)
		assert.False(t, f.DUP)
		assert.Equal(t, message.QoS0, f.QoS)
		assert.False(t, f.RETAIN)

		d, err = message.ParseDisconnect(f, p)
		assert.NoError(t, err)
		assert.Equal(t, message.NormalDisconnection, d.ReasonCode)
		assert.Nil(t, d.Property)
	})

	t.Run("Full case", func(t *testing.T) {
		d := message.NewDisconnect(message.ServerShuttingDown)
		d.Property = &message.DisconnectProperty{
			SessionExpiryInterval: 10000,
			ServerReference:       "mqtt://example.com",
			ReasonString:          "server started to shutting down",
			UserProperty: map[string]string{
				"foo": "bar",
			},
		}
		buf, err := d.Encode()
		assert.NoError(t, err)

		f, p, err := message.ReceiveFrame(bytes.NewReader(buf))
		assert.NoError(t, err)
		assert.Exactly(t, message.DISCONNECT, f.Type)
		assert.False(t, f.DUP)
		assert.Equal(t, message.QoS0, f.QoS)
		assert.False(t, f.RETAIN)

		d, err = message.ParseDisconnect(f, p)
		assert.NoError(t, err)
		assert.Equal(t, message.ServerShuttingDown, d.ReasonCode)
		assert.NotNil(t, d.Property)
		assert.Equal(t, uint32(10000), d.Property.SessionExpiryInterval)
		assert.Equal(t, "mqtt://example.com", d.Property.ServerReference)
		assert.Equal(t, "server started to shutting down", d.Property.ReasonString)
		assert.Contains(t, d.Property.UserProperty, "foo")
		assert.Equal(t, "bar", d.Property.UserProperty["foo"])
	})
}
