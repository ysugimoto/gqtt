package message

import (
	"bytes"
	"fmt"
	"io"
	"sync"
)

var (
	decOne = sync.Pool{
		New: func() interface{} {
			return make([]byte, 1)
		},
	}
	decTwo = sync.Pool{
		New: func() interface{} {
			return make([]byte, 2)
		},
	}
	decFour = sync.Pool{
		New: func() interface{} {
			return make([]byte, 4)
		},
	}
)

type decoder struct {
	r *bytes.Reader
}

func newDecoder(p []byte) *decoder {
	return &decoder{
		r: bytes.NewReader(p),
	}
}

func (d *decoder) Int() (int, error) {
	b := decOne.Get().([]byte)
	defer func() {
		decOne.Put(b)
	}()
	if n, err := d.r.Read(b); err != nil {
		return 0, err
	} else if n != 1 {
		return 0, fmt.Errorf("decoder couldn't read expect bytes %d of 1", n)
	}
	return int(b[0]), nil
}

func (d *decoder) Uint() (uint8, error) {
	if i, err := d.Int(); err != nil {
		return 0, err
	} else {
		return uint8(i), nil
	}
}

func (d *decoder) Int16() (int, error) {
	b := decTwo.Get().([]byte)
	defer func() {
		decTwo.Put(b)
	}()
	if n, err := d.r.Read(b); err != nil {
		return 0, err
	} else if n != 2 {
		return 0, fmt.Errorf("decoder couldn't read expect bytes %d of 2", n)
	}
	return ((int(b[0]) << 8) | int(b[1])), nil
}

func (d *decoder) Uint16() (uint16, error) {
	if i, err := d.Int16(); err != nil {
		return 0, err
	} else {
		return uint16(i), nil
	}
}

func (d *decoder) Int32() (int, error) {
	b := decFour.Get().([]byte)
	defer func() {
		decFour.Put(b)
	}()
	if n, err := d.r.Read(b); err != nil {
		return 0, err
	} else if n != 4 {
		return 0, fmt.Errorf("decoder couldn't read expect bytes %d of 4", n)
	}
	return ((int(b[0]) << 24) | (int(b[1]) << 16) | (int(b[2]) << 8) | int(b[3])), nil
}

func (d *decoder) Uint32() (uint32, error) {
	if i, err := d.Int32(); err != nil {
		return 0, err
	} else {
		return uint32(i), nil
	}
}

func (d *decoder) String() (string, error) {
	if buf, err := d.Binary(); err != nil {
		return "", err
	} else {
		return string(buf), nil
	}
}

func (d *decoder) StringAll() (string, error) {
	if buf, err := d.BinaryAll(); err != nil {
		return "", err
	} else {
		return string(buf), nil
	}
}

func (d *decoder) Binary() ([]byte, error) {
	size, err := d.Int16()
	if err != nil {
		return nil, err
	}
	buf := make([]byte, size)
	if n, err := d.r.Read(buf); err != nil {
		return nil, err
	} else if n != size {
		return nil, fmt.Errorf("decoder couldn't read expect bytes %d of %d", n, size)
	}
	return buf, nil
}

func (d *decoder) BinaryAll() ([]byte, error) {
	remains := d.r.Len()
	if remains == 0 {
		return []byte{}, nil
	}
	buf := make([]byte, remains)
	if n, err := d.r.Read(buf); err != nil {
		return nil, err
	} else if n != remains {
		return nil, fmt.Errorf("decoder couldn't read expect bytes %d of %d", n, remains)
	}
	return buf, nil
}

func (d *decoder) Variable() (uint64, error) {
	var (
		size uint64
		mul  uint64 = 1
	)
	for {
		i, err := d.Int()
		if err != nil {
			return 0, err
		}
		size += uint64(i&0x7F) * mul
		mul *= 0x80
		if i&0x80 == 0 {
			break
		}
	}
	return size, nil
}

func (d *decoder) Property() (*Property, error) {
	size, err := d.Int()
	if err != nil {
		return nil, err
	} else if size == 0 {
		return nil, nil
	}
	prop := &Property{}
	p := make([]byte, size)
	if n, err := d.r.Read(p); err != nil {
		return nil, err
	} else if n != size {
		return nil, fmt.Errorf("decoder couldn't read expect bytes %d of %d", n, size)
	}

	dd := newDecoder(p)
	for {
		sig, err := dd.Uint()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		switch PropertyType(sig) {
		case PayloadFormatIndicator:
			if prop.PayloadFormatIndicator, err = dd.Uint(); err != nil {
				return nil, err
			}
		case MessageExpiryInterval:
			if prop.MessageExpiryInterval, err = dd.Uint32(); err != nil {
				return nil, err
			}
		case ContentType:
			if prop.ContentType, err = dd.String(); err != nil {
				return nil, err
			}
		case ResponseTopic:
			if prop.ResponseTopic, err = dd.String(); err != nil {
				return nil, err
			}
		case CorrelationData:
			if prop.CorrelationData, err = dd.Binary(); err != nil {
				return nil, err
			}
		case SubscriptionIdentifier:
			if prop.SubscriptionIdentifier, err = dd.Variable(); err != nil {
				return nil, err
			}
		case SessionExpiryInterval:
			if prop.SessionExpiryInterval, err = dd.Uint32(); err != nil {
				return nil, err
			}
		case AssignedClientIdentifier:
			if prop.AssignedClientIdentifier, err = dd.String(); err != nil {
				return nil, err
			}
		case ServerKeepAlive:
			if prop.ServerKeepAlive, err = dd.Uint16(); err != nil {
				return nil, err
			}
		case AuthenticationMethod:
			if prop.AuthenticationMethod, err = dd.String(); err != nil {
				return nil, err
			}
		case AuthenticationData:
			if prop.AuthenticationData, err = dd.Binary(); err != nil {
				return nil, err
			}
		case RequestProblemInformation:
			if i, err := dd.Int(); err != nil {
				return nil, err
			} else {
				prop.RequestProblemInformation = i > 0
			}
		case WillDelayInterval:
			if prop.WillDelayInterval, err = dd.Uint32(); err != nil {
				return nil, err
			}
		case RequestResponseInformation:
			if i, err := dd.Int(); err != nil {
				return nil, err
			} else {
				prop.RequestResponseInformation = i > 0
			}
		case ResponseInformation:
			if prop.ResponseInformation, err = dd.String(); err != nil {
				return nil, err
			}
		case ServerReference:
			if prop.ServerReference, err = dd.String(); err != nil {
				return nil, err
			}
		case ReasonString:
			if prop.ReasonString, err = dd.String(); err != nil {
				return nil, err
			}
		case ReceiveMaximum:
			if prop.ReceiveMaximum, err = dd.Uint16(); err != nil {
				return nil, err
			}
		case TopicAliasMaximum:
			if prop.TopicAliasMaximum, err = dd.Uint16(); err != nil {
				return nil, err
			}
		case TopicAlias:
			if prop.TopicAlias, err = dd.Uint16(); err != nil {
				return nil, err
			}
		case MaximumQoS:
			if prop.MaximumQoS, err = dd.Uint(); err != nil {
				return nil, err
			}
		case RetainAvalilable:
			if i, err := dd.Int(); err != nil {
				return nil, err
			} else {
				prop.RetainAvalilable = i > 0
			}
		case UserProperty:
			if key, err := dd.String(); err != nil {
				return nil, err
			} else if val, err := dd.String(); err != nil {
				return nil, err
			} else {
				if prop.UserProperty == nil {
					prop.UserProperty = make(map[string]string)
				}
				prop.UserProperty[key] = val
			}
		case MaximumPacketSize:
			if prop.MaximumPacketSize, err = dd.Uint32(); err != nil {
				return nil, err
			}
		case WildcardSubscriptionAvailable:
			if i, err := dd.Int(); err != nil {
				return nil, err
			} else {
				prop.WildcardSubscriptionAvailable = i > 0
			}
		case SubscrptionIdentifierAvailable:
			if i, err := dd.Int(); err != nil {
				return nil, err
			} else {
				prop.SubscrptionIdentifierAvailable = i > 0
			}
		case SharedSubscriptionsAvaliable:
			if i, err := dd.Int(); err != nil {
				return nil, err
			} else {
				prop.SharedSubscriptionsAvaliable = i > 0
			}
		}
	}
	return prop, nil
}
