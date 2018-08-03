package message

import (
	"errors"
)

type PubAck struct {
	*Frame

	MessageID uint16
}

func ParsePubAck(f *Frame, p []byte) (*PubAck, error) {
	return &PubAck{
		Frame:     f,
		MessageID: uint16(((int(p[0]) << 8) | int(p[1]))),
	}, nil
}

func NewPubAck(f *Frame) *PubAck {
	return &PubAck{
		Frame: f,
	}
}

func (p *PubAck) Validate() error {
	if p.MessageID == 0 {
		return errors.New("MessageID must not be zero")
	}
	return nil
}

func (p *PubAck) Encode() ([]byte, error) {
	if err := p.Validate(); err != nil {
		return nil, err
	}
	payload := make([]byte, 0)
	payload = append(payload, byte(p.MessageID>>8), byte(p.MessageID&0xFF))

	return p.Frame.Encode(payload), nil
}
