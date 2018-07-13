package message

import (
	"errors"
)

type PubRec struct {
	*Frame

	MessageID uint16
}

func ParsePubRec(f *Frame, p []byte) (*PubRec, error) {
	return &PubRec{
		Frame:     f,
		MessageID: uint16(((int(p[0]) << 8) | int(p[1]))),
	}, nil
}

func NewPubRec(f *Frame) *PubRec {
	return &PubRec{
		Frame: f,
	}
}

func (p *PubRec) Validate() error {
	if p.MessageID == 0 {
		return errors.New("MessageID must not be zero")
	}
	return nil
}

func (p *PubRec) Encode() ([]byte, error) {
	if err := p.Validate(); err != nil {
		return nil, err
	}
	payload := make([]byte, 0)
	payload = append(payload, byte(p.MessageID>>8), byte(p.MessageID&0xFF))

	return p.Frame.Encode(payload), nil
}
