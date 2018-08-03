package message

func EncodeBool(b bool) (i int) {
	if b {
		i = 1
	}
	return
}

func DecodeBool(i int) (b bool) {
	if i > 0 {
		b = true
	}
	return
}

func EncodeString(str string) []byte {
	buf := []byte(str)
	size := len(buf)

	return append(EncodeInt16(size), buf...)
}
func EncodeBinary(b []byte) []byte {
	size := len(b)
	return append(EncodeInt16(size), b...)
}

func DecodeString(buffer []byte, start int) (string, int) {
	size := ((int(buffer[start]) << 8) | int(buffer[start+1]))
	start += 2
	data := string(buffer[start:(start + size)])
	return data, start + size
}
func DecodeBinary(buffer []byte, start int) ([]byte, int) {
	size := ((int(buffer[start]) << 8) | int(buffer[start+1]))
	start += 2
	data := buffer[start:(start + size)]
	return data, start + size
}

func EncodeInt(v int) byte {
	return byte(v)
}
func EncodeInt16(v int) []byte {
	return append([]byte{}, byte(v>>8), byte(v&0xFF))
}
func EncodeInt32(v int) []byte {
	return append([]byte{}, byte(v>>24), byte(v>>16), byte(v>>8), byte(v&0xFF))
}
func EncodeUint(v uint8) byte {
	return byte(v)
}
func EncodeUint16(v uint16) []byte {
	iv := int(v)
	return append([]byte{}, byte(iv>>8), byte(iv&0xFF))
}
func EncodeUint32(v uint32) []byte {
	iv := int(v)
	return append([]byte{}, byte(iv>>24), byte(iv>>16), byte(iv>>8), byte(iv&0xFF))
}

func DecodeInt(buffer []byte, start int) (int, int) {
	return int(buffer[start]), start + 1
}
func DecodeInt16(buffer []byte, start int) (int, int) {
	return ((int(buffer[start]) << 8) | int(buffer[start+1])), start + 2
}
func DecodeInt32(buffer []byte, start int) (int, int) {
	return ((int(buffer[start]) << 24) | (int(buffer[start]) << 16) | (int(buffer[start]) << 8) | int(buffer[start+1])), start + 4
}
func DecodeUint(buffer []byte, start int) (uint8, int) {
	return uint8(buffer[start]), start + 1
}
func DecodeUint16(buffer []byte, start int) (uint16, int) {
	return uint16((int(buffer[start]) << 8) | int(buffer[start+1])), start + 2
}
func DecodeUint32(buffer []byte, start int) (uint32, int) {
	return uint32((int(buffer[start]) << 24) | (int(buffer[start]) << 16) | (int(buffer[start]) << 8) | int(buffer[start+1])), start + 4
}
