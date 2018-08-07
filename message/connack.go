package message

import (
	"errors"
	"io"
)

type ConnAck struct {
	*Frame

	Property           *ConnAckProperty
	SessionPresentFlag bool
	ReasonCode         ReasonCode
}

type ConnAckProperty struct {
	SessionExpiryInterval          uint32
	AssignedClientIdentifier       string
	ServerKeepAlive                uint16
	AuthenticationMethod           string
	AuthenticationData             []byte
	ResponseInformation            string
	ServerReference                string
	ReasonString                   string
	ReceiveMaximum                 uint16
	TopicAliasMaximum              uint16
	MaximumQoS                     uint8
	RetainAvalilable               bool
	UserProperty                   map[string]string
	MaximumPacketSize              uint32
	WildcardSubscriptionAvailable  bool
	SubscrptionIdentifierAvailable bool
	SharedSubscriptionsAvaliable   bool
}

func (c *ConnAckProperty) ToProp() *Property {
	return &Property{
		SessionExpiryInterval:          c.SessionExpiryInterval,
		AssignedClientIdentifier:       c.AssignedClientIdentifier,
		ServerKeepAlive:                c.ServerKeepAlive,
		AuthenticationMethod:           c.AuthenticationMethod,
		AuthenticationData:             c.AuthenticationData,
		ResponseInformation:            c.ResponseInformation,
		ServerReference:                c.ServerReference,
		ReasonString:                   c.ReasonString,
		ReceiveMaximum:                 c.ReceiveMaximum,
		TopicAliasMaximum:              c.TopicAliasMaximum,
		MaximumQoS:                     c.MaximumQoS,
		RetainAvalilable:               c.RetainAvalilable,
		UserProperty:                   c.UserProperty,
		MaximumPacketSize:              c.MaximumPacketSize,
		WildcardSubscriptionAvailable:  c.WildcardSubscriptionAvailable,
		SubscrptionIdentifierAvailable: c.SubscrptionIdentifierAvailable,
		SharedSubscriptionsAvaliable:   c.SharedSubscriptionsAvaliable,
	}
}

func ParseConnAck(f *Frame, p []byte) (c *ConnAck, err error) {
	c = &ConnAck{
		Frame: f,
	}
	dec := newDecoder(p)
	if i, err := dec.Int(); err != nil {
		return nil, err
	} else {
		c.SessionPresentFlag = (i & 0x01) > 0
	}
	if rc, err := dec.Uint(); err != nil {
		return nil, err
	} else if !IsReasonCodeAvailable(rc) {
		return nil, errors.New("unexpected reason code supplied")
	} else {
		c.ReasonCode = ReasonCode(rc)
	}

	if prop, err := dec.Property(); err != nil {
		if err != io.EOF {
			return nil, err
		}
	} else if prop != nil {
		c.Property = prop.ToConnAck()
	}
	// no payload
	return c, nil
}

func NewConnAck(code ReasonCode, opts ...option) *ConnAck {
	return &ConnAck{
		Frame:      newFrame(CONNACK, opts...),
		ReasonCode: code,
	}
}

func (c *ConnAck) Validate() error {
	if !IsReasonCodeAvailable(uint8(c.ReasonCode)) {
		return errors.New("Invalid reason code")
	}
	return nil
}

func (c *ConnAck) Encode() ([]byte, error) {
	if err := c.Validate(); err != nil {
		return nil, err
	}
	enc := newEncoder()
	if c.SessionPresentFlag {
		enc.Int(1)
	} else {
		enc.Int(0)
	}
	enc.Byte(c.ReasonCode.Byte())
	if c.Property != nil {
		enc.Property(c.Property.ToProp())
	} else {
		enc.Uint(0)
	}
	return c.Frame.Encode(enc.Get()), nil
}
