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

	subscribe   chan *message.Subscribe
	unsubscribe chan *message.Unsubscribe
	publish     chan *message.Publish

	ping       chan *Client
	disconnect chan *Client
	connect    chan *Client

	mu sync.Mutex
}

func initManager(ctx context.Context) *Manager {
	m := &Manager{
		clients: make(map[string]*Client),
		ctx:     ctx,

		subscribe:   make(chan *message.Subscribe),
		unsubscribe: make(chan *message.Unsubscribe),
		publish:     make(chan *message.Publish),

		ping:       make(chan *Client),
		disconnect: make(chan *Client),
		connect:    make(chan *Client),
	}
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
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

			case pb := <-m.publish:
				log.Printf("handle publish from Client: %+v", pb)

			case c := <-m.ping:
				log.Printf("handle ping Client: %s", c.Id())
				m.handlePing(c)
			}
		}
	}()
	return m
}

func (m *Manager) publishMessage(sender string, pb *message.Publish) error {
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
