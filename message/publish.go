package message

import (
	"errors"
)

type Publish struct {
	*Frame

	TopicName string
	MessageID uint16
	Body      string

	ClientId string
}

func ParsePublish(f *Frame, p []byte) (*Publish, error) {
	pb := &Publish{
		Frame: f,
	}

	var size, i int
	size = ((int(p[i]) << 8) | int(p[i+1]))
	i += 2
	pb.TopicName = string(p[i:(i + size)])
	i += size
	// Message ID will present on QoS is greater than 0
	if f.QoS > 0 {
		pb.MessageID = uint16(((int(p[i]) << 8) | int(p[i+1])))
		i += 2
	}
	pb.Body = string(p[i:])

	return pb, nil
}

func NewPublish(f *Frame) *Publish {
	return &Publish{
		Frame: f,
	}
}

func (p *Publish) Validate() error {
	if p.TopicName == "" {
		return errors.New("TopicName is required")
	}
	if p.Frame.QoS > 0 && p.MessageID == 0 {
		return errors.New("MessageID is required when QoS is greater than 0")
	}
	return nil
}

func (p *Publish) Encode() ([]byte, error) {
	if err := p.Validate(); err != nil {
		return nil, err
	}
	buf := make([]byte, 0)
	var size int

	size = len([]byte(p.TopicName))
	buf = append(buf, byte(size>>8), byte(size&0xFF))
	buf = append(buf, []byte(p.TopicName)...)
	if p.Frame.QoS > 0 {
		buf = append(buf, byte(p.MessageID>>8), byte(p.MessageID&0xFF))
	}
	buf = append(buf, []byte(p.Body)...)

	return p.Frame.Encode(buf), nil
}
