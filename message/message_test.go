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

	f, p, err := message.ReceiveFrame(bytes.NewReader(buf))
	assert.NoError(t, err)
	assert.Exactly(t, f.Type, message.PINGREQ)
	assert.Exactly(t, f.DUP, false)
	assert.Equal(t, f.QoS, uint8(0))
	assert.Exactly(t, f.RETAIN, false)
	assert.Equal(t, f.Size, uint64(0))
	assert.Equal(t, f.Size, uint64(len(p)))

	w.Duplicate()
	buf, err = w.Encode()
	assert.NoError(t, err)

	f, p, err = message.ReceiveFrame(bytes.NewReader(buf))
	assert.NoError(t, err)
	assert.Exactly(t, f.Type, message.PINGREQ)
	assert.Exactly(t, f.DUP, true)
	assert.Equal(t, f.QoS, uint8(0))
	assert.Exactly(t, f.RETAIN, false)
	assert.Equal(t, f.Size, uint64(0))
	assert.Equal(t, f.Size, uint64(len(p)))
}
