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

type ClientOption = message.ConnectProperty
type ServerInfo = message.ConnAckProperty

func NewClientOption() *ClientOption {
	return &ClientOption{}
}

type Client struct {
	packetId     uint16
	url          string
	conn         net.Conn
	ctx          context.Context
	terminate    context.CancelFunc
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

func (c *Client) Connect(ctx context.Context) error {
	return c.ConnectWithOption(ctx, nil)
}

func (c *Client) ConnectWithOption(ctx context.Context, opt *ClientOption) error {
	var err error

	if c.conn, c.ServerInfo, err = connect(c.url, opt); err != nil {
		return err
	}

	log.Debug("connection established!")

	c.ctx, c.terminate = context.WithCancel(ctx)
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
				log.Debug("ping interval is coming. send ping to server")
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

// func (c *Client) sendMessage(m message.Encoder) error {
// 	c.mu.Lock()
// 	defer func() {
// 		// This is a trick for sending mutex blocked message queue properly.
// 		// Wait a tiny microseconds before unlock mutex, it makes socket accepts to write next packet again.
// 		time.Sleep(10 * time.Microsecond)
// 		c.mu.Unlock()
// 	}()
//
// 	buf, err := m.Encode()
// 	if err != nil {
// 		log.Debug("failed to encode message ", err)
// 		return err
// 	}
//
// 	if n, err := c.conn.Write(buf); err != nil {
// 		log.Debug("failed to write packet: ", buf)
// 		return err
// 	} else if n != len(buf) {
// 		log.Debug("could not enough patck")
// 		return errors.New("could not write enough packet")
// 	}
// 	return nil
// }

func (c *Client) mainLoop() {
	defer c.terminate()
	for {
		select {
		case <-c.ctx.Done():
			log.Debugf("terminated")
			c.Disconnect()
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
				pr, err := message.ParsePingResp(frame, payload)
				if err != nil {
					log.Debug("malformed packet: failed to decode to PINGREQ packet: ", err)
					continue
				}
				log.Debugf("PINGRESP received: %+v\n", pr)
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
				log.Debug("PUBLISH message received")
				pb, err := message.ParsePublish(frame, payload)
				if err != nil {
					log.Debug("malformed packet: failed to decode to PUBLISH packet: ", err)
					continue
				}
				c.Message <- pb
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

// func (c *Client) runSession(packetId uint16, mt message.MessageType, msg message.Encoder) (interface{}, error) {
// 	recv := make(chan interface{})
// 	c.sessions[packetId] = session{
// 		Type:    mt,
// 		Channel: recv,
// 	}
// 	ctx, timeout := context.WithTimeout(c.ctx, 10*time.Second)
// 	defer timeout()
//
// 	// Send message
// 	if err := c.sendMessage(msg); err != nil {
// 		return nil, err
// 	}
// 	// wait or timeout
// 	select {
// 	case <-ctx.Done():
// 		return nil, ctx.Err()
// 	case ack := <-recv:
// 		return ack, nil
// 	}
// }

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
		// TODO: Need to save and delete message for QoS1
	case message.QoS2:
		pb.PacketId = c.makePacketId()
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
