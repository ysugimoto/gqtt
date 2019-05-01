package message

import (
	"io"

	"github.com/pkg/errors"
)

type UnsubAck struct {
	*Frame

	PacketId    uint16
	Property    *UnsubAckProperty
	ReasonCodes []ReasonCode
}

type UnsubAckProperty struct {
	ReasonString string
	UserProperty map[string]string
}

func (p *UnsubAckProperty) ToProp() *Property {
	return &Property{
		ReasonString: p.ReasonString,
		UserProperty: p.UserProperty,
	}
}

func ParseUnsubAck(f *Frame, p []byte) (u *UnsubAck, err error) {
	u = &UnsubAck{
		Frame:       f,
		ReasonCodes: make([]ReasonCode, 0),
	}

	dec := newDecoder(p)
	if u.PacketId, err = dec.Uint16(); err != nil {
		return nil, errors.Wrap(err, "failed to decode as uint16")
	}
	if prop, err := dec.Property(); err != nil {
		return nil, errors.Wrap(err, "failed to decode property")
	} else if prop != nil {
		u.Property = prop.ToUnsubAck()
	}
	for {
		if rc, err := dec.Uint(); err != nil {
			if err == io.EOF {
				break
			}
			return nil, errors.Wrap(err, "failed to decode as uint")
		} else if !IsReasonCodeAvailable(rc) {
			return nil, errors.New("invalid reason code supplied")
		} else {
			u.ReasonCodes = append(u.ReasonCodes, ReasonCode(rc))
		}
	}
	return u, nil
}

func NewUnsubAck(rcs ...ReasonCode) *UnsubAck {
	return &UnsubAck{
		Frame:       newFrame(UNSUBACK),
		ReasonCodes: rcs,
	}
}

func (u *UnsubAck) AddReasonCode(rcs ...ReasonCode) {
	for _, v := range rcs {
		u.ReasonCodes = append(u.ReasonCodes, v)
	}
}

func (u *UnsubAck) Validate() error {
	if u.PacketId == 0 {
		return errors.New("PacketId must not be zero")
	}
	if len(u.ReasonCodes) == 0 {
		return errors.New("SubAck Reason codes must exits at least one")
	}
	return nil
}

func (u *UnsubAck) Encode() ([]byte, error) {
	if err := u.Validate(); err != nil {
		return nil, errors.Wrap(err, "UNSUBACK validation error")
	}
	enc := newEncoder()
	enc.Uint16(u.PacketId)
	if u.Property != nil {
		enc.Property(u.Property.ToProp())
	} else {
		enc.Uint(0)
	}
	for _, v := range u.ReasonCodes {
		enc.Uint(v.Byte())
	}

	return u.Frame.Encode(enc.Get()), nil
}
