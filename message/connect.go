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

	var size, i int
	size = ((int(p[i]) << 8) | int(p[i+1]))
	i += 2
	c.ProtocolName = string(p[i:(i + size)])
	i += size
	c.ProtocolVersion = p[i]
	i++
	b := int(p[i])
	c.FlagUsername = toBool((b >> 7) & 0x01)
	c.FlagPassword = toBool((b >> 6) & 0x01)
	c.WillRetain = toBool((b >> 5) & 0x01)
	c.WillQoS = uint8((b >> 3) & 0x03)
	c.FlagWill = toBool((b >> 2) & 0x01)
	c.CleanSession = toBool((b >> 1) & 0x01)
	i++

	c.KeepAlive = uint16(((int(p[i]) << 8) | int(p[i+1])))
	i += 2
	size = ((int(p[i]) << 8) | int(p[i+1]))
	i += 2
	c.ClientID = string(p[i:(i + size)])
	i += size
	if c.FlagWill {
		size = ((int(p[i]) << 8) | int(p[i+1]))
		i += 2
		c.WillTopic = string(p[i:(i + size)])
		i += size
		size = ((int(p[i]) << 8) | int(p[i+1]))
		i += 2
		c.WillTopic = string(p[i:(i + size)])
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

	buf := make([]byte, 0)
	var size int

	size = len([]byte(c.ProtocolName))
	buf = append(buf, byte(size>>8), byte(size&0xFF))
	buf = append(buf, []byte(c.ProtocolName)...)
	buf = append(buf, byte(int(c.ProtocolVersion)))

	flagByte := fromBool(c.FlagUsername)<<7 |
		fromBool(c.FlagPassword)<<6 |
		fromBool(c.WillRetain)<<5 |
		int(c.WillQoS)<<3 |
		fromBool(c.FlagWill)<<2 |
		fromBool(c.CleanSession)<<1
	buf = append(buf, byte(flagByte))

	buf = append(buf, byte(c.KeepAlive>>8), byte(c.KeepAlive&0xFF))
	size = len([]byte(c.ClientID))
	buf = append(buf, byte(size>>8), byte(size&0xFF))
	buf = append(buf, []byte(c.ClientID)...)
	if c.WillTopic != "" {
		size = len([]byte(c.WillTopic))
		buf = append(buf, byte(size>>8), byte(size&0xFF))
		buf = append(buf, []byte(c.WillTopic)...)
	}
	if c.WillMessage != "" {
		size = len([]byte(c.WillMessage))
		buf = append(buf, byte(size>>8), byte(size&0xFF))
		buf = append(buf, []byte(c.WillMessage)...)
	}
	if c.Username != "" {
		size = len([]byte(c.Username))
		buf = append(buf, byte(size>>8), byte(size&0xFF))
		buf = append(buf, []byte(c.Username)...)
	}
	if c.Password != "" {
		size = len([]byte(c.Password))
		buf = append(buf, byte(size>>8), byte(size&0xFF))
		buf = append(buf, []byte(c.Password)...)
	}

	return c.Frame.Encode(buf), nil
}
