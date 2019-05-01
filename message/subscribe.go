package message

import (
	"io"

	"github.com/pkg/errors"
)

type Subscribe struct {
	*Frame

	PacketId      uint16
	Property      *SubscribeProperty
	Subscriptions []SubscribeTopic
}

type SubscribeProperty struct {
	SubscriptionIdentifier uint64
	UserProperty           map[string]string
}

func (s *SubscribeProperty) ToProp() *Property {
	return &Property{
		SubscriptionIdentifier: s.SubscriptionIdentifier,
		UserProperty:           s.UserProperty,
	}
}

type SubscribeTopic struct {
	RetainHandling uint8
	RAP            bool
	NoLocal        bool
	TopicName      string
	QoS            QoSLevel
}

func ParseSubscribe(f *Frame, p []byte) (s *Subscribe, err error) {
	s = &Subscribe{
		Frame:         f,
		Subscriptions: make([]SubscribeTopic, 0),
	}

	dec := newDecoder(p)
	if s.PacketId, err = dec.Uint16(); err != nil {
		return nil, errors.Wrap(err, "failed to decode as uint16")
	}
	if prop, err := dec.Property(); err != nil {
		return nil, errors.Wrap(err, "failed to decode property")
	} else if prop != nil {
		s.Property = prop.ToSubscribe()
	}
	// payload for Topic filter + subscription options, ...
	for {
		var t string
		var b int
		if t, err = dec.String(); err != nil {
			if err == io.EOF {
				break
			}
			return nil, errors.Wrap(err, "failed to decode as string")
		}
		if b, err = dec.Int(); err != nil {
			return nil, errors.Wrap(err, "failed to decode as int")
		}
		st := SubscribeTopic{
			RetainHandling: uint8((b >> 4) & 0x03),
			RAP:            ((b >> 3) & 0x01) > 0,
			NoLocal:        (b >> 2 & 0x01) > 0,
			QoS:            QoSLevel((b & 0x03)),
			TopicName:      t,
		}
		s.Subscriptions = append(s.Subscriptions, st)
	}
	return s, nil
}

func NewSubscribe(opts ...option) *Subscribe {
	return &Subscribe{
		Frame:         newFrame(SUBSCRIBE, opts...),
		Subscriptions: make([]SubscribeTopic, 0),
	}
}

func (s *Subscribe) AddTopic(sts ...SubscribeTopic) {
	for _, st := range sts {
		s.Subscriptions = append(s.Subscriptions, st)
	}
}

func (s *Subscribe) Validate() error {
	if len(s.Subscriptions) == 0 {
		return errors.New("At least one topic should subscribe")
	}
	if s.PacketId == 0 {
		return errors.New("PacketId is required")
	}
	return nil
}

func (s *Subscribe) Encode() ([]byte, error) {
	if err := s.Validate(); err != nil {
		return nil, errors.Wrap(err, "SUBSCRIBE validation error")
	}

	enc := newEncoder()
	enc.Uint16(s.PacketId)
	if s.Property != nil {
		enc.Property(s.Property.ToProp())
	} else {
		enc.Uint(0)
	}

	eb := func(b bool) int {
		if b {
			return 1
		}
		return 0
	}
	for _, v := range s.Subscriptions {
		enc.String(v.TopicName)
		enc.Uint(uint8(int(v.RetainHandling)<<4 | eb(v.RAP)<<3 | eb(v.NoLocal)<<2 | int(v.QoS)))
	}

	return s.Frame.Encode(enc.Get()), nil
}
