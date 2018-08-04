package message

import (
	"errors"
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

func (c *ConnAckProperty) Encode() []byte {
	return encodeProperty(&Property{
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
	})
}

func ParseConnAck(f *Frame, p []byte) (*ConnAck, error) {
	c := &ConnAck{
		Frame: f,
	}
	var i, b, psize int
	var rc uint8

	b, i = decodeInt(p, i)
	c.SessionPresentFlag = decodeBool(b & 0x01)
	rc, i = decodeUint(p, i)
	if !IsReasonCodeAvailable(rc) {
		return nil, errors.New("unpected reason code supplied")
	}
	c.ReasonCode = ReasonCode(rc)
	psize, i = decodeInt(p, i)
	if psize > 0 {
		prop, err := decodeProperty(p[i:(i + psize)])
		if err != nil {
			return nil, err
		}
		c.Property = prop.ToConnAck()
	}
	// no payload
	return c, nil
}

func NewConnAck(code ReasonCode) *ConnAck {
	return &ConnAck{
		Frame:      newFrame(CONNACK),
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
	payload := append([]byte{}, byte(encodeBool(c.SessionPresentFlag)), c.ReasonCode.Byte())
	if c.Property != nil {
		payload = append(payload, c.Property.Encode()...)
	}
	return c.Frame.Encode(payload), nil
}
