package message

import (
	"errors"
)

type UnsubAck struct {
	*Frame

	MessageID uint16
}

func ParseUnsubAck(f *Frame, p []byte) (*UnsubAck, error) {
	u := &UnsubAck{
		Frame:     f,
		MessageID: uint16(((int(p[0]) << 8) | int(p[1]))),
	}
	return u, nil
}

func NewUnsubAck(f *Frame) *UnsubAck {
	return &UnsubAck{
		Frame: f,
	}
}

func (u *UnsubAck) Validate() error {
	if u.MessageID == 0 {
		return errors.New("MessageID must not be zero")
	}
	return nil
}

func (u *UnsubAck) Encode() ([]byte, error) {
	if err := u.Validate(); err != nil {
		return nil, err
	}
	payload := make([]byte, 0)
	payload = append(payload, byte(u.MessageID>>8), byte(u.MessageID&0xFF))

	return u.Frame.Encode(payload), nil
}
