package broker

import (
	"context"
	"log"
	"sync"

	"github.com/ysugimoto/gqtt/message"
)

type Manager struct {
	clients map[string]*Client
	ctx     context.Context

	connect    chan *Client
	disconnect chan *Client

	subscribe   chan *Topic
	unsubscribe chan *Topic

	ping chan *Client

	mu sync.Mutex
}

func initManager(ctx context.Context) *Manager {
	m := &Manager{
		clients: make(map[string]*Client),
		ctx:     ctx,

		connect:    make(chan *Client),
		disconnect: make(chan *Client),

		subscribe:   make(chan *Topic),
		unsubscribe: make(chan *Topic),

		ping: make(chan *Client),
	}
	go func() {
		for {
			select {
			case c := <-m.connect:
				log.Printf("handle connect Client: %s", c.Id())
				m.handleConnect(c)
			case c := <-m.disconnect:
				log.Printf("handle disconnect Client: %s", c.Id())
				m.handleDisconnect(c)

			case t := <-m.subscribe:
				log.Printf("handle subscribe Topic: %+v", t.Names())
				m.handleSubscribe(t)
			case t := <-m.unsubscribe:
				log.Printf("handle unsubscribe Client: %+v", t.Names())
				m.handleUnsubscribe(t)

			case c := <-m.ping:
				log.Printf("handle ping Client: %s", c.Id())
				m.handlePing(c)
			}
		}
	}()
	return m
}

func (m *Manager) publish(sender string, pb *message.Publish) error {
	buf, err := pb.Encode()
	if err != nil {
		return err
	}
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, c := range m.clients {
		if sender == c.Id() {
			continue
		}
		if c.Sendable(pb.TopicName) {
			c.Send(buf)
		}
	}
	return nil
}

func (m *Manager) get(id string) *Client {
	if c, ok := m.clients[id]; ok {
		return c
	}
	return nil
}

func (m *Manager) handleConnect(c *Client) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.clients[c.Id()] = c
}

func (m *Manager) handleDisconnect(c *Client) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if c := m.get(c.Id()); c != nil {
		delete(m.clients, c.Id())
	}
}

func (m *Manager) handleSubscribe(t *Topic) {
	// Ensure Client is connected to broker
	if c := m.get(t.ClientId()); c == nil {
		log.Printf("[subscribe] Client not connected: %s", t.ClientId())
		return
	} else {
		c.Subscribe(t.Names()...)
	}
}

func (m *Manager) handleUnsubscribe(t *Topic) {
	// Ensure Client is connected to broker
	if c := m.get(t.ClientId()); c == nil {
		log.Printf("[unsubscribe] Client not connected: %s", t.ClientId())
		return
	} else {
		c.Unsubscribe(t.Names()...)
	}
}

func (m *Manager) handlePing(c *Client) {
	if c := m.get(c.id); c != nil {
		c.Ping()
	}
}
