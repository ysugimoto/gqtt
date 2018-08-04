package message

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
