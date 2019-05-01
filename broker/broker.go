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

const (
	capEventSize = 100
)

func formatTopicPath(path string) string {
	return "/" + strings.Trim(path, "/")
}

var clients = make(map[string]chan *message.Publish)

type Broker struct {
	port         int
	subscription *Subscription
	willPacketId uint16
	MessageEvent chan interface{}

	mu sync.Mutex
}

func NewBroker(port int) *Broker {
	return &Broker{
		port:         port,
		subscription: NewSubscription(),
		MessageEvent: make(chan interface{}, capEventSize),
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
			log.Debug("Failed to MQTT handshake: ", err.Error())
			s.Close()
			continue
		}
		client := NewClient(s, *info, ctx, b)
		go b.handleConnection(client)
	}
}

func (b *Broker) sendEvent(msg interface{}) {
	// Check overflow channel buffer
	if len(b.MessageEvent) >= capEventSize {
		log.Debug("Event channle overflow. You have to drain message")
	} else {
		b.MessageEvent <- msg
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
		if err := message.WriteFrame(conn, ack); err != nil {
			log.Debug("failed to send CONNACK: ", err)
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
	if err := b.authConnect(conn, cn.Property); err != nil {
		reason = message.NotAuthorized
		log.Debug("connection not authorized")
		return nil, err
	}
	reason = message.Success
	log.Debugf("CONNECT accepted: %+v\n", cn)
	b.sendEvent(cn)
	return cn, nil
}

func (b *Broker) authConnect(conn net.Conn, cp *message.ConnectProperty) error {
	// TODO: control to need to authneication on broker from setting or someway
	if cp == nil {
		return nil
	}
	switch cp.AuthenticationMethod {
	case basicAuthentication:
		return doBasicAuth(conn, cp)
	case loginAuthentication:
		return doLoginAuth(conn, cp)
	default:
		return fmt.Errorf("%s does not support or unrecognized", cp.AuthenticationMethod)
	}
}

func (b *Broker) handleConnection(client *Client) {
	b.addClient(client)

	defer func() {
		log.Debug("====== Client closing ======")
		b.removeClient(client.Id())
		client.Close(true)
	}()

	for {
		select {
		case <-client.Closed():
			log.Debug("client context has been canceled")
			return
		}
	}
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

func (b *Broker) Publish(pb *message.Publish) {
	si := b.subscription.GetClientsByTopic(pb.TopicName)
	if si == nil {
		return
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	log.Debug("start to send publish packet")
	b.sendEvent(pb)

	// Save as retain message is RETAIN bit is active
	if pb.RETAIN {
		// TODO: persistent save to DB, or some backend
		log.Debug("Save retian message to topic: ", pb.TopicName)
		si.RetainMessage = pb
	}
	// And we always set retain as set in order to distinguish retained message or not in client.
	pb.SetRetain(false)

	for cid, qos := range si.Clients {
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
	b.sendEvent(ss)
	return message.NewSubAck(ss.PacketId, rcs...), nil
}

func (b *Broker) will(c message.Connect) {
	if !c.FlagWill {
		log.Debug("client didn't want to use will. Skip")
		return
	}
	log.Debugf("client wants to send will message: qos: %d, topic: %s, body: %s", c.WillQoS, c.WillTopic, c.WillPayload)
	b.willPacketId++
	pb := message.NewPublish(b.willPacketId, message.WithQoS(c.WillQoS))
	pb.SetRetain(c.WillRetain)
	pb.TopicName = c.WillTopic
	pb.Body = []byte(c.WillPayload)
	if c.WillProperty != nil {
		pb.Property = c.WillProperty.ToPublish()
	}
	b.Publish(pb)
}

func (b *Broker) getRetainMessage(topicName string) *message.Publish {
	if si := b.subscription.GetClientsByTopic(topicName); si != nil {
		return si.RetainMessage
	}
	return nil
}

func (b *Broker) deleteRetainMessage(topicName string) {
	if si := b.subscription.GetClientsByTopic(topicName); si != nil {
		// TODO: delete from persistent storage
		si.RetainMessage = nil
	}
}
