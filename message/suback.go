package message

import (
	"errors"
)

type SubAck struct {
	*Frame

	MessageID uint16
	QoSs      []uint8
}

func ParseSubAck(f *Frame, p []byte) (*SubAck, error) {
	s := &SubAck{
		Frame:     f,
		MessageID: uint16(((int(p[0]) << 8) | int(p[1]))),
		QoSs:      make([]uint8, 0),
	}
	for i := 2; i < len(p); i++ {
		s.QoSs = append(s.QoSs, uint8(p[i]))
	}
	return s, nil
}

func NewSubAck(f *Frame) *SubAck {
	return &SubAck{
		Frame: f,
		QoSs:  make([]uint8, 0),
	}
}

func (s *SubAck) AddQoS(qoss ...uint8) {
	for _, v := range qoss {
		s.QoSs = append(s.QoSs, v)
	}
}

func (s *SubAck) Validate() error {
	if s.MessageID == 0 {
		return errors.New("MessageID must not be zero")
	}
	if len(s.QoSs) == 0 {
		return errors.New("Ack QoSs must exits at least one")
	}
	return nil
}

func (s *SubAck) Encode() ([]byte, error) {
	if err := s.Validate(); err != nil {
		return nil, err
	}
	payload := make([]byte, 0)
	payload = append(payload, byte(s.MessageID>>8), byte(s.MessageID&0xFF))
	for _, v := range s.QoSs {
		payload = append(payload, byte(v))
	}

	return s.Frame.Encode(payload), nil
}
