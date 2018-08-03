package broker

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net"
	"strings"
	"time"

	"github.com/satori/go.uuid"
	"github.com/ysugimoto/gqtt/message"
)

func formatTopicPath(path string) string {
	return "/" + strings.Trim(path, "/")
}

type Broker struct {
	port int
}

func NewBroker(port int) {
	return &Broker{
		port: port,
	}
}

func (b *Broker) Run(ctx context.Context) error {
	c, cancel := context.WithCancel(ctx)
	defer cancel()

	manager = initManager(c)

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", b.port))
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

		go func(conn net.Conn) {
			c, cancel := context.WithCancel(ctx)
			defer cancel()

			client := NewClient(conn, c, uuid.NewV4().String())
			if err := client.Handshake(10 * time.Second); err != nil {
				log.Printf("failed to MQTT handshake for %s, %s", client.Id(), err.Error())
				client.Close()
				return
			}
			for {
				f, p, err := message.ReceiveFrame(c)
				if err != nil {
					log.Println("receive frame error: ", err)
					return
				}
				switch f.Type {
				case message.PUBLISH:
					pb, err := message.ParsePublish(f, p)
					if err != nil {
						log.Printf("failed to parse packet to PUBLISH: %s\n", err.Error())
						continue
					}
					pb.ClientId = client.Id()
					manager.publish <- pb
				case message.SUBSCRIBE:
					ss, err := message.ParseSubscribe(f, p)
					if err != nil {
						log.Printf("failed to parse packet to SUBSCRIBE: %s\n", err.Error())
						continue
					}
					ss.ClientId = client.Id()
					manager.subscribe <- ss
				case message.UNSUBSCRIBE:
					uss, err := message.ParseUnsubscribe(f, p)
					if err != nil {
						log.Printf("failed to parse packet to UNSUBSCRIBE: %s\n", err.Error())
						continue
					}
					uss.ClientId = client.Id()
					manager.unsubscribe <- uss
				case message.PINGREQ:
					_, err := message.ParsePingReq(f, p)
					if err != nil {
						log.Printf("failed to parse packet to PINGREQ: %s\n", err.Error())
						continue
					}
					m.ping <- c
				case message.DISCONNECT:
					_, err := message.ParseDisconnect(f, p)
					if err != nil {
						log.Printf("failed to parse packet to DISCONNECT: %s\n", err.Error())
						continue
					}
					m.disconnect <- c
				default:
					log.Printf("unexpected message type: %d", f.Type)
				}
			}
		}(s)
	}
}
