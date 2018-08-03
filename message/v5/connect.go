package message

import (
	"errors"
	m "github.com/ysugimoto/gqtt/messages"
)

type Connect struct {
	*m.Frame

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
	Property     *ConnectionProperty
	WillProperty *ConnectionProperty

	// Payloads
	ClientID    string
	WillTopic   string
	WillPayload string
	Username    string
	Password    []byte
}

type ConnectionProperty struct {
	SessionExpirtInterval      uint32
	AuthenticationMethod       string
	AuthenticationData         []byte
	RequestProblemInformation  bool
	RequestResponseInformation bool
	ReceiveMaximum             uint16
	TopicAliasMaximum          uint16
	UserProperty               map[string]string
	MaximumPacketSize          uint32
}

func (c *ConnectionProperty) Encode() []byte {
	buf := make([]byte, 0)
	if c.SessionExpirtInterval > 0 {
		// buf = append(buf, m.SessionExpirtInterval.Byte())
		buf = append(buf, m.SessionExpirtInterval.Byte(), m.EncodeUint32(c.SessionExpirtInterval)...)
	}
	if c.AuthenticationMethod != "" {
		buf = append(buf, m.AuthenticationMethod.Byte())
		buf = append(buf, m.EncodeString(c.AuthenticationMethod)...)
	}
	if c.AuthenticationData != nil {
		buf = append(buf, m.AuthenticationData.Byte())
		buf = append(buf, m.EncodeBinary(c.AuthenticationData)...)
	}
	if c.RequestProblemInformation {
		buf = append(buf, m.RequestProblemInformation.Byte())
		buf = append(buf, m.EncodeInt(m.EncodeBool(c.RequestProblemInformation)))
	}
	if c.RequestResponseInformation {
		buf = append(buf, m.RequestResponseInformation.Byte())
		buf = append(buf, m.EncodeInt(m.EncodeBool(c.RequestResponseInformation)))
	}
	if c.ReceiveMaximum > 0 {
		buf = append(buf, m.ReceiveMaximum.Byte())
		buf = append(buf, m.EncodeUint16(c.ReceiveMaximum)...)
	}
	if c.TopicAliasMaximum > 0 {
		buf = append(buf, m.TopicAliasMaximum.Byte())
		buf = append(buf, m.EncodeUint16(c.TopicAliasMaximum)...)
	}
	if c.UserProperty != nil {
		for k, v := range c.UserProperty {
			buf = append(buf, m.UserProperty.Byte())
			buf = append(buf, m.EncodeString(k)...)
			buf = append(buf, m.EncodeString(v)...)
		}
	}
	if c.MaximumPacketSize > 0 {
		buf = append(buf, m.EncodeUint32(c.MaximumPacketSize)...)
	}
	return buf
}

func parseProperties(p []byte, start int) (*ConnectionProperty, []byte, int, error) {
	c := &ConnectionProperty{}
	propSize, i := m.DecodeInt(p, start)
	if propSize == 0 {
		return nil, p, i, nil
	}
	var sig uint8
	end := i + propSize
	for i < end {
		sig, i = m.DecodeUint(p, i)
		if !m.IsPropertyTypeAvailable(sig) {
			return nil, p, i, errors.New("undefiend propery type byte found")
		}
		switch m.PropertyType(sig) {
		case m.SessionExpirtInterval:
			c.SessionExpirtInterval, i = m.DecodeUint32(p, i)
		case m.AuthenticationMethod:
			c.AuthenticationMethod, i = m.DecodeString(p, i)
		case m.AuthenticationData:
			c.AuthenticationData, i = m.DecodeBinary(p, i)
		case m.RequestProblemInformation:
			var b int
			b, i = m.DecodeInt(p, i)
			c.RequestResponseInformation = m.DecodeBool(b)
		case m.RequestResponseInformation:
			var b int
			b, i = m.DecodeInt(p, i)
			c.RequestResponseInformation = m.DecodeBool(b)
		case m.ReceiveMaximum:
			c.ReceiveMaximum, i = m.DecodeUint16(p, i)
		case m.TopicAliasMaximum:
			c.TopicAliasMaximum, i = m.DecodeUint16(p, i)
		case m.UserProperty:
			if c.UserProperty == nil {
				c.UserProperty = make(map[string]string)
			}
			var key, value string
			key, i = m.DecodeString(p, i)
			value, i = m.DecodeString(p, i)
			c.UserProperty[key] = value
		case m.MaximumPacketSize:
			c.MaximumPacketSize, i = m.DecodeUint32(p, i)
		default:
			if i > end {
				return nil, p, i, errors.New("unexpected byte length detected")
			}
		}
	}
	return c, p, i, nil
}

func ParseConnect(f *m.Frame, p []byte) (*Connect, error) {
	c := &Connect{
		Frame: f,
	}

	var err error
	var i, b int
	c.ProtocolName, i = m.DecodeString(p, i)
	c.ProtocolVersion, i = m.DecodeUint(p, i)
	b, i = m.DecodeInt(p, i)
	c.FlagUsername = m.DecodeBool((b >> 7) & 0x01)
	c.FlagPassword = m.DecodeBool((b >> 6) & 0x01)
	c.WillRetain = m.DecodeBool((b >> 5) & 0x01)
	c.WillQoS = uint8(((b >> 3) & 0x03))
	c.FlagWill = m.DecodeBool((b >> 2) & 0x01)
	c.CleanStart = m.DecodeBool((b >> 1) & 0x01)
	c.KeepAlive, i = m.DecodeUint16(p, i)

	// Connection variable properties enables on v5
	c.Property, p, i, err = parseProperties(p, i)
	if err != nil {
		return nil, err
	}
	c.ClientID, i = m.DecodeString(p, i)
	// Will properties enables on v5
	c.WillProperty, p, i, err = parseProperties(p, i)
	if err != nil {
		return nil, err
	}
	if c.FlagWill {
		c.WillTopic, i = m.DecodeString(p, i)
	}
	if i < len(p) {
		c.WillPayload, i = m.DecodeString(p, i)
	}
	if i < len(p) {
		c.Username, i = m.DecodeString(p, i)
	}
	if i < len(p) {
		c.Password, i = m.DecodeBinary(p, i)
	}
	return c, nil
}

func NewConnect(f *m.Frame) *Connect {
	return &Connect{
		Frame: f,
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

	buf := append([]byte{}, m.EncodeString(c.ProtocolName)...)
	buf = append(buf, m.EncodeUint(c.ProtocolVersion))

	flagByte := m.EncodeBool(c.FlagUsername)<<7 |
		m.EncodeBool(c.FlagPassword)<<6 |
		m.EncodeBool(c.WillRetain)<<5 |
		int(c.WillQoS)<<3 |
		m.EncodeBool(c.FlagWill)<<2 |
		m.EncodeBool(c.CleanStart)<<1
	buf = append(buf, byte(flagByte))
	buf = append(buf, m.EncodeUint16(c.KeepAlive)...)
	if c.Property != nil {
		buf = append(buf, c.Property.Encode()...)
	}
	buf = append(buf, m.EncodeString(c.ClientID)...)
	if c.WillProperty != nil {
		buf = append(buf, c.WillProperty.Encode()...)
	}
	if c.WillTopic != "" {
		buf = append(buf, m.EncodeString(c.WillTopic)...)
	}
	if c.WillPayload != "" {
		buf = append(buf, m.EncodeString(c.WillPayload)...)
	}
	if c.Username != "" {
		buf = append(buf, m.EncodeString(c.Username)...)
	}
	if c.Password != nil && len(c.Password) > 0 {
		buf = append(buf, m.EncodeBinary(c.Password)...)
	}

	return c.Frame.Encode(buf), nil
}
