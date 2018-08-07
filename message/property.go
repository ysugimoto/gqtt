package message

type Property struct {
	PayloadFormatIndicator         uint8
	MessageExpiryInterval          uint32
	ContentType                    string
	ResponseTopic                  string
	CorrelationData                []byte
	SubscriptionIdentifier         uint64
	SessionExpiryInterval          uint32
	AssignedClientIdentifier       string
	ServerKeepAlive                uint16
	AuthenticationMethod           string
	AuthenticationData             []byte
	RequestProblemInformation      bool
	WillDelayInterval              uint32
	RequestResponseInformation     bool
	ResponseInformation            string
	ServerReference                string
	ReasonString                   string
	ReceiveMaximum                 uint16
	TopicAliasMaximum              uint16
	TopicAlias                     uint16
	MaximumQoS                     uint8
	RetainAvalilable               bool
	UserProperty                   map[string]string
	MaximumPacketSize              uint32
	WildcardSubscriptionAvailable  bool
	SubscrptionIdentifierAvailable bool
	SharedSubscriptionsAvaliable   bool
}

func (p *Property) ToWill() *WillProperty {
	return &WillProperty{
		PayloadFormatIndicator: p.PayloadFormatIndicator,
		MessageExpiryInterval:  p.MessageExpiryInterval,
		ContentType:            p.ContentType,
		ResponseTopic:          p.ResponseTopic,
		CorrelationData:        p.CorrelationData,
		WillDelayInterval:      p.WillDelayInterval,
		UserProperty:           p.UserProperty,
	}
}

func (p *Property) ToPublish() *PublishProperty {
	return &PublishProperty{
		PayloadFormatIndicator: p.PayloadFormatIndicator,
		MessageExpiryInterval:  p.MessageExpiryInterval,
		ContentType:            p.ContentType,
		ResponseTopic:          p.ResponseTopic,
		CorrelationData:        p.CorrelationData,
		SubscriptionIdentifier: p.SubscriptionIdentifier,
		TopicAlias:             p.TopicAlias,
		UserProperty:           p.UserProperty,
	}
}

func (p *Property) ToSubscribe() *SubscribeProperty {
	return &SubscribeProperty{
		SubscriptionIdentifier: p.SubscriptionIdentifier,
		UserProperty:           p.UserProperty,
	}
}

func (p *Property) ToConnect() *ConnectProperty {
	return &ConnectProperty{
		SessionExpiryInterval:      p.SessionExpiryInterval,
		AuthenticationMethod:       p.AuthenticationMethod,
		AuthenticationData:         p.AuthenticationData,
		RequestProblemInformation:  p.RequestProblemInformation,
		RequestResponseInformation: p.RequestResponseInformation,
		ReceiveMaximum:             p.ReceiveMaximum,
		TopicAliasMaximum:          p.TopicAliasMaximum,
		UserProperty:               p.UserProperty,
		MaximumPacketSize:          p.MaximumPacketSize,
	}
}

func (p *Property) ToConnAck() *ConnAckProperty {
	return &ConnAckProperty{
		SessionExpiryInterval:          p.SessionExpiryInterval,
		AssignedClientIdentifier:       p.AssignedClientIdentifier,
		ServerKeepAlive:                p.ServerKeepAlive,
		AuthenticationMethod:           p.AuthenticationMethod,
		AuthenticationData:             p.AuthenticationData,
		ResponseInformation:            p.ResponseInformation,
		ServerReference:                p.ServerReference,
		ReasonString:                   p.ReasonString,
		ReceiveMaximum:                 p.ReceiveMaximum,
		TopicAliasMaximum:              p.TopicAliasMaximum,
		MaximumQoS:                     p.MaximumQoS,
		RetainAvalilable:               p.RetainAvalilable,
		UserProperty:                   p.UserProperty,
		MaximumPacketSize:              p.MaximumPacketSize,
		WildcardSubscriptionAvailable:  p.WildcardSubscriptionAvailable,
		SubscrptionIdentifierAvailable: p.SubscrptionIdentifierAvailable,
		SharedSubscriptionsAvaliable:   p.SharedSubscriptionsAvaliable,
	}
}

func (p *Property) ToDisconnect() *DisconnectProperty {
	return &DisconnectProperty{
		SessionExpiryInterval: p.SessionExpiryInterval,
		ServerReference:       p.ServerReference,
		ReasonString:          p.ReasonString,
		UserProperty:          p.UserProperty,
	}
}

func (p *Property) ToAuth() *AuthProperty {
	return &AuthProperty{
		AuthenticationMethod: p.AuthenticationMethod,
		AuthenticationData:   p.AuthenticationData,
		ReasonString:         p.ReasonString,
		UserProperty:         p.UserProperty,
	}
}

func (p *Property) ToPubAck() *PubAckProperty {
	return &PubAckProperty{
		ReasonString: p.ReasonString,
		UserProperty: p.UserProperty,
	}
}

func (p *Property) ToPubRec() *PubRecProperty {
	return &PubRecProperty{
		ReasonString: p.ReasonString,
		UserProperty: p.UserProperty,
	}
}

func (p *Property) ToPubRel() *PubRelProperty {
	return &PubRelProperty{
		ReasonString: p.ReasonString,
		UserProperty: p.UserProperty,
	}
}

func (p *Property) ToPubComp() *PubCompProperty {
	return &PubCompProperty{
		ReasonString: p.ReasonString,
		UserProperty: p.UserProperty,
	}
}

func (p *Property) ToSubAck() *SubAckProperty {
	return &SubAckProperty{
		ReasonString: p.ReasonString,
		UserProperty: p.UserProperty,
	}
}

func (p *Property) ToUnsubscribe() *UnsubscribeProperty {
	return &UnsubscribeProperty{
		UserProperty: p.UserProperty,
	}
}

func (p *Property) ToUnsubAck() *UnsubAckProperty {
	return &UnsubAckProperty{
		ReasonString: p.ReasonString,
		UserProperty: p.UserProperty,
	}
}
