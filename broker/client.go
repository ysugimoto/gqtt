package broker

import (
	"context"
	"errors"
	"net"
	"sync"
	"time"

	"github.com/satori/go.uuid"
	"github.com/ysugimoto/gqtt/internal/log"
	"github.com/ysugimoto/gqtt/message"
	"github.com/ysugimoto/gqtt/session"
)

const defaultKeepAlive = 30

type Client struct {
	id        string
	ctx       context.Context
	conn      net.Conn
	timeout   *time.Timer
	Publisher chan *message.Publish
	terminate context.CancelFunc
	session   *session.Session

	once         sync.Once
	info         message.Connect
	mu           sync.Mutex
	broker       *Broker
	pingInterval time.Duration
}

func NewClient(conn net.Conn, info message.Connect, ctx context.Context, b *Broker) *Client {
	cctx, terminate := context.WithCancel(ctx)
	client := &Client{
		id:        uuid.NewV4().String(),
		conn:      conn,
		Publisher: make(chan *message.Publish),
		info:      info,
		broker:    b,
		ctx:       cctx,
		terminate: terminate,
		session:   session.New(conn, cctx),
	}
	if info.KeepAlive > 0 {
		client.pingInterval = time.Duration(info.KeepAlive) * time.Second
	} else {
		client.pingInterval = time.Duration(defaultKeepAlive) * time.Second
	}
	log.Debug(client.pingInterval)
	client.timeout = time.AfterFunc(client.pingInterval, client.terminate)

	go func() {
		for {
			select {
			case pb := <-client.Publisher:
				if pb == nil {
					return
				}
				if err := client.publish(pb); err != nil {
					client.Close()
				}
			}
		}
	}()
	go client.loop()

	return client
}

func (c *Client) publish(pb *message.Publish) error {
	switch pb.QoS {
	case message.QoS0:
		if err := c.session.Write(pb); err != nil {
			return err
		}
	case message.QoS1:
		if ack, err := c.session.Start(pb.PacketId, message.PUBACK, pb, session.MaxRetries); err != nil {
			log.Debug("failed to publish session for OoS1: ", err)
			return err
		} else if _, ok := ack.(*message.PubAck); !ok {
			log.Debug("failed to type conversion for OoS1")
			return errors.New("failed to type conversion for OoS1")
		}
		// TODO: Need to save and delete message for QoS1
	case message.QoS2:
		if ack, err := c.session.Start(pb.PacketId, message.PUBREC, pb, session.MaxRetries); err != nil {
			log.Debug("failed to publish session for OoS2: ", err)
			return err
		} else if _, ok := ack.(*message.PubRec); !ok {
			log.Debug("failed to type conversion fto PUBREC or OoS2")
			return errors.New("failed to type conversion fto PUBREC or OoS2")
		}
		// On QoS2, need to send more packet for PUBREL
		pl := message.NewPubRel(pb.PacketId)
		if ack, err := c.session.Start(pb.PacketId, message.PUBCOMP, pl, session.MaxRetries); err != nil {
			log.Debug("failed to pubrel session for OoS2: ", err)
			return err
		} else if _, ok := ack.(*message.PubComp); !ok {
			log.Debug("failed to type conversion to PUBCOMP for OoS2")
			return errors.New("failed to type conversion to PUBCOMP for OoS2")
		}
	}
	return nil
}

func (c *Client) Closed() <-chan struct{} {
	return c.ctx.Done()
}

func (c *Client) Id() string {
	return c.id
}

func (c *Client) Close() {
	c.once.Do(func() {
		c.terminate()
		c.conn.Close()
		c.timeout.Stop()
	})
}

func (c *Client) loop() {
	defer c.terminate()

	for {
		frame, payload, err := message.ReceiveFrame(c.conn)
		if err != nil {
			log.Debug("client packet receive failed")
			return
		}
		log.Debug("client packet received")
		if frame == nil {
			log.Debug("Received empty frame packet")
			continue
		}
		var ack message.Encoder
		switch frame.Type {
		case message.PINGREQ:
			if _, err := message.ParsePingReq(frame, payload); err != nil {
				log.Debugf("failed to parse packet to PINGREQ: %s\n", err.Error())
				return
			}
			log.Debug("client PINGREQ received")

			// Extend expiration
			c.timeout.Reset(c.pingInterval)
			ack = message.NewPingResp()
		case message.SUBSCRIBE:
			ss, err := message.ParseSubscribe(frame, payload)
			if err != nil {
				log.Debugf("failed to parse packet to SUBSCRIBE: %s\n", err.Error())
				return
			}
			log.Debug("client SUBSCRIBE received")
			ack, err = c.broker.subscribe(c, ss)
			if err != nil {
				log.Debugf("failed to add subscribe: %s\n", err.Error())
				return
			}
		case message.PUBLISH:
			pb, err := message.ParsePublish(frame, payload)
			if err != nil {
				log.Debugf("failed to parse packet to PUBLISH: %s\n", err.Error())
				return
			}
			log.Debugf("Publish message received from: %s, body: %s\n", c.Id(), string(pb.Body))
			ack, err = c.broker.publish(pb)
			if err != nil {
				log.Debugf("failed to publish message: %s\n", err.Error())
				return
			} else if ack == nil {
				log.Debug("ack is nil due to Qos0")
				continue
			}
		default:
			log.Debugf("not implement packet type: %d\n", frame.Type)
			continue
		}

		// If client need to send ack message, send it
		if ack != nil {
			if err := c.session.Write(ack); err != nil {
				log.Debug("failed to send message packet: ", err)
				return
			}
		}
	}
}

// func (c *Client) sendMessage(m message.Encoder) error {
// 	c.mu.Lock()
// 	defer c.mu.Unlock()
//
// 	buf, err := m.Encode()
// 	if err != nil {
// 		log.Debugf("failed to encode ack packet: %s\n", err.Error())
// 		return err
// 	}
//
// 	w := bufio.NewWriter(c.conn)
// 	if n, err := w.Write(buf); err != nil {
// 		log.Debug("failed to write packet: ", buf)
// 		return err
// 	} else if n != len(buf) {
// 		log.Debug("failed to write enough packet: ", buf)
// 		return fmt.Errorf("failed to write enough packet")
// 	} else if err := w.Flush(); err != nil {
// 		log.Debug("failed to flush packet: ", buf)
// 		return err
// 	}
// 	return nil
// }
