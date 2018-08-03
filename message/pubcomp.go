package message

import (
	"errors"
)

type PubComp struct {
	*Frame

	MessageID uint16
}

func ParsePubComp(f *Frame, p []byte) (*PubComp, error) {
	return &PubComp{
		Frame:     f,
		MessageID: uint16(((int(p[0]) << 8) | int(p[1]))),
	}, nil
}

func NewPubComp(f *Frame) *PubComp {
	return &PubComp{
		Frame: f,
	}
}

func (p *PubComp) Validate() error {
	if p.MessageID == 0 {
		return errors.New("MessageID must not be zero")
	}
	return nil
}

func (p *PubComp) Encode() ([]byte, error) {
	if err := p.Validate(); err != nil {
		return nil, err
	}
	payload := make([]byte, 0)
	payload = append(payload, byte(p.MessageID>>8), byte(p.MessageID&0xFF))

	return p.Frame.Encode(payload), nil
}
