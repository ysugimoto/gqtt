package client

import (
	"context"
	"net"
	"sync"
	"time"

	"sync/atomic"

	"github.com/pkg/errors"
	"github.com/ysugimoto/gqtt/internal/log"
	"github.com/ysugimoto/gqtt/message"
	"github.com/ysugimoto/gqtt/session"
)

type ConnectionOption = message.ConnectProperty
type ServerInfo = message.ConnAckProperty

type Client struct {
	packetId *uint32
	url      string
	conn     net.Conn
	ctx      context.Context
	session  *session.Session

	Closed  chan struct{}
	Message chan *message.Publish

	once       sync.Once
	ServerInfo *ServerInfo
	mu         sync.Mutex
}

func NewClient(u string) *Client {
	var packetId uint32 = 0
	return &Client{
		packetId: &packetId,
		url:      u,
	}
}

func (c *Client) Connect(ctx context.Context, options ...ClientOption) error {
	var err error

	if c.conn, c.ServerInfo, err = connect(c.url, options); err != nil {
		return errors.Wrap(err, "failed to connect to "+c.url)
	}

	log.Debug("connection established!")

	c.ctx = ctx
	c.Closed = make(chan struct{})
	c.Message = make(chan *message.Publish)
	c.session = session.New(c.conn, c.ctx)

	go c.mainLoop()

	return nil
}

func (c *Client) Disconnect() {
	c.once.Do(func() {
		log.Debug("============================ Client closing =======================")

		dc := message.NewDisconnect(message.NormalDisconnection)
		if err := message.WriteFrame(c.conn, dc); err != nil {
			log.Debug("failed to send DISCONNECT message: ", err)
		}
		log.Debug("Closing connection")
		c.conn.Close()
		log.Debug("connection closed, send channel")
		c.Closed <- struct{}{}
		log.Debug("channel sent")
	})
}

func (c *Client) mainLoop() {
	pingInterval := time.NewTicker(5 * time.Second)
	defer pingInterval.Stop()
	defer c.Disconnect()

	for {
		select {
		case <-c.ctx.Done():
			log.Debugf("terminated")
			return
		case <-pingInterval.C:
			message.WriteFrame(c.conn, message.NewPingReq())
		default:
			frame, payload, err := message.ReceiveFrame(c.conn)
			if err != nil {
				log.Debug("failed to receive message: ", err)
				if nerr, ok := err.(net.Error); ok {
					if nerr.Temporary() {
						log.Debug("buffer is temporary, backoff")
						continue
					}
				}
				return
			}
			switch frame.Type {
			case message.PINGRESP:
				if _, err := message.ParsePingResp(frame, payload); err != nil {
					log.Debug("malformed packet: failed to decode to PINGREQ packet: ", err)
					continue
				}
			case message.SUBACK:
				ack, err := message.ParseSubAck(frame, payload)
				if err != nil {
					log.Debug("malformed packet: failed to decode to SUBACK packet: ", err)
					continue
				}
				if err := c.session.Meet(ack.PacketId, message.SUBACK, ack); err != nil {
					log.Debug("malformed packet: unexpected packet identifier received: ", err)
					continue
				}
			case message.PUBACK:
				ack, err := message.ParsePubAck(frame, payload)
				if err != nil {
					log.Debug("malformed packet: failed to decode to PUBACK packet: ", err)
					continue
				}
				if err := c.session.Meet(ack.PacketId, message.PUBACK, ack); err != nil {
					log.Debug("malformed packet: unexpected packet identifier received: ", err)
					continue
				}
			case message.PUBLISH:
				pb, err := message.ParsePublish(frame, payload)
				if err != nil {
					log.Debug("malformed packet: failed to decode to PUBLISH packet: ", err)
					continue
				} else if err := c.receivePublish(pb); err != nil {
					log.Debug("client failed to process publish pakcet: ", err)
					continue
				}
			case message.PUBREC:
				ack, err := message.ParsePubRec(frame, payload)
				if err != nil {
					log.Debug("malformed packet: failed to decode to PUBREC packet: ", err)
					continue
				}
				if err := c.session.Meet(ack.PacketId, message.PUBREC, ack); err != nil {
					log.Debug("malformed packet: unexpected packet identifier received: ", err)
					continue
				}
			case message.PUBREL:
				pl, err := message.ParsePubRel(frame, payload)
				if err != nil {
					log.Debug("malformed packet: failed to decode to PUBREL packet: ", err)
					continue
				}
				log.Debug("PUBREL package received")
				pb, ok := c.session.LoadMessage(pl.PacketId)
				if !ok {
					log.Debug("Client received PUBREL packet, but message wan't saved")
					continue
				}
				log.Debug("Message found")
				pc := message.NewPubComp(pl.PacketId)
				if err := message.WriteFrame(c.conn, pc); err != nil {
					log.Debug("failed to send PUBCOMP packet to publisher")
					continue
				}
				log.Debug("Send PUBCOMP")
				c.Message <- pb
				c.session.DeleteMessage(pl.PacketId)
			case message.PUBCOMP:
				ack, err := message.ParsePubComp(frame, payload)
				if err != nil {
					log.Debug("malformed packet: failed to decode to PUBCOMP packet: ", err)
					continue
				}
				if err := c.session.Meet(ack.PacketId, message.PUBCOMP, ack); err != nil {
					log.Debug("malformed packet: unexpected packet identifier received: ", err)
					continue
				}
			}
		}
	}
}

func (c *Client) makePacketId() uint16 {
	if *c.packetId == 0xFFFF {
		atomic.StoreUint32(c.packetId, 1)
	} else {
		atomic.AddUint32(c.packetId, 1)
	}
	return uint16(*c.packetId)
}

func (c *Client) Subscribe(topic string, qos message.QoSLevel) error {
	packetId := c.makePacketId()

	ss := message.NewSubscribe()
	ss.PacketId = packetId
	// TODO: enable to set Retain handling, NoLocal flag
	ss.AddTopic(message.SubscribeTopic{
		TopicName: topic,
		QoS:       qos,
	})

	log.Debug("send subscribe")
	if ack, err := c.session.Start(packetId, message.SUBACK, ss, 0); err != nil {
		log.Debug("failed to finish session: ", err)
		return err
	} else if _, ok := ack.(*message.SubAck); !ok {
		log.Debug("unexpected ack received")
		return errors.New("unexpected ack received")
	}
	log.Debug("sent subscribe")
	return nil
}

func (c *Client) Publish(topic string, body []byte, opts ...ClientOption) error {
	pb := message.NewPublish(0, message.WithQoS(message.QoS0))
	for _, o := range opts {
		switch o.name {
		case nameRetain:
			pb.SetRetain(true)
		case nameQoS:
			pb.SetQoS(o.value.(message.QoSLevel))
		}
	}
	pb.TopicName = topic
	pb.Body = body
	switch pb.QoS {
	case message.QoS0:
		// If OoS is zero, we don't need packet identifier and any acknowledgment
		if err := message.WriteFrame(c.conn, pb); err != nil {
			log.Debug("failed to send publish with QoS0 ", err)
			return errors.Wrap(err, "failed to send publish with QoS0")
		}
	case message.QoS1:
		pb.PacketId = c.makePacketId()
		if ack, err := c.session.Start(pb.PacketId, message.PUBACK, pb, session.MaxRetries); err != nil {
			log.Debug("failed to publish session for QoS1: ", err)
			return errors.Wrap(err, "failed to publish session for QoS1")
		} else if _, ok := ack.(*message.PubAck); !ok {
			log.Debug("failed to type conversion for OoS1")
			return errors.New("failed to type conversion for QoS1")
		}
	case message.QoS2:
		pb.PacketId = c.makePacketId()
		if ack, err := c.session.Start(pb.PacketId, message.PUBREC, pb, session.MaxRetries); err != nil {
			log.Debug("failed to publish session for QoS2: ", err)
			return errors.Wrap(err, "failed to publish session for QoS2")
		} else if _, ok := ack.(*message.PubRec); !ok {
			log.Debug("failed to type conversion fto PUBREC or QoS2")
			return errors.New("failed to type conversion fto PUBREC or QoS2")
		}
		log.Debug("PUBREC received. Send PUBREL")
		time.Sleep(10 * time.Millisecond)
		// On QoS2, need to send more packet for PUBREL
		pl := message.NewPubRel(pb.PacketId)
		if ack, err := c.session.Start(pb.PacketId, message.PUBCOMP, pl, session.MaxRetries); err != nil {
			log.Debug("failed to pubrel session for QoS2: ", err)
			return errors.Wrap(err, "failed to pubrel session for QoS2")
		} else if _, ok := ack.(*message.PubComp); !ok {
			log.Debug("failed to type conversion to PUBCOMP for QoS2")
			return errors.New("failed to type conversion to PUBCOMP for QoS2")
		}
	}
	return nil
}

func (c *Client) receivePublish(pb *message.Publish) error {
	log.Debugf("PUBLISH message received with QoS: %d\n", pb.QoS)
	switch pb.QoS {
	case message.QoS0:
		c.Message <- pb
	case message.QoS1:
		log.Debug("Send PUBACK to the publisher")
		if err := message.WriteFrame(c.conn, message.NewPubAck(pb.PacketId)); err != nil {
			log.Debug("failed to send PUBACK packet")
			return errors.Wrap(err, "failed to send PUBACK packet")
		}
		c.Message <- pb
	case message.QoS2:
		c.session.StoreMessage(pb)
		if err := message.WriteFrame(c.conn, message.NewPubRec(pb.PacketId)); err != nil {
			log.Debug("failed to send PUBREC packet")
			return errors.Wrap(err, "failed to send PUBREC packet")
		}
		log.Debugf("Message saved for packetID: %d\n", pb.PacketId)
	default:
		return errors.New("unexpected QoS")
	}
	return nil
}
