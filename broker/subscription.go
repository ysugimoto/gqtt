package broker

import (
	"github.com/ysugimoto/gqtt/internal/log"
	"github.com/ysugimoto/gqtt/message"
	"sync"
)

type Subscription struct {
	topics map[string]map[string]struct{}
	mu     sync.Mutex
}

func NewSubscription() *Subscription {
	return &Subscription{
		topics: map[string]map[string]struct{}{},
	}
}

func (s *Subscription) RemoveAll(clientId string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for n, cs := range s.topics {
		if _, ok := cs[clientId]; ok {
			delete(s.topics[n], clientId)
		}
	}
}

func (s *Subscription) Unsubscribe(clientId, topic string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for n, cs := range s.topics {
		if n != topic {
			continue
		}
		if _, ok := cs[clientId]; ok {
			delete(s.topics[n], clientId)
		}
	}
}

func (s *Subscription) Add(clientId string, t message.SubscribeTopic) message.ReasonCode {
	s.mu.Lock()
	defer s.mu.Unlock()

	if v, ok := s.topics[t.TopicName]; ok {
		v[clientId] = struct{}{}
	} else {
		s.topics[t.TopicName] = map[string]struct{}{
			clientId: struct{}{},
		}
	}
	log.Debugf("topic added for %s, all: %+v\n", t.TopicName, s.topics[t.TopicName])
	switch t.QoS {
	case message.QoS0:
		return message.GrantedQoS0
	case message.QoS1:
		return message.GrantedQoS1
	case message.QoS2:
		return message.GrantedQoS2
	default:
		return message.UnspecifiedError
	}
}

func (s *Subscription) FindAll(topic string) []string {
	ids := make([]string, 0)
	s.mu.Lock()
	defer s.mu.Unlock()

	log.Debugf("find all clients fot topic: %s\n", topic)

	stack := map[string]struct{}{}
	for t, clients := range s.topics {
		if t != topic {
			continue
		}
		log.Debug("clients for topic: %s %+v\n", t, clients)
		for c, _ := range clients {
			if _, ok := stack[c]; ok {
				continue
			}
			log.Debugf("client found: %s\n", c)
			ids = append(ids, c)
			stack[c] = struct{}{}
		}
	}
	return ids
}
