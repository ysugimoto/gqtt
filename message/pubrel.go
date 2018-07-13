package message

import (
	"errors"
)

type PubRel struct {
	*Frame

	MessageID uint16
}

func ParsePubRel(f *Frame, p []byte) (*PubRel, error) {
	return &PubRel{
		Frame:     f,
		MessageID: uint16(((int(p[0]) << 8) | int(p[1]))),
	}, nil
}

func NewPubRel(f *Frame) *PubRel {
	return &PubRel{
		Frame: f,
	}
}

func (p *PubRel) Validate() error {
	if p.MessageID == 0 {
		return errors.New("MessageID must not be zero")
	}
	return nil
}

func (p *PubRel) Encode() ([]byte, error) {
	if err := p.Validate(); err != nil {
		return nil, err
	}
	payload := make([]byte, 0)
	payload = append(payload, byte(p.MessageID>>8), byte(p.MessageID&0xFF))

	return p.Frame.Encode(payload), nil
}
