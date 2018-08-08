package broker

import (
	"context"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/ysugimoto/gqtt/internal/log"
	"github.com/ysugimoto/gqtt/message"
)

func formatTopicPath(path string) string {
	return "/" + strings.Trim(path, "/")
}

type Broker struct {
	port         int
	clients      map[string]chan []byte
	subscription *Subscription

	mu sync.Mutex
}

func NewBroker(port int) *Broker {
	return &Broker{
		port:         port,
		clients:      make(map[string]chan []byte),
		subscription: NewSubscription(),
	}
}

func (b *Broker) ListenAndServe(ctx context.Context) error {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", b.port))
	if err != nil {
		return err
	}

	defer listener.Close()
	log.Debugf("Broker server started at :%d", b.port)

	for {
		s, err := listener.Accept()
		if err != nil {
			log.Debug(err)
			continue
		}

		go b.handleConnection(s, ctx)
	}
	return nil
}

func (b *Broker) handleConnection(conn net.Conn, ctx context.Context) {
	client := NewClient(conn, ctx)
	defer func() {
		b.removeClient(client.Id())
		client.Close()
	}()

	if err := client.Handshake(10 * time.Second); err != nil {
		log.Debug("Failed to MQTT handshake: %s", err.Error())
		return
	}

	// Add start read/write thread loop
	go client.readLoop()

	// And message passing between server <-> client
	b.addClient(client.Id(), client.Publisher)

	for {
		select {
		case <-client.Closed():
			log.Debug("client context has been canceled")
			return
		case <-client.Timeout():
			log.Debug("client keep alive has been expired")
			return
		case packet := <-client.Packet:
			log.Debug("client packet received")
			if packet.Frame == nil {
				log.Debug("Received empty frame packet")
				continue
			}
			switch packet.Frame.Type {
			case message.PINGREQ:
				if _, err := message.ParsePingReq(packet.Frame, packet.Payload); err != nil {
					log.Debugf("failed to parse packet to PINGREQ: %s\n", err.Error())
					return
				}
				log.Debug("client PINGREQ received")
				client.Ping()
			case message.SUBSCRIBE:
				ss, err := message.ParseSubscribe(packet.Frame, packet.Payload)
				if err != nil {
					log.Debugf("failed to parse packet to SUBSCRIBE: %s\n", err.Error())
					return
				}
				log.Debug("client SUBSCRIBE received")
				rcs := []message.ReasonCode{}
				for _, t := range ss.Subscriptions {
					rcs = append(rcs, b.subscription.Add(client.Id(), t))
				}
				ack := message.NewSubAck(ss.PacketId, rcs...)

				if buf, err := ack.Encode(); err != nil {
					log.Debugf("failed to encode SUBACK packet: %s\n", err.Error())
					return
				} else {
					log.Debug("Sending SUBACK packet to the client")
					client.Send(buf)
				}
			case message.PUBLISH:
				log.Debug("receive PUBLISH packet: ", packet.Payload)
				pb, err := message.ParsePublish(packet.Frame, packet.Payload)
				if err != nil {
					log.Debugf("failed to parse packet to PUBLISH: %s\n", err.Error())
					return
				}
				ack, err := b.publishMessage(pb)
				if err != nil {
					log.Debugf("failed to publish message: %s\n", err.Error())
					return
				} else if ack == nil {
					log.Debug("ack is nil due to Qos0")
					return
				}
				if buf, err := ack.Encode(); err != nil {
					log.Debugf("failed to encode PUBACK/PUBREC packet: %s\n", err.Error())
					return
				} else {
					log.Debug("Sending PUBACK/PUBREC packet to the client")
					client.Send(buf)
				}
			default:
				log.Debugf("not implement packet type: %d\n", packet.Frame.Type)
			}
		}
	}
}

func (b *Broker) publishMessage(pb *message.Publish) (message.Encoder, error) {
	ids := b.subscription.FindAll(pb.TopicName)
	if len(ids) > 0 {
		buf, err := pb.Encode()
		if err != nil {
			return nil, err
		}
		b.mu.Lock()
		defer b.mu.Unlock()
		for _, id := range ids {
			if c, ok := b.clients[id]; ok {
				log.Debugf("send publish message to: %s\n", id)
				c <- buf
			}
		}
	}
	switch pb.QoS {
	case message.QoS0:
		return nil, nil
	case message.QoS1:
		return message.NewPubAck(pb.PacketId), nil
	case message.QoS2:
		return message.NewPubRel(pb.PacketId), nil
	default:
		// unreachable line
		return nil, nil
	}
}

func (b *Broker) addClient(clientId string, publisher chan []byte) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.clients[clientId] = publisher
}

func (b *Broker) removeClient(clientId string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if c, ok := b.clients[clientId]; ok {
		close(c)
		delete(b.clients, clientId)
	}
}
