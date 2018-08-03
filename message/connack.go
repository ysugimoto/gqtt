package message

import (
	"errors"
)

type ConnAckCode uint8

const (
	ConnAckOK ConnAckCode = iota
	ConnAckForbiddenProtocol
	ConnAckForbiddenID
	ConnAckServerNotAvailable
	ConnAckAuthFailed
	ConnAckPermissionDenied
)

type ConnAck struct {
	*Frame

	ReturnCode ConnAckCode
}

func ParseConnAck(f *Frame, p []byte) (*ConnAck, error) {
	return &ConnAck{
		Frame:      f,
		ReturnCode: ConnAckCode(uint8(p[1])),
	}, nil
}

func NewConnAck(f *Frame) *ConnAck {
	return &ConnAck{
		Frame: f,
	}
}

func (c *ConnAck) Validate() error {
	if c.ReturnCode > ConnAckPermissionDenied {
		return errors.New("Invalid return code")
	}
	return nil
}

func (c *ConnAck) Encode() ([]byte, error) {
	if err := c.Validate(); err != nil {
		return nil, err
	}
	payload := []byte{byte(c.ReturnCode), 0}
	return c.Frame.Encode(payload), nil
}
