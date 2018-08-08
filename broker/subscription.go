package broker

import (
	"github.com/ysugimoto/gqtt/internal/log"
	"github.com/ysugimoto/gqtt/message"
	"sync"
)

type Subscription struct {
	topics map[string][]string
	mu     sync.Mutex
}

func NewSubscription() *Subscription {
	return &Subscription{
		topics: map[string][]string{},
	}
}

func (s *Subscription) Add(clientId string, t message.SubscribeTopic) message.ReasonCode {
	if v, ok := s.topics[t.TopicName]; ok {
		v = append(v, clientId)
	} else {
		s.topics[t.TopicName] = []string{clientId}
	}
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
		for _, c := range clients {
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
