package broker

import (
	"bufio"
	"context"
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
	timeout   *time.Ticker
	Packet    chan *message.Packet
	Publisher chan []byte
	send      chan []byte
	terminate context.CancelFunc

	once   sync.Once
	writer *bufio.Writer
	info   message.Connect
}

func NewClient(conn net.Conn, ctx context.Context) *Client {
	c, cancel := context.WithCancel(ctx)
	return &Client{
		id:        uuid.NewV4().String(),
		conn:      conn,
		ctx:       c,
		terminate: cancel,
		Packet:    make(chan *message.Packet),
		Publisher: make(chan []byte),
		send:      make(chan []byte),
	}
}

func (c *Client) Closed() <-chan struct{} {
	return c.ctx.Done()
}

func (c *Client) Timeout() <-chan time.Time {
	if c.timeout == nil {
		return nil
	}
	return c.timeout.C
}

func (c *Client) Id() string {
	return c.id
}

func (c *Client) Read(b []byte) (int, error) {
	return c.conn.Read(b)
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

func (c *Client) Handshake(timeout time.Duration) (err error) {
	c.conn.SetDeadline(time.Now().Add(timeout))
	var (
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
			c.Send(buf)
		}
		c.conn.SetDeadline(time.Time{})
	}()

	frame, payload, err = message.ReceiveFrame(c)
	if err != nil {
		log.Debug("receive frame error: ", err)
		reason = message.MalformedPacket
		return err
	}
	cn, err = message.ParseConnect(frame, payload)
	if err != nil {
		reason = message.MalformedPacket
		log.Debug("frame expects connect package: ", err)
		return err
	}
	c.info = *cn
	log.Debugf("CONNECT accepted: %+v\n", cn)
	reason = message.Success
	return nil
}

func (c *Client) readLoop() {
	var to time.Duration
	if c.info.KeepAlive > 0 {
		to = time.Duration(c.info.KeepAlive)
	} else {
		to = time.Duration(defaultKeepAlive)
	}

	c.timeout = time.NewTicker(to * time.Second)

	for {
		frame, payload, err := message.ReceiveFrame(c)
		if err != nil {
			c.Close()
		}
		log.Debug("readLoop(); packet received")
		c.Packet <- message.NewPacket(frame, payload)
	}
}

func (c *Client) Send(m []byte) {
	if c.writer == nil {
		go c.writeLoop()
	}
	c.send <- m
}

func (c *Client) writeLoop() {
	log.Debug("Start write loop")
	c.writer = bufio.NewWriter(c.conn)
	for {
		select {
		case <-c.Closed():
			c.Close()
			return
		case buf := <-c.send:
			log.Debug("accept send buffer")
			if _, err := c.writer.Write(buf); err != nil {
				log.Debugf("socket write error: %s", err.Error())
				// Backoff when error is temporary net error
				if ne, ok := err.(net.Error); ok {
					if ne.Temporary() {
						log.Debugf("socket error is temporary, backoff")
						time.Sleep(10 * time.Millisecond)
						c.send <- buf
						break
					}
				}
				return
			}
			if err := c.writer.Flush(); err != nil {
				log.Debugf("writer flush error: %s", err.Error())
				return
			}
			log.Debug("buffer sent successfuuly")
		case buf := <-c.Publisher:
			log.Debug("Received buffer from publisher")
			if _, err := c.writer.Write(buf); err != nil {
				log.Debugf("socket write error: %s", err.Error())
				// Backoff when error is temporary net error
				if ne, ok := err.(net.Error); ok {
					if ne.Temporary() {
						log.Debugf("socket error is temporary, backoff")
						time.Sleep(10 * time.Millisecond)
						c.send <- buf
						break
					}
				}
				return
			}
			if err := c.writer.Flush(); err != nil {
				log.Debugf("writer flush error: %s", err.Error())
				return
			}
		}
	}
}

func (c *Client) Ping() {
	if c.timeout != nil {
		c.timeout.Stop()
	}
	var to time.Duration
	if c.info.KeepAlive > 0 {
		to = time.Duration(c.info.KeepAlive)
	} else {
		to = time.Duration(defaultKeepAlive)
	}
	c.timeout = time.NewTicker(to * time.Second)
	pr, _ := message.NewPingResp().Encode()
	c.Send(pr)
}
