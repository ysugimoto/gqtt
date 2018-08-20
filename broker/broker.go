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

var clients = make(map[string]chan *message.Publish)

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

		info, err := b.handshake(s, 10*time.Second)
		if err != nil {
			log.Debug("Failed to MQTT handshake: %s", err.Error())
			s.Close()
			continue
		}
		client := NewClient(s, *info, ctx, b)
		go b.handleConnection(client)
	}
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

func (b *Broker) publish(pb *message.Publish) {

	clientQoS := b.subscription.GetClientsByTopic(pb.TopicName)
	if len(clientQoS) == 0 {
		return
	}
	b.mu.Lock()
	defer b.mu.Unlock()

	for cid, qos := range clientQoS {
		c, ok := clients[cid]
		if !ok {
			continue
		}

		// Downgrade QoS if we need
		if pb.QoS > qos {
			log.Debugf("send publish message to: %s (downgraded %d -> %d)\n", cid, pb.QoS, qos)
			c <- pb.Downgrade(qos)
		} else {
			log.Debugf("send publish message to: %s with qos: %d", cid, pb.QoS)
			c <- pb
		}
	}
}

func (b *Broker) subscribe(client *Client, ss *message.Subscribe) (message.Encoder, error) {
	rcs := []message.ReasonCode{}
	// TODO: confirm subscription settings e.g. max QoS, ...
	for _, t := range ss.Subscriptions {
		rc, err := b.subscription.Subscribe(client.Id(), t)
		if err != nil {
			return nil, err
		}
		rcs = append(rcs, rc)
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
		b.subscription.UnsubscribeAll(clientId)
	}
}
