package message

import (
	"errors"
	"io"
)

type Auth struct {
	*Frame

	ReasonCode ReasonCode
	Property   *AuthProperty
}

type AuthProperty struct {
	AuthenticationMethod string
	AuthenticationData   []byte
	ReasonString         string
	UserProperty         map[string]string
}

func (a *AuthProperty) ToProp() *Property {
	return &Property{
		AuthenticationMethod: a.AuthenticationMethod,
		AuthenticationData:   a.AuthenticationData,
		ReasonString:         a.ReasonString,
		UserProperty:         a.UserProperty,
	}
}

func ParseAuth(f *Frame, p []byte) (a *Auth, err error) {
	a = &Auth{
		Frame:      f,
		ReasonCode: Success,
	}

	dec := newDecoder(p)
	if rc, err := dec.Uint(); err != nil {
		if err != io.EOF {
			return nil, err
		}
		return a, nil
	} else if !IsReasonCodeAvailable(rc) {
		return nil, errors.New("invalid reason code specified")
	} else {
		a.ReasonCode = ReasonCode(rc)
	}

	if prop, err := dec.Property(); err != nil {
		if err != io.EOF {
			return nil, err
		}
	} else if prop != nil {
		a.Property = prop.ToAuth()
	}

	return a, nil
}

func NewAuth(opts ...option) *Auth {
	return &Auth{
		Frame:      newFrame(AUTH, opts...),
		ReasonCode: Success,
	}
}

func (a *Auth) Encode() ([]byte, error) {
	enc := newEncoder()
	enc.Byte(a.ReasonCode.Byte())
	if a.Property != nil {
		enc.Property(a.Property.ToProp())
	}
	return a.Frame.Encode(enc.Get()), nil
}