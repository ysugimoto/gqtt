package message

import (
	"errors"
)

type Publish struct {
	*Frame

	TopicName string
	PacketId  uint16
	Body      []byte

	Property *PublishProperty
}

// Downgrade QoS.
// Create new pointer in order to avoid unpexected copy data
func (p *Publish) Downgrade(qos QoSLevel) *Publish {
	buf := make([]byte, len(p.Body))
	copy(buf, p.Body)

	downgraded := &Publish{
		Frame:     newFrame(PUBLISH, WithQoS(qos)),
		PacketId:  p.PacketId,
		TopicName: p.TopicName,
		Body:      buf,
	}
	if p.Property != nil {
		downgraded.Property = &PublishProperty{
			PayloadFormatIndicator: p.Property.PayloadFormatIndicator,
			MessageExpiryInterval:  p.Property.MessageExpiryInterval,
			ContentType:            p.Property.ContentType,
			ResponseTopic:          p.Property.ResponseTopic,
			CorrelationData:        p.Property.CorrelationData,
			SubscriptionIdentifier: p.Property.SubscriptionIdentifier,
			TopicAlias:             p.Property.TopicAlias,
			UserProperty:           p.Property.UserProperty,
		}
	}
	return downgraded
}

type PublishProperty struct {
	PayloadFormatIndicator uint8
	MessageExpiryInterval  uint32
	ContentType            string
	ResponseTopic          string
	CorrelationData        []byte
	SubscriptionIdentifier uint64
	TopicAlias             uint16
	UserProperty           map[string]string
}

func (p *PublishProperty) ToProp() *Property {
	return &Property{
		PayloadFormatIndicator: p.PayloadFormatIndicator,
		MessageExpiryInterval:  p.MessageExpiryInterval,
		ContentType:            p.ContentType,
		ResponseTopic:          p.ResponseTopic,
		CorrelationData:        p.CorrelationData,
		SubscriptionIdentifier: p.SubscriptionIdentifier,
		TopicAlias:             p.TopicAlias,
		UserProperty:           p.UserProperty,
	}
}

func ParsePublish(f *Frame, p []byte) (pb *Publish, err error) {
	pb = &Publish{
		Frame: f,
	}
	dec := newDecoder(p)
	if pb.TopicName, err = dec.String(); err != nil {
		return nil, err
	}
	// PacketId exists only QoS is greater than 0
	if f.QoS > 0 {
		if pb.PacketId, err = dec.Uint16(); err != nil {
			return nil, err
		}
	}
	if prop, err := dec.Property(); err != nil {
		return nil, err
	} else if prop != nil {
		pb.Property = prop.ToPublish()
	}
	if pb.Body, err = dec.BinaryAll(); err != nil {
		return nil, err
	}
	return pb, nil
}

func NewPublish(packetId uint16, opts ...option) *Publish {
	return &Publish{
		Frame:    newFrame(PUBLISH, opts...),
		PacketId: packetId,
	}
}

func (p *Publish) Validate() error {
	if p.TopicName == "" {
		return errors.New("TopicName is required")
	}
	if p.Frame.QoS > 0 && p.PacketId == 0 {
		return errors.New("PacketId is required when QoS is greater than 0")
	}
	return nil
}

func (p *Publish) Encode() ([]byte, error) {
	if err := p.Validate(); err != nil {
		return nil, err
	}
	enc := newEncoder()
	enc.String(p.TopicName)
	if p.Frame.QoS > 0 {
		enc.Uint16(p.PacketId)
	}
	if p.Property != nil {
		enc.Property(p.Property.ToProp())
	} else {
		enc.Uint(0)
	}
	enc.BinaryAll(p.Body)

	return p.Frame.Encode(enc.Get()), nil
}
