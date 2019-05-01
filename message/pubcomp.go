package message

import (
	"io"

	"github.com/pkg/errors"
)

type PubComp struct {
	*Frame

	PacketId   uint16
	ReasonCode ReasonCode
	Property   *PubCompProperty
}

type PubCompProperty struct {
	ReasonString string
	UserProperty map[string]string
}

func (p *PubCompProperty) ToProp() *Property {
	return &Property{
		ReasonString: p.ReasonString,
		UserProperty: p.UserProperty,
	}
}

func ParsePubComp(f *Frame, p []byte) (pc *PubComp, err error) {
	pc = &PubComp{
		Frame:      f,
		ReasonCode: Success,
	}
	dec := newDecoder(p)
	if pc.PacketId, err = dec.Uint16(); err != nil {
		return nil, errors.Wrap(err, "failed to decode as uint16")
	}
	if rc, err := dec.Uint(); err != nil {
		if err != io.EOF {
			return nil, errors.Wrap(err, "failed to decode as uint")
		}
		return pc, nil
	} else if !IsReasonCodeAvailable(rc) {
		return nil, errors.New("unexpected reason code spcified")
	} else {
		pc.ReasonCode = ReasonCode(rc)
	}

	if prop, err := dec.Property(); err != nil {
		if err != io.EOF {
			return nil, errors.Wrap(err, "failed to decode property")
		}
	} else if prop != nil {
		pc.Property = prop.ToPubComp()
	}
	return pc, nil
}

func NewPubComp(packetId uint16, opts ...option) *PubComp {
	return &PubComp{
		Frame:    newFrame(PUBCOMP, opts...),
		PacketId: packetId,
	}
}

func (p *PubComp) Validate() error {
	if p.PacketId == 0 {
		return errors.New("PacketId must not be zero")
	}
	return nil
}

func (p *PubComp) Encode() ([]byte, error) {
	if err := p.Validate(); err != nil {
		return nil, errors.Wrap(err, "PUBCOMP validation error")
	}

	enc := newEncoder()
	enc.Uint16(p.PacketId)
	enc.Byte(p.ReasonCode.Byte())
	if p.Property != nil {
		enc.Property(p.Property.ToProp())
	}

	return p.Frame.Encode(enc.Get()), nil
}
