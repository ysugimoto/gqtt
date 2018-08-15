package client

import (
	"crypto/tls"
	"errors"
	"net"
	"net/url"
	"time"

	"github.com/satori/go.uuid"
	"github.com/ysugimoto/gqtt/internal/log"
	"github.com/ysugimoto/gqtt/message"
)

func connect(u string, opt *ClientOption) (net.Conn, *ServerInfo, error) {
	var (
		conn   net.Conn
		err    error
		parsed *url.URL
		info   *ServerInfo
	)

	parsed, err = url.Parse(u)
	if err != nil {
		return nil, nil, err
	}

	switch parsed.Scheme {
	case "mqtt":
		if conn, err = net.Dial("tcp", parsed.Host); err != nil {
			return nil, nil, err
		}
	case "mqtts":
		conf := &tls.Config{
			ServerName: parsed.Host,
		}
		if conn, err = tls.Dial("tcp", parsed.Host, conf); err != nil {
			return nil, nil, err
		}
	default:
		return nil, nil, errors.New("connection protocol must start with mqtt(s)://")
	}

	if info, err = handshake(conn, opt); err != nil {
		log.Debug("failed to handshake with server: ", err)
		conn.Close()
		return nil, nil, err
	}
	return conn, info, nil
}

func handshake(conn net.Conn, opt *ClientOption) (*ServerInfo, error) {
	// if opt == nil {
	// 	opt = NewClientOption()
	// }
	connect := message.NewConnect()
	connect.ClientId = uuid.NewV4().String()

	packet, err := connect.Encode()
	if err != nil {
		log.Debug("failed to encode CONNECT packet: ", err)
		return nil, err
	}
	if _, err := conn.Write(packet); err != nil {
		log.Debug("failed to write packet: ", err)
		return nil, err
	}

	log.Debug("Client wrote connect packet, wait for CONNACK...")
	// Set read timeout
	conn.SetReadDeadline(time.Now().Add(10 * time.Second))
	defer conn.SetReadDeadline(time.Time{})

	frame, payload, err := message.ReceiveFrame(conn)
	if err != nil {
		log.Debug("failed to read CONNACK packet: ", err)
		return nil, err
	}
	switch frame.Type {
	case message.CONNACK:
		ack, err := message.ParseConnAck(frame, payload)
		if err != nil {
			log.Debug("packet parse error for CONNACK: ", err)
			return nil, err
		} else if ack.ReasonCode != message.Success {
			log.Debug("CONNACK doesn't reply success code: ", ack.ReasonCode)
			return nil, errors.New("CONNACK doesn't reply success code")
		} else {
			log.Debug("CONNACK received, clientId is ", connect.ClientId)
			return ack.Property, nil
		}
	case message.AUTH:
		log.Debug("not implement yet for AUTH packet")
		return nil, errors.New("not implement yet for AUTH packet")
	default:
		return nil, errors.New("unexpected frame received on handhshake")
	}
}
