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
	Packet    chan *message.Packet
	Publisher chan []byte
	send      chan []byte
	terminate context.CancelFunc

	once         sync.Once
	writer       *bufio.Writer
	info         message.Connect
	mu           sync.Mutex
	broker       *Broker
	pingInterval time.Duration
}

func NewClient(conn net.Conn, info message.Connect, ctx context.Context, b *Broker) *Client {
	c, cancel := context.WithCancel(ctx)
	var pingInterval time.Duration
	if info.KeepAlive > 0 {
		pingInterval = time.Duration(info.KeepAlive)
	} else {
		pingInterval = time.Duration(defaultKeepAlive) * time.Second
	}

	client := &Client{
		id:           uuid.NewV4().String(),
		conn:         conn,
		ctx:          c,
		terminate:    cancel,
		Publisher:    make(chan []byte),
		send:         make(chan []byte),
		info:         info,
		broker:       b,
		pingInterval: pingInterval,
		timeout:      time.AfterFunc(pingInterval, cancel),
	}

	go client.loop()
	go func() {
		for {
			select {
			case msg := <-client.Publisher:
				client.sendPacket(msg)
			}
		}
	}()
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
		if c.timeout != nil {
			c.timeout.Stop()
		}
	})
}

func (c *Client) loop() {
	for {
		frame, payload, err := message.ReceiveFrame(c.conn)
		if err != nil {
			c.terminate()
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

			// Enhanced expiration
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
		if ack == nil {
			continue
		}
		if buf, err := ack.Encode(); err != nil {
			log.Debugf("failed to encode ack packet: %s\n", err.Error())
			return
		} else {
			log.Debug("Sending ack packet to the client")
			c.sendPacket(buf)
		}
	}
}

func (c *Client) sendPacket(m []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	w := bufio.NewWriter(c.conn)
	if n, err := w.Write(m); err != nil {
		log.Debug("failed to write packet: ", m)
		return err
	} else if n != len(m) {
		log.Debug("failed to write enough packet: ", m)
		return fmt.Errorf("failed to write enough packet")
	} else if err := w.Flush(); err != nil {
		log.Debug("failed to flush packet: ", m)
		return err
	}
	return nil
}
