package main

import "fmt"

func main() {
	message := []byte{16, 23, 0, 6, 77, 81, 73, 115, 100, 112, 3, 2, 0, 30, 0, 9, 103, 111, 45, 115, 105, 109, 112, 108, 10}
	// message := []byte{16, 21, 0, 4, 77, 81, 84, 84, 4, 2, 0, 30, 0, 9, 103, 111, 45, 115, 105, 109, 112, 108, 101}
	message = parseFixedHeader(message)
	message = parseRemaininLength(message)
	message = parseVariableHeaders(message)
	for _, v := range message {
		fmt.Println(v, string(v))
	}
}

func parseFixedHeader(m []byte) []byte {
	h := m[0]
	fmt.Println("==== parse fixed header")
	fmt.Printf("message type: %04b\n", (h>>4)&0x0F)
	fmt.Printf("DUP: %b\n", (h>>3)&0x01)
	fmt.Printf("QoS: %02b\n", (h>>1)&0x03)
	fmt.Printf("RETAIN: %b\n", (h & 0x01))
	return m[1:]
}

func parseRemaininLength(m []byte) []byte {
	mul := 1
	length := 0
	i := 0
	for ; ; i++ {
		h := m[i]
		length += int(h&0x7F) * mul
		mul *= 0x80
		if h&0x80 == 0 {
			break
		}
	}

	fmt.Println("==== parse remaining length")
	fmt.Printf("remaining length: %d, actual: %d\n", length, len(m[i+1:]))

	return m[i+1:]
}

func parseVariableHeaders(m []byte) []byte {
	fmt.Println("==== parse variable headers")
	fmt.Printf("protocol name MSB: %d\n", m[0])
	fmt.Printf("protocol name LSB: %d\n", m[1])
	for _, v := range m[2 : 2+m[1]] {
		fmt.Println(v, string(v))
	}
	fmt.Printf("protocol name: %s\n", string(m[2:2+6]))
	fmt.Printf("protocol version: %d\n", m[8])

	b := m[9]
	fmt.Printf("username flag: %b\n", (b>>7)&0x01)
	fmt.Printf("password flag: %b\n", (b>>6)&0x01)
	fmt.Printf("will retain: %b\n", (b>>5)&0x01)
	fmt.Printf("will QoS: %02b\n", (b>>3)&0x03)
	fmt.Printf("will flag: %b\n", (b>>2)&0x01)
	fmt.Printf("clean session: %b\n", (b>>1)&0x01)
	fmt.Printf("reserved: %b\n", (b & 0x01))

	fmt.Printf("keep alive MSB: %d\n", m[10])
	fmt.Printf("keep alive LSB: %d\n", m[11])
	return m[12:]
}
