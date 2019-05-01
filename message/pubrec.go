package message

import (
	"io"

	"github.com/pkg/errors"
)

type PubRec struct {
	*Frame

	PacketId   uint16
	ReasonCode ReasonCode
	Property   *PubRecProperty
}

type PubRecProperty struct {
	ReasonString string
	UserProperty map[string]string
}

func (p *PubRecProperty) ToProp() *Property {
	return &Property{
		ReasonString: p.ReasonString,
		UserProperty: p.UserProperty,
	}
}

func ParsePubRec(f *Frame, p []byte) (pr *PubRec, err error) {
	pr = &PubRec{
		Frame:      f,
		ReasonCode: Success,
	}
	dec := newDecoder(p)
	if pr.PacketId, err = dec.Uint16(); err != nil {
		return nil, errors.Wrap(err, "failed to decode as uint16")
	}
	if rc, err := dec.Uint(); err != nil {
		if err != io.EOF {
			return nil, errors.Wrap(err, "failed to decode as uint")
		}
		return pr, nil
	} else if !IsReasonCodeAvailable(rc) {
		return nil, errors.New("unexpected reason code spcified")
	} else {
		pr.ReasonCode = ReasonCode(rc)
	}

	if prop, err := dec.Property(); err != nil {
		if err != io.EOF {
			return nil, errors.Wrap(err, "failed to decode property")
		}
	} else if prop != nil {
		pr.Property = prop.ToPubRec()
	}
	return pr, nil
}

func NewPubRec(packetId uint16, opts ...option) *PubRec {
	return &PubRec{
		Frame:    newFrame(PUBREC, opts...),
		PacketId: packetId,
	}
}

func (p *PubRec) Validate() error {
	if p.PacketId == 0 {
		return errors.New("Packet ID must not be zero")
	}
	return nil
}

func (p *PubRec) Encode() ([]byte, error) {
	if err := p.Validate(); err != nil {
		return nil, errors.Wrap(err, "PUBREC validation error")
	}

	enc := newEncoder()
	enc.Uint16(p.PacketId)
	enc.Byte(p.ReasonCode.Byte())
	if p.Property != nil {
		enc.Property(p.Property.ToProp())
	}

	return p.Frame.Encode(enc.Get()), nil
}
