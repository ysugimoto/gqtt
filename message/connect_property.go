package message

type ConnectProperty struct {
	SessionExpiryInterval      uint32
	AuthenticationMethod       string
	AuthenticationData         []byte
	RequestProblemInformation  bool
	RequestResponseInformation bool
	ReceiveMaximum             uint16
	TopicAliasMaximum          uint16
	UserProperty               map[string]string
	MaximumPacketSize          uint32
}

func (c *ConnectProperty) Encode() []byte {
	return encodeProperty(&Property{
		SessionExpiryInterval:      c.SessionExpiryInterval,
		AuthenticationMethod:       c.AuthenticationMethod,
		AuthenticationData:         c.AuthenticationData,
		RequestProblemInformation:  c.RequestProblemInformation,
		RequestResponseInformation: c.RequestResponseInformation,
		ReceiveMaximum:             c.ReceiveMaximum,
		TopicAliasMaximum:          c.TopicAliasMaximum,
		UserProperty:               c.UserProperty,
		MaximumPacketSize:          c.MaximumPacketSize,
	})
}
