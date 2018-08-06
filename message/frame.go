package message

import (
	"bufio"
	"fmt"
	"io"
)

type Frame struct {
	Type   MessageType
	DUP    bool
	QoS    QoSLevel
	RETAIN bool
	Size   uint64
}

func newFrame(mt MessageType, options ...option) *Frame {
	f := &Frame{
		Type: mt,
	}
	for _, o := range options {
		switch o.name {
		case optionNameRetain:
			f.RETAIN = o.value.(bool)
		case optionNameQoS:
			f.QoS = o.value.(QoSLevel)
		}
	}
	return f
}

func (f *Frame) SetQoS(qos QoSLevel) {
	f.QoS = qos
}
func (f *Frame) SetRetain(retain bool) {
	f.RETAIN = retain
}
func (f *Frame) Duplicate() {
	f.DUP = true
}

func (f *Frame) Encode(payload []byte) []byte {
	header := []byte{byte(int(f.Type<<4) | encodeBool(f.DUP)<<3 | int(f.QoS)<<1 | encodeBool(f.RETAIN))}
	varHeader := []byte{0}
	if len(payload) > 0 {
		varHeader = encodeVariable(len(payload))
	}
	header = append(header, varHeader...)

	return append(header, payload...)
}

func ReceiveFrame(r io.Reader) (*Frame, []byte, error) {
	var packet byte
	var err error
	var size uint64

	reader := bufio.NewReader(r)

	// Read and extract first byte
	packet, err = reader.ReadByte()
	if err != nil {
		return nil, nil, err
	}
	b := int(packet)
	f := &Frame{
		Type:   MessageType((b >> 4) & 0x0F),
		DUP:    decodeBool(((b >> 3) & 0x01)),
		QoS:    QoSLevel(((b >> 1) & 0x03)),
		RETAIN: decodeBool((b & 0x01)),
	}
	if !IsQoSAvaliable(uint8(f.QoS)) {
		return nil, nil, fmt.Errorf("invalid QoS level specified: %x", f.QoS)
	}

	// Read variable remain length
	var mul uint64 = 1
	for {
		if packet, err = reader.ReadByte(); err != nil {
			return nil, nil, err
		}
		size += uint64(packet&0x7F) * mul
		mul *= 0x80
		if packet&0x80 == 0 {
			break
		}
	}
	f.Size = size
	payload := make([]byte, size)
	// There are case that length is zero on PINGREQ, PINGRESP
	if _, err = reader.Read(payload); err != nil {
		return f, nil, err
	}
	return f, payload, nil
}
