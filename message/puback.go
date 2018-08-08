package message

import (
	"errors"
	"io"
)

type PubAck struct {
	*Frame

	PacketId   uint16
	ReasonCode ReasonCode
	Property   *PubAckProperty
}

type PubAckProperty struct {
	ReasonString string
	UserProperty map[string]string
}

func (p *PubAckProperty) ToProp() *Property {
	return &Property{
		ReasonString: p.ReasonString,
		UserProperty: p.UserProperty,
	}
}

func ParsePubAck(f *Frame, p []byte) (pa *PubAck, err error) {
	pa = &PubAck{
		Frame:      f,
		ReasonCode: Success,
	}
	dec := newDecoder(p)
	if pa.PacketId, err = dec.Uint16(); err != nil {
		return nil, err
	}
	if rc, err := dec.Uint(); err != nil {
		if err != io.EOF {
			return nil, err
		}
		return pa, nil
	} else if !IsReasonCodeAvailable(rc) {
		return nil, errors.New("unexpected reason code spcified")
	} else {
		pa.ReasonCode = ReasonCode(rc)
	}

	if prop, err := dec.Property(); err != nil {
		if err != io.EOF {
			return nil, err
		}
	} else if prop != nil {
		pa.Property = prop.ToPubAck()
	}
	return pa, nil
}

func NewPubAck(packetId uint16, opts ...option) *PubAck {
	return &PubAck{
		Frame:    newFrame(PUBACK, opts...),
		PacketId: packetId,
	}
}

func (p *PubAck) Validate() error {
	if p.PacketId == 0 {
		return errors.New("packet ID must not be zero")
	}
	return nil
}

func (p *PubAck) Encode() ([]byte, error) {
	if err := p.Validate(); err != nil {
		return nil, err
	}
	enc := newEncoder()
	enc.Uint16(p.PacketId)
	enc.Byte(p.ReasonCode.Byte())
	if p.Property != nil {
		enc.Property(p.Property.ToProp())
	}

	return p.Frame.Encode(enc.Get()), nil
}
