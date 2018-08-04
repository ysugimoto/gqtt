package message

type PublishProperty struct {
	PayloadFormatIndicator uint8
	MessageExpiryInterval  uint32
	ContentType            string
	ResponseTopic          string
	CorrelationData        []byte
	SubscriptionIdentifier uint64
	TopicAlias             uint16
	UserProperty           map[string]string
}

func (p *PublishProperty) Encode() []byte {
	return encodeProperty(&Property{
		PayloadFormatIndicator: p.PayloadFormatIndicator,
		MessageExpiryInterval:  p.MessageExpiryInterval,
		ContentType:            p.ContentType,
		ResponseTopic:          p.ResponseTopic,
		CorrelationData:        p.CorrelationData,
		SubscriptionIdentifier: p.SubscriptionIdentifier,
		TopicAlias:             p.TopicAlias,
		UserProperty:           p.UserProperty,
	})
}
