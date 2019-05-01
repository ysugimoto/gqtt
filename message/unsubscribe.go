package message

import (
	"io"

	"github.com/pkg/errors"
)

type Unsubscribe struct {
	*Frame

	PacketId uint16
	Topics   []string
	Property *UnsubscribeProperty
}

type UnsubscribeProperty struct {
	UserProperty map[string]string
}

func (p *UnsubscribeProperty) ToProp() *Property {
	return &Property{
		UserProperty: p.UserProperty,
	}
}

func ParseUnsubscribe(f *Frame, p []byte) (u *Unsubscribe, err error) {
	u = &Unsubscribe{
		Frame:  f,
		Topics: make([]string, 0),
	}

	dec := newDecoder(p)
	if u.PacketId, err = dec.Uint16(); err != nil {
		return nil, errors.Wrap(err, "failed to decode as uint16")
	}
	if prop, err := dec.Property(); err != nil {
		return nil, errors.Wrap(err, "failed to decode property")
	} else if prop != nil {
		u.Property = prop.ToUnsubscribe()
	}

	for {
		str, err := dec.String()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, errors.Wrap(err, "failed to decode as string")
		}
		u.Topics = append(u.Topics, str)
	}

	return u, nil
}

func NewUnsubscribe(opts ...option) *Unsubscribe {
	return &Unsubscribe{
		Frame:  newFrame(UNSUBSCRIBE, opts...),
		Topics: make([]string, 0),
	}
}

func (u *Unsubscribe) AddTopic(topics ...string) {
	for _, v := range topics {
		u.Topics = append(u.Topics, v)
	}
}

func (u *Unsubscribe) Validate() error {
	if u.PacketId == 0 {
		return errors.New("PacketId is required")
	}
	if len(u.Topics) == 0 {
		return errors.New("At least one topic should unsubscribe")
	}
	return nil
}

func (u *Unsubscribe) Encode() ([]byte, error) {
	if err := u.Validate(); err != nil {
		return nil, errors.Wrap(err, "UNSUBSCRIBE validation error")
	}

	enc := newEncoder()
	enc.Uint16(u.PacketId)
	if u.Property != nil {
		enc.Property(u.Property.ToProp())
	} else {
		enc.Uint(0)
	}

	for _, v := range u.Topics {
		enc.String(v)
	}

	return u.Frame.Encode(enc.Get()), nil
}
