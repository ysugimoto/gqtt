package broker

import (
	"context"
	"net"
	"sync"
	"time"

	"github.com/pkg/errors"
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
	client.timeout = time.AfterFunc(client.pingInterval, func() {
		log.Debug("keepalive timeout")
		client.terminate()
	})

	go func() {
		for {
			select {
			case pb := <-client.Publisher:
				if pb == nil {
					return
				}
				if err := client.publish(pb); err != nil {
					client.Close(true)
				}
			}
		}
	}()
	go client.loop()

	return client
}

func (c *Client) publish(pb *message.Publish) error {
	log.Debugf("broker publish to client: qos: %d, message: %s\n", pb.QoS, string(pb.Body))
	switch pb.QoS {
	case message.QoS0:
		if err := message.WriteFrame(c.conn, pb); err != nil {
			return errors.Wrap(err, "failed to write publish packet")
		}
	case message.QoS1:
		if ack, err := c.session.Start(pb.PacketId, message.PUBACK, pb, session.MaxRetries); err != nil {
			log.Debug("failed to publish session for OoS1: ", err)
			return errors.Wrap(err, "failed to publish session for QoS1")
		} else if _, ok := ack.(*message.PubAck); !ok {
			log.Debug("failed to type conversion for OoS1")
			return errors.New("failed to type conversion for OoS1")
		}
		// TODO: Need to save and delete message for QoS1
	case message.QoS2:
		if ack, err := c.session.Start(pb.PacketId, message.PUBREC, pb, session.MaxRetries); err != nil {
			log.Debug("failed to publish session for OoS2: ", err)
			return errors.Wrap(err, "failed to publish session for QoS2")
		} else if _, ok := ack.(*message.PubRec); !ok {
			log.Debug("failed to type conversion fto PUBREC or OoS2")
			return errors.New("failed to type conversion fto PUBREC or OoS2")
		}
		time.Sleep(10 * time.Millisecond)
		// On QoS2, need to send more packet for PUBREL
		pl := message.NewPubRel(pb.PacketId)
		log.Debug("success to receive PUBREC: ", pl.PacketId)
		if ack, err := c.session.Start(pb.PacketId, message.PUBCOMP, pl, session.MaxRetries); err != nil {
			log.Debug("failed to pubrel session for OoS2: ", err)
			return errors.Wrap(err, "failed to pubrel session for QoS2")
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

func (c *Client) Close(isWill bool) {
	c.once.Do(func() {
		c.terminate()
		c.timeout.Stop()
		c.conn.Close()
		if isWill {
			c.broker.will(c.info)
		}
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
		if frame == nil {
			log.Debug("Received empty frame packet")
			continue
		}
		var ack message.Encoder
		switch frame.Type {
		case message.DISCONNECT:
			c.Close(false)
			return
		case message.PINGREQ:
			if _, err := message.ParsePingReq(frame, payload); err != nil {
				log.Debugf("failed to parse packet to PINGREQ: %s\n", err.Error())
				return
			}
			// Extend expiration and respond PINGRESP
			c.timeout.Reset(c.pingInterval)
			if err := message.WriteFrame(c.conn, message.NewPingResp()); err != nil {
				log.Debug("failed to send PINGRESP: ", err)
				return
			}
		case message.SUBSCRIBE:
			ss, err := message.ParseSubscribe(frame, payload)
			if err != nil {
				log.Debugf("failed to parse packet to SUBSCRIBE: %s\n", err.Error())
				return
			}
			log.Debug("client SUBSCRIBE received")
			if ack, err = c.broker.subscribe(c, ss); err != nil {
				log.Debugf("failed to add subscribe: %s\n", err.Error())
				return
			} else if err := message.WriteFrame(c.conn, ack); err != nil {
				log.Debug("failed to send SUBACK: ", err)
				return
			}
			// Send retain message if exists
			for _, s := range ss.Subscriptions {
				retain := c.broker.getRetainMessage(s.TopicName)
				if retain != nil {
					log.Debug("Send retain message for topic: ", s.TopicName)
					retain.SetRetain(true)
					if err := message.WriteFrame(c.conn, retain); err != nil {
						log.Debug("failed to send retain message: ", err)
					}
				}
			}
		case message.PUBLISH:
			pb, err := message.ParsePublish(frame, payload)
			if err != nil {
				log.Debugf("failed to parse packet to PUBLISH: %s\n", err.Error())
				return
			}
			log.Debugf("Publish message received with QoS: %d from: %s, body: %s\n", pb.QoS, c.Id(), string(pb.Body))

			// Check retain message deletion
			// If RETAIN flag is on and message size is zero, then we delete retain message
			if pb.RETAIN && len(pb.Body) == 0 {
				log.Debugf("Retain message deletion for topic: %s\n", pb.TopicName)
				c.broker.deleteRetainMessage(pb.TopicName)
				continue
			}
			switch pb.QoS {
			case message.QoS0:
				// QoS0 publishes message immediately
				c.broker.Publish(pb)
			case message.QoS1:
				// QoS1 publishes message and respond PUBACK
				if err := message.WriteFrame(c.conn, message.NewPubAck(pb.PacketId)); err != nil {
					log.Debug("failed to send PUBACK: ", err)
					return
				}
				c.broker.Publish(pb)
			case message.QoS2:
				// QoS2 stores message and publish after PUBREL packet received
				c.session.StoreMessage(pb)
				if err := message.WriteFrame(c.conn, message.NewPubRec(pb.PacketId)); err != nil {
					log.Debug("failed to send PUBREC: ", err)
				}
			}
		case message.PUBACK:
			pa, err := message.ParsePubAck(frame, payload)
			if err != nil {
				log.Debug("malformed packet: failed to decode to PUBACK packet: ", err)
				continue
			}
			if err := c.session.Meet(pa.PacketId, message.PUBACK, pa); err != nil {
				log.Debug("malformed packet: unexpected packet identifier received: ", err)
				continue
			}
		case message.PUBREC:
			pr, err := message.ParsePubRec(frame, payload)
			if err != nil {
				log.Debug("malformed packet: failed to decode to PUBREC packet: ", err)
				continue
			}
			if err := c.session.Meet(pr.PacketId, message.PUBREC, pr); err != nil {
				log.Debug("malformed packet: unexpected packet identifier received: ", err)
				continue
			}
		case message.PUBREL:
			pl, err := message.ParsePubRel(frame, payload)
			if err != nil {
				log.Debug("malformed packet: failed to decode to PUBREL packet: ", err)
				continue
			}
			pb, ok := c.session.LoadMessage(pl.PacketId)
			if !ok {
				log.Debug("Broker recevied PUBREL packet, but message didn't exist")
				continue
			}
			if err := message.WriteFrame(c.conn, message.NewPubComp(pl.PacketId)); err != nil {
				log.Debug("failed to send PUBCOMP pakcet: ", err)
				continue
			}
			c.session.DeleteMessage(pl.PacketId)
			c.broker.Publish(pb)
		case message.PUBCOMP:
			pc, err := message.ParsePubComp(frame, payload)
			if err != nil {
				log.Debug("malformed packet: failed to decode to PUBCOMP packet: ", err)
				continue
			}
			if err := c.session.Meet(pc.PacketId, message.PUBCOMP, pc); err != nil {
				log.Debug("malformed packet: unexpected packet identifier received: ", err)
				continue
			}
		default:
			log.Debugf("not implement packet type: %d\n", frame.Type)
			continue
		}
	}
}
