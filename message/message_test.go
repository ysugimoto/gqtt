package message_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ysugimoto/gqtt/message"
)

type frameWrapper struct {
	*message.Frame
}

func (f *frameWrapper) Encode() ([]byte, error) {
	return f.Frame.Encode([]byte{}), nil
}

func TestMessageDuplicate(t *testing.T) {
	w := frameWrapper{
		Frame: &message.Frame{
			Type: message.PINGREQ,
		},
	}
	buf, err := w.Encode()
	assert.NoError(t, err)

	f, _, err := message.ReceiveFrame(bytes.NewReader(buf))
	assert.NoError(t, err)
	assert.Exactly(t, message.PINGREQ, f.Type)
	assert.False(t, f.DUP)
	assert.Equal(t, message.QoS0, f.QoS)
	assert.False(t, f.RETAIN)
	assert.Equal(t, uint64(0), f.Size)

	w.Duplicate()
	buf, err = w.Encode()
	assert.NoError(t, err)

	f, _, err = message.ReceiveFrame(bytes.NewReader(buf))
	assert.NoError(t, err)
	assert.Exactly(t, message.PINGREQ, f.Type)
	assert.True(t, f.DUP)
	assert.Equal(t, message.QoS0, f.QoS)
	assert.False(t, f.RETAIN)
	assert.Equal(t, uint64(0), f.Size)
}
