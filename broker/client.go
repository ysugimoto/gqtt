package broker

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/satori/go.uuid"
	"github.com/ysugimoto/gqtt/internal/log"
	"github.com/ysugimoto/gqtt/message"
)

const defaultKeepAlive = 30

type Client struct {
	id        string
	ctx       context.Context
	conn      net.Conn
	timeout   *time.Timer
	Publisher chan message.Encoder
	terminate context.CancelFunc

	once         sync.Once
	info         message.Connect
	mu           sync.Mutex
	broker       *Broker
	pingInterval time.Duration
}

func NewClient(conn net.Conn, info message.Connect, ctx context.Context, b *Broker) *Client {
	client := &Client{
		id:        uuid.NewV4().String(),
		conn:      conn,
		Publisher: make(chan message.Encoder),
		info:      info,
		broker:    b,
	}
	client.ctx, client.terminate = context.WithCancel(ctx)
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
			case msg := <-client.Publisher:
				if msg == nil {
					return
				}
				if err := client.sendMessage(msg); err != nil {
					client.Close()
				}
			}
		}
	}()
	go client.loop()

	return client
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
			if err := c.sendMessage(ack); err != nil {
				log.Debug("failed to send message packet: ", err)
				return
			}
		}
	}
}

func (c *Client) sendMessage(m message.Encoder) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	buf, err := m.Encode()
	if err != nil {
		log.Debugf("failed to encode ack packet: %s\n", err.Error())
		return err
	}

	w := bufio.NewWriter(c.conn)
	if n, err := w.Write(buf); err != nil {
		log.Debug("failed to write packet: ", buf)
		return err
	} else if n != len(buf) {
		log.Debug("failed to write enough packet: ", buf)
		return fmt.Errorf("failed to write enough packet")
	} else if err := w.Flush(); err != nil {
		log.Debug("failed to flush packet: ", buf)
		return err
	}
	return nil
}
