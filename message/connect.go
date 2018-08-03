package message

import (
	"errors"
)

type Connect struct {
	*Frame

	ProtocolName    string
	ProtocolVersion uint8
	FlagUsername    bool
	FlagPassword    bool
	WillRetain      bool
	WillQoS         uint8
	FlagWill        bool
	CleanSession    bool
	KeepAlive       uint16
	ClientID        string
	WillTopic       string
	WillMessage     string
	Username        string
	Password        string
}

func ParseConnect(f *Frame, p []byte) (*Connect, error) {
	c := &Connect{
		Frame: f,
	}

	var i, b int
	var j uint8
	c.ProtocolName, i = decodeString(p, i)
	c.ProtocolVersion, i = decodeInt(p, i)
	j, i = decodeInt(p, i)
	b = int(j)
	c.FlagUsername = decodeBool((b >> 7) & 0x01)
	c.FlagPassword = decodeBool((b >> 6) & 0x01)
	c.WillRetain = decodeBool((b >> 5) & 0x01)
	c.WillQoS = uint8(((b >> 3) & 0x03))
	c.FlagWill = decodeBool((b >> 2) & 0x01)
	c.CleanSession = decodeBool((b >> 1) & 0x01)
	c.KeepAlive, i = decodeInt16(p, i)
	c.ClientID, i = decodeString(p, i)
	if c.FlagWill {
		c.WillTopic, i = decodeString(p, i)
	}
	return c, nil
}

func NewConnect(f *Frame) *Connect {
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

	buf := append([]byte{}, encodeString(c.ProtocolName)...)
	buf = append(buf, encodeInt(c.ProtocolVersion))

	flagByte := encodeBool(c.FlagUsername)<<7 |
		encodeBool(c.FlagPassword)<<6 |
		encodeBool(c.WillRetain)<<5 |
		int(c.WillQoS)<<3 |
		encodeBool(c.FlagWill)<<2 |
		encodeBool(c.CleanSession)<<1
	buf = append(buf, byte(flagByte))
	buf = append(buf, encodeInt16(c.KeepAlive)...)
	buf = append(buf, encodeString(c.ClientID)...)
	if c.WillTopic != "" {
		buf = append(buf, encodeString(c.WillTopic)...)
	}
	if c.WillMessage != "" {
		buf = append(buf, encodeString(c.WillMessage)...)
	}
	if c.Username != "" {
		buf = append(buf, encodeString(c.Username)...)
	}
	if c.Password != "" {
		buf = append(buf, encodeString(c.Password)...)
	}

	return c.Frame.Encode(buf), nil
}
