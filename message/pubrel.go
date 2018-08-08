package message

import (
	"errors"
	"io"
)

type PubRel struct {
	*Frame

	PacketId   uint16
	ReasonCode ReasonCode
	Property   *PubRelProperty
}

type PubRelProperty struct {
	ReasonString string
	UserProperty map[string]string
}

func (p *PubRelProperty) ToProp() *Property {
	return &Property{
		ReasonString: p.ReasonString,
		UserProperty: p.UserProperty,
	}
}

func ParsePubRel(f *Frame, p []byte) (pr *PubRel, err error) {
	pr = &PubRel{
		Frame:      f,
		ReasonCode: Success,
	}
	dec := newDecoder(p)
	if pr.PacketId, err = dec.Uint16(); err != nil {
		return nil, err
	}
	if rc, err := dec.Uint(); err != nil {
		if err != io.EOF {
			return nil, err
		}
		return pr, nil
	} else if !IsReasonCodeAvailable(rc) {
		return nil, errors.New("unexpected reason code spcified")
	} else {
		pr.ReasonCode = ReasonCode(rc)
	}

	if prop, err := dec.Property(); err != nil {
		if err != io.EOF {
			return nil, err
		}
	} else if prop != nil {
		pr.Property = prop.ToPubRel()
	}
	return pr, nil
}

func NewPubRel(packetId uint16, opts ...option) *PubRel {
	return &PubRel{
		Frame:    newFrame(PUBREL, opts...),
		PacketId: packetId,
	}
}

func (p *PubRel) Validate() error {
	if p.PacketId == 0 {
		return errors.New("PacketId must not be zero")
	}
	return nil
}

func (p *PubRel) Encode() ([]byte, error) {
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
