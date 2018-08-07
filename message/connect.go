package message

import (
	"errors"
)

type Connect struct {
	*Frame

	// Protocol defintions
	ProtocolName    string
	ProtocolVersion uint8

	// Connection flags
	FlagUsername bool
	FlagPassword bool
	WillRetain   bool
	WillQoS      QoSLevel
	FlagWill     bool
	CleanStart   bool
	KeepAlive    uint16

	// Connection properties
	Property     *ConnectProperty
	WillProperty *WillProperty

	// Payloads
	ClientId    string
	WillTopic   string
	WillPayload string
	Username    string
	Password    []byte
}

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

func (c *ConnectProperty) ToProp() *Property {
	return &Property{
		SessionExpiryInterval:      c.SessionExpiryInterval,
		AuthenticationMethod:       c.AuthenticationMethod,
		AuthenticationData:         c.AuthenticationData,
		RequestProblemInformation:  c.RequestProblemInformation,
		RequestResponseInformation: c.RequestResponseInformation,
		ReceiveMaximum:             c.ReceiveMaximum,
		TopicAliasMaximum:          c.TopicAliasMaximum,
		UserProperty:               c.UserProperty,
		MaximumPacketSize:          c.MaximumPacketSize,
	}
}

type WillProperty struct {
	PayloadFormatIndicator uint8
	MessageExpiryInterval  uint32
	ContentType            string
	ResponseTopic          string
	CorrelationData        []byte
	WillDelayInterval      uint32
	UserProperty           map[string]string
}

func (w *WillProperty) ToProp() *Property {
	return &Property{
		PayloadFormatIndicator: w.PayloadFormatIndicator,
		MessageExpiryInterval:  w.MessageExpiryInterval,
		ContentType:            w.ContentType,
		ResponseTopic:          w.ResponseTopic,
		CorrelationData:        w.CorrelationData,
		WillDelayInterval:      w.WillDelayInterval,
		UserProperty:           w.UserProperty,
	}
}

func ParseConnect(f *Frame, p []byte) (c *Connect, err error) {
	c = &Connect{
		Frame: f,
	}

	var b int
	dec := newDecoder(p)
	if c.ProtocolName, err = dec.String(); err != nil {
		return nil, err
	}
	if c.ProtocolVersion, err = dec.Uint(); err != nil {
		return nil, err
	}
	if b, err = dec.Int(); err != nil {
		return nil, err
	}
	c.FlagUsername = ((b >> 7) & 0x01) > 0
	c.FlagPassword = ((b >> 6) & 0x01) > 0
	c.WillRetain = ((b >> 5) & 0x01) > 0
	c.WillQoS = QoSLevel(((b >> 3) & 0x03))
	c.FlagWill = ((b >> 2) & 0x01) > 0
	c.CleanStart = ((b >> 1) & 0x01) > 0

	if c.KeepAlive, err = dec.Uint16(); err != nil {
		return nil, err
	}
	// Connection variable properties enables on v5
	if prop, err := dec.Property(); err != nil {
		return nil, err
	} else if prop != nil {
		c.Property = prop.ToConnect()
	}
	if c.ClientId, err = dec.String(); err != nil {
		return nil, err
	}
	// Will properties enables on v5
	if prop, err := dec.Property(); err != nil {
		return nil, err
	} else if prop != nil {
		c.WillProperty = prop.ToWill()
	}
	if c.FlagWill {
		if c.WillTopic, err = dec.String(); err != nil {
			return nil, err
		}
		if c.WillPayload, err = dec.String(); err != nil {
			return nil, err
		}
	}
	if c.FlagUsername {
		if c.Username, err = dec.String(); err != nil {
			return nil, err
		}
	}
	if c.FlagPassword {
		if c.Password, err = dec.Binary(); err != nil {
			return nil, err
		}
	}
	return c, nil
}

func NewConnect(opts ...option) *Connect {
	return &Connect{
		Frame: newFrame(CONNECT, opts...),
	}
}

func (c *Connect) Validate() error {
	if c.ClientId == "" {
		return errors.New("clientIS must be specified")
	}
	return nil
}

func (c *Connect) Encode() ([]byte, error) {
	if err := c.Validate(); err != nil {
		return nil, err
	}

	eb := func(b bool) int {
		if b {
			return 1
		}
		return 0
	}

	enc := newEncoder()
	enc.String(c.ProtocolName)
	enc.Uint(c.ProtocolVersion)

	flag := (eb(c.FlagUsername)<<7 | eb(c.FlagPassword)<<6 | eb(c.WillRetain)<<5 | int(c.WillQoS)<<3 | eb(c.FlagWill)<<2 | eb(c.CleanStart)<<1)
	enc.Int(flag)
	enc.Uint16(c.KeepAlive)
	if c.Property != nil {
		enc.Property(c.Property.ToProp())
	} else {
		enc.Uint(0)
	}
	enc.String(c.ClientId)
	if c.WillProperty != nil {
		enc.Property(c.WillProperty.ToProp())
	} else {
		enc.Uint(0)
	}
	if c.WillTopic != "" {
		enc.String(c.WillTopic)
	}
	if c.WillPayload != "" {
		enc.String(c.WillPayload)
	}
	if c.Username != "" {
		enc.String(c.Username)
	}
	if c.Password != nil && len(c.Password) > 0 {
		enc.Binary(c.Password)
	}

	return c.Frame.Encode(enc.Get()), nil
}
