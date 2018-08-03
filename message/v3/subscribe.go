package message

import (
	"errors"
)

type SubscribeTopic struct {
	Name string
	QoS  uint8
}

type Subscribe struct {
	*Frame

	MessageID uint16
	Topics    []SubscribeTopic

	ClientId string
}

func ParseSubscribe(f *Frame, p []byte) (*Subscribe, error) {
	s := &Subscribe{
		Frame:  f,
		Topics: make([]SubscribeTopic, 0),
	}

	var size, i int
	s.MessageID = uint16(((int(p[i]) << 8) | int(p[i+1])))
	i += 2
	for i < len(p) {
		t := SubscribeTopic{}
		size = ((int(p[i]) << 8) | int(p[i+1]))
		i += 2
		t.Name = string(p[i:(i + size)])
		i += size
		t.QoS = uint8(p[i])
		i++
		s.Topics = append(s.Topics, t)
	}
	return s, nil
}

func NewSubscribe(f *Frame) *Subscribe {
	return &Subscribe{
		Frame:  f,
		Topics: make([]SubscribeTopic, 0),
	}
}

func (s *Subscribe) AddTopic(name string, qos uint8) {
	s.Topics = append(s.Topics, SubscribeTopic{
		Name: name,
		QoS:  qos,
	})
}

func (s *Subscribe) Validate() error {
	if len(s.Topics) == 0 {
		return errors.New("At least one topic should subscribe")
	}
	if s.MessageID == 0 {
		return errors.New("MessageID is required")
	}
	return nil
}

func (s *Subscribe) Encode() ([]byte, error) {
	if err := s.Validate(); err != nil {
		return nil, err
	}
	buf := make([]byte, 0)
	var size int

	buf = append(buf, byte(s.MessageID>>8), byte(s.MessageID&0xFF))
	for _, v := range s.Topics {
		size = len([]byte(v.Name))
		buf = append(buf, byte(size>>8), byte(size&0xFF))
		buf = append(buf, []byte(v.Name)...)
		buf = append(buf, byte(v.QoS))
	}

	return s.Frame.Encode(buf), nil
}
