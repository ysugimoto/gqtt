package message_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ysugimoto/gqtt/message"
)

func TestAuthMessageEncodeDecode(t *testing.T) {
	t.Run("Omit property should be success", func(t *testing.T) {
		a := message.NewAuth()
		buf, err := a.Encode()
		assert.NoError(t, err)

		f, p, err := message.ReceiveFrame(bytes.NewReader(buf))
		assert.NoError(t, err)
		assert.Exactly(t, message.AUTH, f.Type)
		assert.False(t, f.DUP)
		assert.Equal(t, message.QoS0, f.QoS)
		assert.False(t, f.RETAIN)

		a, err = message.ParseAuth(f, p)
		assert.NoError(t, err)
		assert.Equal(t, message.Success, a.ReasonCode)
		assert.Nil(t, a.Property)
	})

	t.Run("Full case", func(t *testing.T) {
		a := message.NewAuth()
		a.Property = &message.AuthProperty{
			AuthenticationMethod: "BASIC",
			AuthenticationData:   []byte("foo:Bar"),
			ReasonString:         "need authenticate",
			UserProperty: map[string]string{
				"foo": "bar",
			},
		}
		buf, err := a.Encode()
		assert.NoError(t, err)

		f, p, err := message.ReceiveFrame(bytes.NewReader(buf))
		assert.NoError(t, err)
		assert.Exactly(t, message.AUTH, f.Type)
		assert.False(t, f.DUP)
		assert.Equal(t, message.QoS0, f.QoS)
		assert.False(t, f.RETAIN)

		a, err = message.ParseAuth(f, p)
		assert.NoError(t, err)
		assert.Equal(t, message.Success, a.ReasonCode)
		assert.NotNil(t, a.Property)
		assert.Equal(t, "BASIC", a.Property.AuthenticationMethod)
		assert.Equal(t, []byte("foo:Bar"), a.Property.AuthenticationData)
		assert.Equal(t, "need authenticate", a.Property.ReasonString)
		assert.Contains(t, a.Property.UserProperty, "foo")
		assert.Equal(t, "bar", a.Property.UserProperty["foo"])
	})
}
