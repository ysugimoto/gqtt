package message

func encodeBool(b bool) (i int) {
	if b {
		i = 1
	}
	return
}

func decodeBool(i int) (b bool) {
	if i > 0 {
		b = true
	}
	return
}

func encodeString(str string) []byte {
	buf := []byte(str)
	size := len(buf)

	return append(encodeInt(size), buf...)
}

func decodeString(buffer []byte, start int) (string, int) {
	size := ((int(buffer[start]) << 8) | int(buffer[start+1]))
	start += 2
	data := string(buffer[start:(start + size)])
	return data, start + size
}

func encodeInt(v int) []byte {
	return append([]byte{}, byte(v>>8), byte(v&0xFF))
}

func decodeInt(buffer []byte, start int) (int, int) {
	return ((int(buffer[start]) << 8) | int(buffer[start+1])), start + 2
}
