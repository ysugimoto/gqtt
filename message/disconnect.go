package message

import (
	"io"

	"github.com/pkg/errors"
)

type Disconnect struct {
	*Frame

	ReasonCode ReasonCode
	Property   *DisconnectProperty
}

type DisconnectProperty struct {
	SessionExpiryInterval uint32
	ServerReference       string
	ReasonString          string
	UserProperty          map[string]string
}

func (d *DisconnectProperty) ToProp() *Property {
	return &Property{
		SessionExpiryInterval: d.SessionExpiryInterval,
		ServerReference:       d.ServerReference,
		ReasonString:          d.ReasonString,
		UserProperty:          d.UserProperty,
	}
}

func ParseDisconnect(f *Frame, p []byte) (d *Disconnect, err error) {
	d = &Disconnect{
		Frame:      f,
		ReasonCode: NormalDisconnection,
	}

	dec := newDecoder(p)
	if rc, err := dec.Uint(); err != nil {
		if err != io.EOF {
			return nil, errors.Wrap(err, "failed to decode as uint")
		}
		return d, nil
	} else if !IsReasonCodeAvailable(rc) {
		return nil, errors.New("invalid reason code specified")
	} else {
		d.ReasonCode = ReasonCode(rc)
	}

	if prop, err := dec.Property(); err != nil {
		if err != io.EOF {
			return nil, errors.Wrap(err, "failed to decode property")
		}
	} else if prop != nil {
		d.Property = prop.ToDisconnect()
	}

	return d, nil
}

func NewDisconnect(code ReasonCode, opts ...option) *Disconnect {
	return &Disconnect{
		Frame:      newFrame(DISCONNECT, opts...),
		ReasonCode: code,
	}
}

func (d *Disconnect) Encode() ([]byte, error) {
	enc := newEncoder()
	enc.Byte(d.ReasonCode.Byte())
	if d.Property != nil {
		enc.Property(d.Property.ToProp())
	}
	return d.Frame.Encode(enc.Get()), nil
}
