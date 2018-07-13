package main

import (
	"bufio"
	"fmt"
	"net"
	"time"
)

func main() {
	listener, err := net.Listen("tcp", "localhost:9999")
	if err != nil {
		panic(err)
	}
	defer listener.Close()
	fmt.Println("TCP server started at localhost:9999")

	for {
		c, err := listener.Accept()
		if err != nil {
			fmt.Println("accept error: ", err)
			continue
		}
		go func() {
			c.SetReadDeadline(time.Now().Add(10 * time.Second))
			defer c.Close()
			s := bufio.NewScanner(bufio.NewReader(c))
			for s.Scan() {
				fmt.Println(s.Bytes())
			}
		}()
	}
}
