package message

import (
	"bytes"
)

type encoder struct {
	w *bytes.Buffer
}

func newEncoder() *encoder {
	return &encoder{
		w: new(bytes.Buffer),
	}
}

func (e *encoder) Get() []byte {
	return e.w.Bytes()
}

func (e *encoder) Byte(b byte) {
	e.w.WriteByte(b)
}

func (e *encoder) Int(i int) {
	e.w.WriteByte(byte(i))
}

func (e *encoder) Uint(i uint8) {
	e.Int(int(i))
}

func (e *encoder) Int16(i int) {
	e.w.Write([]byte{byte(i >> 8), byte(i & 0xFF)})
}

func (e *encoder) Uint16(i uint16) {
	e.Int16(int(i))
}

func (e *encoder) Int32(i int) {
	e.w.Write([]byte{
		byte(i >> 24),
		byte(i >> 16),
		byte(i >> 8),
		byte(i & 0xFF),
	})
}

func (e *encoder) Uint32(i uint32) {
	e.Int32(int(i))
}

func (e *encoder) String(str string) {
	e.Binary([]byte(str))
}

func (e *encoder) StringAll(str string) {
	e.BinaryAll([]byte(str))
}

func (e *encoder) Binary(b []byte) {
	e.Int16(len(b))
	e.w.Write(b)
}

func (e *encoder) BinaryAll(b []byte) {
	e.w.Write(b)
}

func (e *encoder) Variable(v int) {
	b := []byte{}
	for v > 0 {
		digit := v % 0x80
		v /= 0x80
		if v > 0 {
			digit |= 0x80
		}
		b = append(b, byte(digit))
	}
	e.w.Write(b)
}

func (e *encoder) Property(p *Property) {
	buf := make([]byte, 0)
	if p.PayloadFormatIndicator > 0 {
		buf = append(buf, PayloadFormatIndicator.Byte())
		buf = append(buf, encodeUint(p.PayloadFormatIndicator))
	}
	if p.MessageExpiryInterval > 0 {
		buf = append(buf, MessageExpiryInterval.Byte())
		buf = append(buf, encodeUint32(p.MessageExpiryInterval)...)
	}
	if p.ContentType != "" {
		buf = append(buf, ContentType.Byte())
		buf = append(buf, encodeString(p.ContentType)...)
	}
	if p.ResponseTopic != "" {
		buf = append(buf, ResponseTopic.Byte())
		buf = append(buf, encodeString(p.ResponseTopic)...)
	}
	if p.CorrelationData != nil && len(p.CorrelationData) > 0 {
		buf = append(buf, CorrelationData.Byte())
		buf = append(buf, encodeBinary(p.CorrelationData)...)
	}
	if p.SubscriptionIdentifier > 0 {
		buf = append(buf, SubscriptionIdentifier.Byte())
		buf = append(buf, encodeVariable(int(p.SubscriptionIdentifier))...)
	}
	if p.SessionExpiryInterval > 0 {
		buf = append(buf, SessionExpiryInterval.Byte())
		buf = append(buf, encodeUint32(p.SessionExpiryInterval)...)
	}
	if p.AssignedClientIdentifier != "" {
		buf = append(buf, AssignedClientIdentifier.Byte())
		buf = append(buf, encodeString(p.AssignedClientIdentifier)...)
	}
	if p.ServerKeepAlive > 0 {
		buf = append(buf, ServerKeepAlive.Byte())
		buf = append(buf, encodeUint16(p.ServerKeepAlive)...)
	}
	if p.AuthenticationMethod != "" {
		buf = append(buf, AuthenticationMethod.Byte())
		buf = append(buf, encodeString(p.AuthenticationMethod)...)
	}
	if p.AuthenticationData != nil && len(p.AuthenticationData) > 0 {
		buf = append(buf, AuthenticationData.Byte())
		buf = append(buf, encodeBinary(p.AuthenticationData)...)
	}
	if p.RequestProblemInformation {
		buf = append(buf, RequestProblemInformation.Byte())
		buf = append(buf, encodeInt(encodeBool(p.RequestProblemInformation)))
	}
	if p.WillDelayInterval > 0 {
		buf = append(buf, WillDelayInterval.Byte())
		buf = append(buf, encodeUint32(p.WillDelayInterval)...)
	}
	if p.RequestResponseInformation {
		buf = append(buf, RequestResponseInformation.Byte())
		buf = append(buf, encodeInt(encodeBool(p.RequestResponseInformation)))
	}
	if p.ResponseInformation != "" {
		buf = append(buf, ResponseInformation.Byte())
		buf = append(buf, encodeString(p.ResponseInformation)...)
	}
	if p.ServerReference != "" {
		buf = append(buf, ServerReference.Byte())
		buf = append(buf, encodeString(p.ServerReference)...)
	}
	if p.ReasonString != "" {
		buf = append(buf, ReasonString.Byte())
		buf = append(buf, encodeString(p.ReasonString)...)
	}
	if p.ReceiveMaximum > 0 {
		buf = append(buf, ReceiveMaximum.Byte())
		buf = append(buf, encodeUint16(p.ReceiveMaximum)...)
	}
	if p.TopicAliasMaximum > 0 {
		buf = append(buf, TopicAliasMaximum.Byte())
		buf = append(buf, encodeUint16(p.TopicAliasMaximum)...)
	}
	if p.TopicAlias > 0 {
		buf = append(buf, TopicAlias.Byte())
		buf = append(buf, encodeUint16(p.TopicAlias)...)
	}
	if p.MaximumQoS > 0 {
		buf = append(buf, MaximumQoS.Byte())
		buf = append(buf, encodeUint(p.MaximumQoS))
	}
	if p.RetainAvalilable {
		buf = append(buf, RetainAvalilable.Byte())
		buf = append(buf, encodeInt(encodeBool(p.RetainAvalilable)))
	}
	if p.UserProperty != nil && len(p.UserProperty) > 0 {
		for k, v := range p.UserProperty {
			buf = append(buf, UserProperty.Byte())
			buf = append(buf, encodeString(k)...)
			buf = append(buf, encodeString(v)...)
		}
	}
	if p.MaximumPacketSize > 0 {
		buf = append(buf, MaximumPacketSize.Byte())
		buf = append(buf, encodeUint32(p.MaximumPacketSize)...)
	}
	if p.WildcardSubscriptionAvailable {
		buf = append(buf, WildcardSubscriptionAvailable.Byte())
		buf = append(buf, encodeInt(encodeBool(p.WildcardSubscriptionAvailable)))
	}
	if p.SubscrptionIdentifierAvailable {
		buf = append(buf, SubscrptionIdentifierAvailable.Byte())
		buf = append(buf, encodeInt(encodeBool(p.SubscrptionIdentifierAvailable)))
	}
	if p.SharedSubscriptionsAvaliable {
		buf = append(buf, SharedSubscriptionsAvaliable.Byte())
		buf = append(buf, encodeInt(encodeBool(p.SharedSubscriptionsAvaliable)))
	}

	e.Variable(len(buf))
	e.w.Write(buf)
}
