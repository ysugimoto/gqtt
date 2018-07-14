package broker

import (
	"context"
	"log"
	"net"
	"regexp"
	"strings"
	"sync"
	"time"
)

func formatTopicPath(path string) string {
	return "/" + strings.Trim(path, "/")
}

type Client struct {
	ctx     context.Context
	conn    net.Conn
	id      string
	timeout time.Time

	topics map[string]uint8
	mu     sync.Mutex
	once   sync.Once

	send chan []byte
}

func NewClient(conn net.Conn, ctx context.Context, id string) *Client {
	return &Client{
		ctx:    ctx,
		id:     id,
		conn:   conn,
		topics: make(map[string]uint8),
	}
}

func (c *Client) loop() {
	for {
		select {
		case <-c.ctx.Done():
			c.Close()
		case b := <-c.send:
			c.write(b)
		}
	}
}

func (c *Client) write(b []byte) {
}

func (c *Client) Close() {
	c.once.Do(func() {
		c.conn.Close()
	})
}

func (c *Client) Id() string {
	return c.id
}

func (c *Client) Send(m []byte) {
	c.send <- m
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
