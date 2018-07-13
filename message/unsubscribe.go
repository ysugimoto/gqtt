package message

import (
	"errors"
)

type Unsubscribe struct {
	*Frame

	MessageID uint16
	Topics    []string
}

func ParseUnsubscribe(f *Frame, p []byte) (*Unsubscribe, error) {
	u := &Unsubscribe{
		Frame:  f,
		Topics: make([]string, 0),
	}

	var size, i int
	u.MessageID = uint16(((int(p[i]) << 8) | int(p[i+1])))
	i += 2
	for i < len(p) {
		size = ((int(p[i]) << 8) | int(p[i+1]))
		i += 2
		u.Topics = append(u.Topics, string(p[i:(i+size)]))
		i += size
	}
	return u, nil
}

func NewUnsubscribe(f *Frame) *Unsubscribe {
	return &Unsubscribe{
		Frame:  f,
		Topics: make([]string, 0),
	}
}

func (u *Unsubscribe) AddTopic(topics ...string) {
	for _, v := range topics {
		u.Topics = append(u.Topics, v)
	}
}

func (u *Unsubscribe) Validate() error {
	if len(u.Topics) == 0 {
		return errors.New("At least one topic should subscribe")
	}
	if u.MessageID == 0 {
		return errors.New("MessageID is required")
	}
	return nil
}

func (u *Unsubscribe) Encode() ([]byte, error) {
	if err := u.Validate(); err != nil {
		return nil, err
	}
	buf := make([]byte, 0)
	var size int

	buf = append(buf, byte(u.MessageID>>8), byte(u.MessageID&0xFF))
	for _, v := range u.Topics {
		size = len([]byte(v))
		buf = append(buf, byte(size>>8), byte(size&0xFF))
		buf = append(buf, []byte(v)...)
	}

	return u.Frame.Encode(buf), nil
}
