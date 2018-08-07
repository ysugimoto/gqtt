package message

import (
	"fmt"
)

func encodeBool(b bool) (i int) {
	if b {
		i = 1
	}
	return
}

func decodeBool(i int) (b bool) {
	if i > 0 {
		b = true
	}
	return
}

func encodeString(str string) []byte {
	buf := []byte(str)
	size := len(buf)

	return append(encodeInt16(size), buf...)
}
func encodeBinary(b []byte) []byte {
	size := len(b)
	return append(encodeInt16(size), b...)
}
func encodeVariable(size int) []byte {
	ret := []byte{}
	for size > 0 {
		digit := size % 0x80
		size /= 0x80
		if size > 0 {
			digit |= 0x80
		}
		ret = append(ret, byte(digit))
	}
	return ret
}

func decodeString(buffer []byte, start int) (string, int) {
	size := ((int(buffer[start]) << 8) | int(buffer[start+1]))
	start += 2
	data := string(buffer[start:(start + size)])
	return data, start + size
}
func decodeBinary(buffer []byte, start int) ([]byte, int) {
	size := ((int(buffer[start]) << 8) | int(buffer[start+1]))
	start += 2
	data := buffer[start:(start + size)]
	return data, start + size
}
func decodeVariable(buffer []byte, start int) (uint64, int) {
	var (
		b    int
		size uint64
		mul  uint64 = 1
	)
	for {
		b, start = decodeInt(buffer, start)
		size += uint64(b&0x7F) * mul
		mul *= 0x80
		if b&0x80 == 0 {
			break
		}
	}
	return size, start
}

func encodeInt(v int) byte {
	return byte(v)
}
func encodeInt16(v int) []byte {
	return append([]byte{}, byte(v>>8), byte(v&0xFF))
}
func encodeInt32(v int) []byte {
	return append([]byte{}, byte(v>>24), byte(v>>16), byte(v>>8), byte(v&0xFF))
}
func encodeUint(v uint8) byte {
	return byte(v)
}
func encodeUint16(v uint16) []byte {
	iv := int(v)
	return append([]byte{}, byte(iv>>8), byte(iv&0xFF))
}
func encodeUint32(v uint32) []byte {
	iv := int(v)
	return append([]byte{}, byte(iv>>24), byte(iv>>16), byte(iv>>8), byte(iv&0xFF))
}

func decodeInt(buffer []byte, start int) (int, int) {
	return int(buffer[start]), start + 1
}
func decodeInt16(buffer []byte, start int) (int, int) {
	return ((int(buffer[start]) << 8) | int(buffer[start+1])), start + 2
}
func decodeInt32(buffer []byte, start int) (int, int) {
	return ((int(buffer[start]) << 24) | (int(buffer[start+1]) << 16) | (int(buffer[start+2]) << 8) | int(buffer[start+3])), start + 4
}
func decodeUint(buffer []byte, start int) (uint8, int) {
	return uint8(buffer[start]), start + 1
}
func decodeUint16(buffer []byte, start int) (uint16, int) {
	return uint16((int(buffer[start]) << 8) | int(buffer[start+1])), start + 2
}
func decodeUint32(buffer []byte, start int) (uint32, int) {
	return uint32((int(buffer[start]) << 24) | (int(buffer[start+1]) << 16) | (int(buffer[start+2]) << 8) | int(buffer[start+3])), start + 4
}

func decodeProperty(p []byte) (*Property, error) {
	var (
		i, b       int
		key, value string
		sig        uint8
	)
	prop := &Property{}

	for i < len(p) {
		sig, i = decodeUint(p, i)
		if !isPropertyTypeAvailable(sig) {
			return nil, fmt.Errorf("undefiend property byte found: %x", sig)
		}
		switch PropertyType(sig) {
		case PayloadFormatIndicator:
			prop.PayloadFormatIndicator, i = decodeUint(p, i)
		case MessageExpiryInterval:
			prop.MessageExpiryInterval, i = decodeUint32(p, i)
		case ContentType:
			prop.ContentType, i = decodeString(p, i)
		case ResponseTopic:
			prop.ResponseTopic, i = decodeString(p, i)
		case CorrelationData:
			prop.CorrelationData, i = decodeBinary(p, i)
		case SubscriptionIdentifier:
			prop.SubscriptionIdentifier, i = decodeVariable(p, i)
		case SessionExpiryInterval:
			prop.SessionExpiryInterval, i = decodeUint32(p, i)
		case AssignedClientIdentifier:
			prop.AssignedClientIdentifier, i = decodeString(p, i)
		case ServerKeepAlive:
			prop.ServerKeepAlive, i = decodeUint16(p, i)
		case AuthenticationMethod:
			prop.AuthenticationMethod, i = decodeString(p, i)
		case AuthenticationData:
			prop.AuthenticationData, i = decodeBinary(p, i)
		case RequestProblemInformation:
			b, i = decodeInt(p, i)
			prop.RequestProblemInformation = decodeBool(b)
		case WillDelayInterval:
			prop.WillDelayInterval, i = decodeUint32(p, i)
		case RequestResponseInformation:
			b, i = decodeInt(p, i)
			prop.RequestResponseInformation = decodeBool(b)
		case ResponseInformation:
			prop.ResponseInformation, i = decodeString(p, i)
		case ServerReference:
			prop.ServerReference, i = decodeString(p, i)
		case ReasonString:
			prop.ReasonString, i = decodeString(p, i)
		case ReceiveMaximum:
			prop.ReceiveMaximum, i = decodeUint16(p, i)
		case TopicAliasMaximum:
			prop.TopicAliasMaximum, i = decodeUint16(p, i)
		case TopicAlias:
			prop.TopicAlias, i = decodeUint16(p, i)
		case MaximumQoS:
			prop.MaximumQoS, i = decodeUint(p, i)
		case RetainAvalilable:
			b, i = decodeInt(p, i)
			prop.RetainAvalilable = decodeBool(b)
		case UserProperty:
			if prop.UserProperty == nil {
				prop.UserProperty = make(map[string]string)
			}
			key, i = decodeString(p, i)
			value, i = decodeString(p, i)
			prop.UserProperty[key] = value
		case MaximumPacketSize:
			prop.MaximumPacketSize, i = decodeUint32(p, i)
		case WildcardSubscriptionAvailable:
			b, i = decodeInt(p, i)
			prop.WildcardSubscriptionAvailable = decodeBool(b)
		case SubscrptionIdentifierAvailable:
			b, i = decodeInt(p, i)
			prop.SubscrptionIdentifierAvailable = decodeBool(b)
		case SharedSubscriptionsAvaliable:
			b, i = decodeInt(p, i)
			prop.SharedSubscriptionsAvaliable = decodeBool(b)
		}
	}
	return prop, nil
}

func encodeProperty(p *Property) []byte {
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

	return append(encodeVariable(len(buf)), buf...)
}
