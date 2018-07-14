package broker

import (
	"bufio"
	"context"
	"log"
	"net"
	"regexp"
	"strings"
	"sync"
	"time"
)

type ConnectionState uint8

const (
	CONNECTING ConnectionState = iota
	CONNECTED
	CLOSED
)

type Client struct {
	State   ConnectionState
	ctx     context.Context
	conn    net.Conn
	id      string
	timeout time.Time

	topics map[string]uint8
	mu     sync.Mutex
	once   sync.Once

	send   chan []byte
	writer *bufio.Writer
}

func NewClient(conn net.Conn, ctx context.Context, id string) *Client {
	return &Client{
		State:  CONNECTING,
		ctx:    c,
		id:     id,
		conn:   conn,
		topics: make(map[string]uint8),
	}
}

func (c *Client) Read(b []byte) (int, error) {
	return c.conn.Read(b)
}

func (c *Client) Close() {
	c.once.Do(func() {
		c.conn.Close()
		c.State = CLOSED
	})
}

func (c *Client) Id() string {
	return c.id
}

func (c *Client) Send(m []byte) {
	if c.writer == nil {
		go c.writeLoop()
	}
	c.send <- m
}

func (c *Client) writeLoop() {
	c.writer = bufio.NewWriter(c.conn)
	for {
		select {
		case <-c.ctx.Done():
			c.Close()
			return
		case buf := <-c.send:
			if _, err := c.writer.Write(buf); err != nil {
				log.Printf("socket write error: %s", err.Error())
				// Backoff when error is temporary net error
				if ne, ok := err.(net.Error); ok {
					if ne.Temporary() {
						log.Printf("socket error is temporary, backoff")
						time.Sleep(10 * time.Millisecond)
						c.send <- buf
						break
					}
				}
				return
			}
			if err := c.writer.Flush(); err != nil {
				log.Printf("writer flush error: %s", err.Error())
				return
			}
		}
	}
}

func (c *Client) Sendable(topicName string) bool {
	topicName = strings.Replace(topicName, "+", ".+", -1)
	topicName = strings.Replace(topicName, "/#", ".*", -1)
	topicName += "$"

	r, err := regexp.Compile(topicName)
	if err != nil {
		log.Printf("failed to complie topicName to regexp: %s: %s\n", topicName, err.Error())
		return false
	}
	for t, _ := range c.topics {
		if r.MatchString(t) {
			return true
		}
	}
	return false
}

func (c *Client) Ping() {
	c.timeout = time.Now().Add(30 * time.Minute)
}

func (c *Client) Subscribe(topics ...string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, t := range topics {
		c.topics[formatTopicPath(t)] = 0 // temporary QoS 0
	}
}

func (c *Client) Unsubscribe(topics ...string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, t := range topics {
		t = formatTopicPath(t)
		if _, ok := c.topics[t]; ok {
			delete(c.topics, t)
		}
	}
}
