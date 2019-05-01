package broker

import (
	"fmt"
	"regexp"
	"strings"
	"sync"

	"github.com/ysugimoto/gqtt/internal/log"
	"github.com/ysugimoto/gqtt/message"
)

type SubscriptionInfo struct {
	Clients       map[string]message.QoSLevel
	RetainMessage *message.Publish
}

type Subscription struct {
	topics sync.Map
}

func NewSubscription() *Subscription {
	return &Subscription{
		topics: sync.Map{},
	}
}

func (s *Subscription) UnsubscribeAll(clientId string) {
	s.topics.Range(func(k, v interface{}) bool {
		info := v.(*SubscriptionInfo)
		if _, ok := info.Clients[clientId]; ok {
			delete(info.Clients, clientId)
		}
		return true
	})
}

func (s *Subscription) Unsubscribe(clientId, topic string) {
	v, ok := s.topics.Load(topic)
	if !ok {
		return
	}
	info := v.(*SubscriptionInfo)
	if _, ok := info.Clients[clientId]; ok {
		delete(info.Clients, clientId)
	}
}

func (s *Subscription) Subscribe(clientId string, t message.SubscribeTopic) (message.ReasonCode, error) {
	targets, err := s.FindTopics(t.TopicName)
	if err != nil {
		return message.UnspecifiedError, err
	} else if len(targets) == 0 {
		// If target hasn't created yet, create new topic.
		// But, topic name has wildcard, we'll skip it
		if !strings.Contains(t.TopicName, "#") && !strings.Contains(t.TopicName, "+") {
			targets = append(targets, t.TopicName)
		} else {
			return message.NoSubscriptionExisted, nil
		}
	}

	for _, topic := range targets {
		var info *SubscriptionInfo
		if v, ok := s.topics.Load(topic); ok {
			info = v.(*SubscriptionInfo)
		} else {
			info = &SubscriptionInfo{
				Clients: make(map[string]message.QoSLevel),
			}
		}
		info.Clients[clientId] = t.QoS
		s.topics.Store(topic, info)
		log.Debugf("client %s subscribed top for %s\n", clientId, topic)
	}

	switch t.QoS {
	case message.QoS0:
		return message.GrantedQoS0, nil
	case message.QoS1:
		return message.GrantedQoS1, nil
	case message.QoS2:
		return message.GrantedQoS2, nil
	default:
		return message.QoSNotSupported, fmt.Errorf("Unexpected Qos")
	}
}

func (s *Subscription) CompileTopicRegex(topic string) (*regexp.Regexp, error) {
	// Validate multi-level wildcard (MLW) if exists
	if mlw := strings.Index(topic, "#"); mlw != -1 {
		// It's OK to use only MLW character
		if topic != "#" {
			// MLW must present after topic division character
			if topic[mlw-1] != '/' {
				return nil, fmt.Errorf("Multi-level wildcard must present after topic division chacater of `/`")
			}
			// MLW must present at last character of topic name
			if mlw != len(topic)-1 {
				return nil, fmt.Errorf("Multi-level wildcard must present at last character")
			}
		}
	}
	// Validate single-level wildcard (SLW) if exists
	if slw := strings.Index(topic, "+"); slw != -1 {
		// It's OK to use only SLW character
		if topic != "+" {
			// SLW must present after topic division character
			if topic[slw-1] != '/' {
				return nil, fmt.Errorf("Single-level wildcard must present after topic division chacater of `/`")
			}
		}
	}

	t := strings.Replace(topic, "+", "[^/]+", -1)
	t = strings.Replace(t, "/#", ".*", -1)
	t += "$"

	return regexp.Compile(t)
}

func (s *Subscription) FindTopics(topic string) ([]string, error) {
	// FIXME: Now we are using regexp in order to find matching topics easily.
	r, err := s.CompileTopicRegex(topic)
	if err != nil {
		log.Debug("failed to compile regex: ", topic, err)
		return nil, err
	}
	topics := []string{}
	s.topics.Range(func(k, v interface{}) bool {
		topicName := k.(string)
		if r.MatchString(topicName) {
			topics = append(topics, topicName)
		}
		return true
	})
	return topics, nil
}

func (s *Subscription) GetClientsByTopic(topic string) *SubscriptionInfo {
	log.Debugf("find all clients fot topic: %s\n", topic)

	if v, ok := s.topics.Load(topic); ok {
		return v.(*SubscriptionInfo)
	}
	return nil
}
