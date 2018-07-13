package broker

import (
	"bufio"
	"fmt"
	"log"
	"net"

	"github.com/ysugimoto/gqtt/message"
)

type Broker struct {
}

func (b *Broker) Run(port int) error {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return err
	}

	defer listener.Close()
	log.Println("Broker server started.")
	for {
		s, err := listener.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		if err := handshake(s); err != nil {
			log.Println("Handshake failed. Close connection")
			s.Close()
			continue
		}
		log.Println("Handshake success!")
		s.Close()
		// go handle(s)
	}
}

func handshake(s net.Conn) error {
	f, p, err := message.ReceiveFrame(s)
	if err != nil {
		log.Println("receive frame error: ", err)
		return err
	}
	c, err := message.ParseConnect(f, p)
	if err != nil {
		log.Println("frame expects connect package: ", err)
		return err
	}
	log.Printf("CONNECT accepted: %+v\n", c)
	w := bufio.NewWriter(s)
	ack := message.NewConnAck(&message.Frame{
		Type: message.CONNACK,
	})
	if buf, err := ack.Encode(); err != nil {
		log.Println("CONNACK encode error: ", err)
		return err
	} else if _, err := w.Write(buf); err != nil {
		log.Println("CONNACK write error: ", err)
		return err
	}
	if err := w.Flush(); err != nil {
		log.Println("CONNACK flush error: ", err)
		return err
	}

	return nil
}
