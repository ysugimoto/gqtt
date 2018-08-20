package message

//go:generate stringer -type=MessageType
type MessageType uint8

//go:generate stringer -type=ReasonCode
type ReasonCode uint8

type (
	PropertyType uint8
	QoSLevel     uint8
	Encoder      interface {
		Duplicate()
		GetType() MessageType
		Encode() ([]byte, error)
	}
)

const (
	QoS0 QoSLevel = iota
	QoS1
	QoS2
)

const (
	_ MessageType = iota
	CONNECT
	CONNACK
	PUBLISH
	PUBACK
	PUBREC
	PUBREL
	PUBCOMP
	SUBSCRIBE
	SUBACK
	UNSUBSCRIBE
	UNSUBACK
	PINGREQ
	PINGRESP
	DISCONNECT
	AUTH
)

// func (m MessageType) String() string {
// 	switch m {
// 	case CONNECT:
// 		return "CONNECT"
// 	case CONNACK:
// 		return "CONNACK"
// 	case PUBLISH:
// 		return "PUBLISH"
// 	case PUBACK:
// 		return "PUBACK"
// 	case PUBREC:
// 		return "PUBREC"
// 	case PUBREL:
// 		return "PUBREL"
// 	case PUBCOMP:
// 		return "PUBCOMP"
// 	case SUBSCRIBE:
// 		return "SUBSCRIBE"
// 	case SUBACK:
// 		return "SUBACK"
// 	case UNSUBSCRIBE:
// 		return "UNSUBSCRIBE"
// 	case UNSUBACK:
// 		return "UNSUBACK"
// 	case PINGREQ:
// 		return "PINGREQ"
// 	case PINGRESP:
// 		return "PINGRESP"
// 	case DISCONNECT:
// 		return "DISCONNECT"
// 	case AUTH:
// 		return "AUTH"
// 	default:
// 		return "UNDEFINED"
// 	}
// }

const (
	PayloadFormatIndicator         PropertyType = 0x01
	MessageExpiryInterval          PropertyType = 0x02
	ContentType                    PropertyType = 0x03
	ResponseTopic                  PropertyType = 0x08
	CorrelationData                PropertyType = 0x09
	SubscriptionIdentifier         PropertyType = 0x0B
	SessionExpiryInterval          PropertyType = 0x11
	AssignedClientIdentifier       PropertyType = 0x12
	ServerKeepAlive                PropertyType = 0x13
	AuthenticationMethod           PropertyType = 0x15
	AuthenticationData             PropertyType = 0x16
	RequestProblemInformation      PropertyType = 0x17
	WillDelayInterval              PropertyType = 0x18
	RequestResponseInformation     PropertyType = 0x19
	ResponseInformation            PropertyType = 0x1A
	ServerReference                PropertyType = 0x1C
	ReasonString                   PropertyType = 0x1F
	ReceiveMaximum                 PropertyType = 0x21
	TopicAliasMaximum              PropertyType = 0x22
	TopicAlias                     PropertyType = 0x23
	MaximumQoS                     PropertyType = 0x24
	RetainAvalilable               PropertyType = 0x25
	UserProperty                   PropertyType = 0x26
	MaximumPacketSize              PropertyType = 0x27
	WildcardSubscriptionAvailable  PropertyType = 0x28
	SubscrptionIdentifierAvailable PropertyType = 0x29
	SharedSubscriptionsAvaliable   PropertyType = 0x2A
)

const (
	Success                             ReasonCode = 0x00
	NormalDisconnection                 ReasonCode = 0x00
	GrantedQoS0                         ReasonCode = 0x00
	GrantedQoS1                         ReasonCode = 0x01
	GrantedQoS2                         ReasonCode = 0x02
	DisconnectWithWillMessage           ReasonCode = 0x04
	NoMatchingSubscribers               ReasonCode = 0x10
	NoSubscriptionExisted               ReasonCode = 0x11
	ContinueAuthentication              ReasonCode = 0x18
	ReAuthenticate                      ReasonCode = 0x19
	UnspecifiedError                    ReasonCode = 0x80
	MalformedPacket                     ReasonCode = 0x81
	ProtocolError                       ReasonCode = 0x82
	ImplementationSpecificError         ReasonCode = 0x83
	UnsupportedProtocolVersion          ReasonCode = 0x84
	ClientIdentifierNotValid            ReasonCode = 0x85
	BadUsernameOrPassword               ReasonCode = 0x86
	NotAuthorized                       ReasonCode = 0x87
	ServerUnavailable                   ReasonCode = 0x88
	ServerBusy                          ReasonCode = 0x89
	Banned                              ReasonCode = 0x8A
	ServerShuttingDown                  ReasonCode = 0x8B
	BadAuthenticationMethod             ReasonCode = 0x8C
	KeepAliveTimeout                    ReasonCode = 0x8D
	SessionTakenOver                    ReasonCode = 0x8E
	TopicFilterInvalid                  ReasonCode = 0x8F
	TopicNameInvalid                    ReasonCode = 0x90
	PacketIdentifierInUse               ReasonCode = 0x91
	PacketIdentifierNotFound            ReasonCode = 0x92
	ReceiveMaximumExceeded              ReasonCode = 0x93
	TopicAliasInvalid                   ReasonCode = 0x94
	PacketTooLarge                      ReasonCode = 0x95
	MessageRateTooHigh                  ReasonCode = 0x96
	QuotaExceeded                       ReasonCode = 0x97
	AdministrativeAction                ReasonCode = 0x98
	PayloadFormatInvalid                ReasonCode = 0x99
	RetianlNotSupported                 ReasonCode = 0x9A
	QoSNotSupported                     ReasonCode = 0x9B
	UseAnotherServer                    ReasonCode = 0x9C
	ServerMoved                         ReasonCode = 0x9D
	SharedSubscriptionsNotSupported     ReasonCode = 0x9E
	ConnectionRateExceeded              ReasonCode = 0x9F
	MaximumConnectionTime               ReasonCode = 0xA0
	SubscriptionIdentifiersNotSupported ReasonCode = 0xA1
	WildcardSubscriptionsNotSupported   ReasonCode = 0xA2
)

func IsQoSAvaliable(v uint8) bool {
	if v >= 0 && v <= 2 {
		return true
	}
	return false
}

func IsMessageTypeAvailable(v uint8) bool {
	if v > 0 && v < 16 {
		return true
	}
	return false
}

func (p PropertyType) Byte() byte {
	return byte(p)
}

func isPropertyTypeAvailable(v uint8) bool {
	if (v >= 0x01 && v <= 0x03) ||
		(v >= 0x08 && v <= 0x09) ||
		v == 0x0B ||
		(v >= 0x11 && v <= 0x13) ||
		(v >= 0x15 && v <= 0x1A) ||
		v == 0x1C ||
		v == 0x1F ||
		(v >= 0x21 && v <= 0x2A) {
		return true
	}
	return false
}

func (r ReasonCode) Byte() byte {
	return byte(r)
}

func IsReasonCodeAvailable(v uint8) bool {
	if (v >= 0x00 && v <= 0x02) ||
		v == 0x04 ||
		(v >= 0x10 && v <= 0x11) ||
		(v >= 0x18 && v <= 0x19) ||
		(v >= 0x80 && v <= 0xA2) {
		return true
	}
	return false
}
