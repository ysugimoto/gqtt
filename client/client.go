package client

import (
	"context"
	"errors"
	"github.com/ysugimoto/gqtt/internal/log"
	"github.com/ysugimoto/gqtt/message"
	"github.com/ysugimoto/gqtt/session"
	"net"
	"sync"
	"time"
)

type ConnectionOption = message.ConnectProperty
type ServerInfo = message.ConnAckProperty

type Client struct {
	packetId     uint16
	url          string
	conn         net.Conn
	ctx          context.Context
	session      *session.Session
	pingInterval *time.Ticker

	Closed  chan struct{}
	Message chan *message.Publish

	once       sync.Once
	ServerInfo *ServerInfo
	mu         sync.Mutex
}

func NewClient(u string) *Client {
	return &Client{
		url: u,
	}
}

func (c *Client) Connect(ctx context.Context, options ...ClientOption) error {
	var err error

	if c.conn, c.ServerInfo, err = connect(c.url, options); err != nil {
		return err
	}

	log.Debug("connection established!")

	c.ctx = ctx
	c.Closed = make(chan struct{})
	c.Message = make(chan *message.Publish)
	c.session = session.New(c.conn, c.ctx)

	// TODO: set duration from option
	go func() {
		c.pingInterval = time.NewTicker(5 * time.Second)
		for {
			select {
			case <-c.ctx.Done():
				c.Disconnect()
				return
			case <-c.pingInterval.C:
				c.session.Write(message.NewPingReq())
			}
		}
	}()
	go c.mainLoop()

	return nil
}

func (c *Client) Disconnect() {
	c.once.Do(func() {
		log.Debug("============================ Client closing =======================")
		c.pingInterval.Stop()

		dc := message.NewDisconnect(message.NormalDisconnection)
		if err := c.session.Write(dc); err != nil {
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
	defer c.Disconnect()
	for {
		select {
		case <-c.ctx.Done():
			log.Debugf("terminated")
			return
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
				log.Debug("PINGRESP received from broker")
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
				pb, ok := c.session.LoadMessage(pl.PacketId)
				if !ok {
					log.Debug("Client received PUBREL packet, but message wan't saved")
					continue
				}
				pc := message.NewPubComp(pl.PacketId)
				if err := c.session.Write(pc); err != nil {
					log.Debug("failed to send PUBCOMP packet to publisher")
					continue
				}
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
	if c.packetId == 0xFFFF {
		c.packetId = 0
	}
	c.packetId++
	return c.packetId
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

func (c *Client) Publish(topic string, qos message.QoSLevel, body []byte) error {
	pb := message.NewPublish(0, message.WithQoS(qos))
	pb.TopicName = topic
	pb.Body = body
	switch qos {
	case message.QoS0:
		// If OoS is zero, we don't need packet identifier and any acknowledgment
		if err := c.session.Write(pb); err != nil {
			log.Debug("failed to send publish with QoS0 ", err)
			return err
		}
	case message.QoS1:
		pb.PacketId = c.makePacketId()
		if ack, err := c.session.Start(pb.PacketId, message.PUBACK, pb, session.MaxRetries); err != nil {
			log.Debug("failed to publish session for OoS1: ", err)
			return err
		} else if _, ok := ack.(*message.PubAck); !ok {
			log.Debug("failed to type conversion for OoS1")
			return errors.New("failed to type conversion for OoS1")
		}
	case message.QoS2:
		pb.PacketId = c.makePacketId()
		if ack, err := c.session.Start(pb.PacketId, message.PUBREC, pb, session.MaxRetries); err != nil {
			log.Debug("failed to publish session for OoS2: ", err)
			return err
		} else if _, ok := ack.(*message.PubRec); !ok {
			log.Debug("failed to type conversion fto PUBREC or OoS2")
			return errors.New("failed to type conversion fto PUBREC or OoS2")
		}
		log.Debug("PUBREC received. Send PUBREL")
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

func (c *Client) receivePublish(pb *message.Publish) error {
	log.Debugf("PUBLISH message received with QoS: %d\n", pb.QoS)
	switch pb.QoS {
	case message.QoS0:
		c.Message <- pb
	case message.QoS1:
		log.Debug("Send PUBACK to the publisher")
		if err := c.session.Write(message.NewPubAck(pb.PacketId)); err != nil {
			log.Debug("failed to send PUBACK packet")
			return err
		}
		c.Message <- pb
	case message.QoS2:
		c.session.StoreMessage(pb)
		if err := c.session.Write(message.NewPubRec(pb.PacketId)); err != nil {
			log.Debug("failed to send PUBREC packet")
			return err
		}
	default:
		return errors.New("unexpected QoS")
	}
	return nil
}
