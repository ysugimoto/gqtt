package message_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ysugimoto/gqtt/message"
)

func TestPubCompMessageFailedIfMessageIDIsZero(t *testing.T) {
	a := message.NewPubComp(&message.Frame{
		Type: message.PUBCOMP,
	})
	buf, err := a.Encode()
	assert.Error(t, err)
	assert.Nil(t, buf)
}

func TestPubCompMessageEncodeDecodeOK(t *testing.T) {
	a := message.NewPubComp(&message.Frame{
		Type: message.PUBCOMP,
	})
	a.MessageID = uint16(1000)
	buf, err := a.Encode()
	assert.NoError(t, err)

	f, p, err := message.ReceiveFrame(bytes.NewReader(buf))
	assert.NoError(t, err)
	assert.Exactly(t, f.Type, message.PUBCOMP)
	assert.Exactly(t, f.DUP, false)
	assert.Equal(t, f.QoS, uint8(0))
	assert.Exactly(t, f.RETAIN, false)
	assert.Equal(t, f.Size, uint64(len(p)))

	a, err = message.ParsePubComp(f, p)
	assert.NoError(t, err)
	assert.Equal(t, a.MessageID, uint16(1000))
}
