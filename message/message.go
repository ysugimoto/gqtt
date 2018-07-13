package message

import (
	"bufio"
	"io"
)

type MessageType uint8

const (
	reserved0 MessageType = iota
	CONNECT
	CONNACK
	PUBLISH
	PUBACK
	PUBREC
	PUBREL
	PUBCOMP
	SUBSCRIBE
	SUBACK
	UNSUBSCRIBE
	UNSUBACK
	PINGREQ
	PINGRESP
	DISCONNECT
	reserved15
)

const (
	QoS0 uint8 = iota
	QoS1
	QoS2
)

type Frame struct {
	Type   MessageType
	DUP    bool
	QoS    uint8
	RETAIN bool
	Size   uint64
}

func (f *Frame) Encode(payload []byte) []byte {
	header := []byte{byte(int(f.Type<<4) | fromBool(f.DUP)<<3 | int(f.QoS)<<1 | fromBool(f.RETAIN))}

	size := len(payload)
	if size == 0 {
		header = append(header, 0)
	} else {
		for size > 0 {
			digit := size % 0x80
			size /= 0x80
			if size > 0 {
				digit |= 0x80
			}
			header = append(header, byte(digit))
		}
	}

	return append(header, payload...)
}

func (f *Frame) Duplicate() {
	f.DUP = true
}

func toBool(i int) (b bool) {
	if i > 0 {
		b = true
	}
	return
}

func fromBool(b bool) (i int) {
	if b {
		i = 1
	}
	return
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
		DUP:    toBool(((b >> 3) & 0x01)),
		QoS:    uint8(((b >> 1) & 0x03)),
		RETAIN: toBool((b & 0x01)),
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
