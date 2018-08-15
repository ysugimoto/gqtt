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

var clients = make(map[string]chan []byte)

type Broker struct {
	port         int
	subscription *Subscription

	mu sync.Mutex
}

func NewBroker(port int) *Broker {
	return &Broker{
		port:         port,
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

		if info, err := b.handshake(s, 10*time.Second); err != nil {
			log.Debug("Failed to MQTT handshake: %s", err.Error())
			s.Close()
			continue
		} else {
			client := NewClient(s, *info, ctx, b)
			go b.handleConnection(client)
		}
	}
	return nil
}

func (b *Broker) handshake(conn net.Conn, timeout time.Duration) (*message.Connect, error) {
	conn.SetDeadline(time.Now().Add(timeout))
	var (
		err     error
		reason  message.ReasonCode
		frame   *message.Frame
		payload []byte
		cn      *message.Connect
	)
	defer func() {
		log.Debug("defer: send CONNACK")
		ack := message.NewConnAck(reason)
		if err != nil {
			ack.Property = &message.ConnAckProperty{
				ReasonString: err.Error(),
			}
		}
		if buf, err := ack.Encode(); err != nil {
			log.Debug("CONNACK encode error: ", err)
		} else {
			conn.Write(buf)
		}
		conn.SetDeadline(time.Time{})
	}()

	frame, payload, err = message.ReceiveFrame(conn)
	if err != nil {
		log.Debug("receive frame error: ", err)
		reason = message.MalformedPacket
		return nil, err
	}
	cn, err = message.ParseConnect(frame, payload)
	if err != nil {
		reason = message.MalformedPacket
		log.Debug("frame expects connect package: ", err)
		return nil, err
	}
	reason = message.Success
	log.Debugf("CONNECT accepted: %+v\n", cn)
	return cn, nil
}

func (b *Broker) handleConnection(client *Client) {
	b.addClient(client)

	defer func() {
		log.Debug("============================ Client closing =======================")
		b.removeClient(client.Id())
		client.Close()
	}()

	for {
		select {
		case <-client.Closed():
			log.Debug("client context has been canceled")
			return
		}
	}
}

func (b *Broker) publish(pb *message.Publish) (message.Encoder, error) {
	ids := b.subscription.FindAll(pb.TopicName)
	log.Debugf("targets: %+v\n", ids)
	if len(ids) > 0 {
		buf, err := pb.Encode()
		if err != nil {
			log.Debug("failed to encode publish message: ", err)
			return nil, err
		}
		b.mu.Lock()
		defer b.mu.Unlock()
		log.Debugf("current clients: %+v\n", clients)
		for _, id := range ids {
			if c, ok := clients[id]; ok {
				log.Debugf("send publish message to: %s\n", id)
				c <- buf
			} else {
				log.Debugf("client %s not found\n", id)
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

func (b *Broker) subscribe(client *Client, ss *message.Subscribe) (message.Encoder, error) {
	rcs := []message.ReasonCode{}
	// TODO: confirm subscription settings e.g. max QoS, ...
	for _, t := range ss.Subscriptions {
		rcs = append(rcs, b.subscription.Add(client.Id(), t))
	}
	return message.NewSubAck(ss.PacketId, rcs...), nil
}

func (b *Broker) addClient(client *Client) {
	b.mu.Lock()
	defer b.mu.Unlock()
	clients[client.Id()] = client.Publisher
}

func (b *Broker) removeClient(clientId string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if c, ok := clients[clientId]; ok {
		close(c)
		delete(clients, clientId)
		b.subscription.RemoveAll(clientId)
	}
}
