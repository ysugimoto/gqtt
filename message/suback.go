package message

import (
	"errors"
	"io"
)

type SubAck struct {
	*Frame

	PacketId    uint16
	Property    *SubAckProperty
	ReasonCodes []ReasonCode
}

type SubAckProperty struct {
	ReasonString string
	UserProperty map[string]string
}

func (p *SubAckProperty) ToProp() *Property {
	return &Property{
		ReasonString: p.ReasonString,
		UserProperty: p.UserProperty,
	}
}

func ParseSubAck(f *Frame, p []byte) (s *SubAck, err error) {
	s = &SubAck{
		Frame:       f,
		ReasonCodes: make([]ReasonCode, 0),
	}

	dec := newDecoder(p)
	if s.PacketId, err = dec.Uint16(); err != nil {
		return nil, err
	}
	if prop, err := dec.Property(); err != nil {
		return nil, err
	} else if prop != nil {
		s.Property = prop.ToSubAck()
	}
	for {
		if rc, err := dec.Uint(); err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		} else if !IsReasonCodeAvailable(rc) {
			return nil, errors.New("invalid reason code supplied")
		} else {
			s.ReasonCodes = append(s.ReasonCodes, ReasonCode(rc))
		}
	}
	return s, nil
}

func NewSubAck(packetId uint16, rcs ...ReasonCode) *SubAck {
	return &SubAck{
		Frame:       newFrame(SUBACK),
		PacketId:    packetId,
		ReasonCodes: rcs,
	}
}

func (s *SubAck) AddReasonCode(rcs ...ReasonCode) {
	for _, v := range rcs {
		s.ReasonCodes = append(s.ReasonCodes, v)
	}
}

func (s *SubAck) Validate() error {
	if s.PacketId == 0 {
		return errors.New("PacketId must not be zero")
	}
	if len(s.ReasonCodes) == 0 {
		return errors.New("SubAck Reason codes must exits at least one")
	}
	return nil
}

func (s *SubAck) Encode() ([]byte, error) {
	if err := s.Validate(); err != nil {
		return nil, err
	}
	enc := newEncoder()
	enc.Uint16(s.PacketId)
	if s.Property != nil {
		enc.Property(s.Property.ToProp())
	} else {
		enc.Uint(0)
	}
	for _, v := range s.ReasonCodes {
		enc.Uint(v.Byte())
	}

	return s.Frame.Encode(enc.Get()), nil
}
