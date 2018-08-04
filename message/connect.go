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
	WillQoS      uint8
	FlagWill     bool
	CleanStart   bool
	KeepAlive    uint16

	// Connection properties
	Property     *ConnectProperty
	WillProperty *WillProperty

	// Payloads
	ClientID    string
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

func ParseConnect(f *Frame, p []byte) (*Connect, error) {
	c := &Connect{
		Frame: f,
	}

	var i, b, psize int
	c.ProtocolName, i = decodeString(p, i)
	c.ProtocolVersion, i = decodeUint(p, i)
	b, i = decodeInt(p, i)
	c.FlagUsername = decodeBool((b >> 7) & 0x01)
	c.FlagPassword = decodeBool((b >> 6) & 0x01)
	c.WillRetain = decodeBool((b >> 5) & 0x01)
	c.WillQoS = uint8(((b >> 3) & 0x03))
	c.FlagWill = decodeBool((b >> 2) & 0x01)
	c.CleanStart = decodeBool((b >> 1) & 0x01)
	c.KeepAlive, i = decodeUint16(p, i)

	// Connection variable properties enables on v5
	psize, i = decodeInt(p, i)
	if psize > 0 {
		prop, err := decodeProperty(p[i:(i + psize)])
		if err != nil {
			return nil, err
		}
		c.Property = prop.ToConnect()
		i += psize
	}
	c.ClientID, i = decodeString(p, i)
	// Will properties enables on v5
	psize, i = decodeInt(p, i)
	if psize > 0 {
		prop, err := decodeProperty(p[i:(i + psize)])
		if err != nil {
			return nil, err
		}
		c.WillProperty = prop.ToWill()
		i += psize
	}
	if c.FlagWill {
		c.WillTopic, i = decodeString(p, i)
		c.WillPayload, i = decodeString(p, i)
	}
	if c.FlagUsername {
		c.Username, i = decodeString(p, i)
	}
	if c.FlagPassword {
		c.Password, i = decodeBinary(p, i)
	}
	return c, nil
}

func NewConnect(opts ...option) *Connect {
	return &Connect{
		Frame: newFrame(CONNECT, opts...),
	}
}

func (c *Connect) Validate() error {
	if c.ClientID == "" {
		return errors.New("clientIS must be specified")
	}
	return nil
}

func (c *Connect) Encode() ([]byte, error) {
	if err := c.Validate(); err != nil {
		return nil, err
	}

	buf := append([]byte{}, encodeString(c.ProtocolName)...)
	buf = append(buf, encodeUint(c.ProtocolVersion))

	flagByte := encodeBool(c.FlagUsername)<<7 |
		encodeBool(c.FlagPassword)<<6 |
		encodeBool(c.WillRetain)<<5 |
		int(c.WillQoS)<<3 |
		encodeBool(c.FlagWill)<<2 |
		encodeBool(c.CleanStart)<<1
	buf = append(buf, byte(flagByte))
	buf = append(buf, encodeUint16(c.KeepAlive)...)
	if c.Property != nil {
		buf = append(buf, c.Property.Encode()...)
	} else {
		buf = append(buf, byte(0))
	}
	buf = append(buf, encodeString(c.ClientID)...)
	if c.WillProperty != nil {
		buf = append(buf, c.WillProperty.Encode()...)
	} else {
		buf = append(buf, byte(0))
	}
	if c.WillTopic != "" {
		buf = append(buf, encodeString(c.WillTopic)...)
	}
	if c.WillPayload != "" {
		buf = append(buf, encodeString(c.WillPayload)...)
	}
	if c.Username != "" {
		buf = append(buf, encodeString(c.Username)...)
	}
	if c.Password != nil && len(c.Password) > 0 {
		buf = append(buf, encodeBinary(c.Password)...)
	}

	return c.Frame.Encode(buf), nil
}
