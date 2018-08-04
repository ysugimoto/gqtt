package message

import (
	"errors"
)

type Publish struct {
	*Frame

	TopicName string
	MessageId uint16
	Body      []byte

	Property *PublishProperty
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

func (p *PublishProperty) Encode() []byte {
	return encodeProperty(&Property{
		PayloadFormatIndicator: p.PayloadFormatIndicator,
		MessageExpiryInterval:  p.MessageExpiryInterval,
		ContentType:            p.ContentType,
		ResponseTopic:          p.ResponseTopic,
		CorrelationData:        p.CorrelationData,
		SubscriptionIdentifier: p.SubscriptionIdentifier,
		TopicAlias:             p.TopicAlias,
		UserProperty:           p.UserProperty,
	})
}

func ParsePublish(f *Frame, p []byte) (*Publish, error) {
	pb := &Publish{
		Frame: f,
	}
	var i, psize int
	pb.TopicName, i = decodeString(p, i)
	pb.MessageId, i = decodeUint16(p, i)
	psize, i = decodeInt(p, i)
	if psize > 0 {
		prop, err := decodeProperty(p[i:(i + psize)])
		if err != nil {
			return nil, err
		}
		pb.Property = prop.ToPublish()
		i += psize
	}
	pb.Body = p[i:]

	return pb, nil
}

func NewPublish(opts ...option) *Publish {
	return &Publish{
		Frame: newFrame(PUBLISH, opts...),
	}
}

func (p *Publish) Validate() error {
	if p.TopicName == "" {
		return errors.New("TopicName is required")
	}
	if p.Frame.QoS > 0 && p.MessageId == 0 {
		return errors.New("MessageID is required when QoS is greater than 0")
	}
	return nil
}

func (p *Publish) Encode() ([]byte, error) {
	if err := p.Validate(); err != nil {
		return nil, err
	}
	payload := make([]byte, 0)
	payload = append(payload, encodeString(p.TopicName)...)
	payload = append(payload, encodeUint16(p.MessageId)...)
	if p.Property != nil {
		payload = append(payload, p.Property.Encode()...)
	}
	payload = append(payload, p.Body...)

	return p.Frame.Encode(payload), nil
}
