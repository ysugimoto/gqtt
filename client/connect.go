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

func connect(u *url.URL, opt *ClientOption) (net.Conn, *ServerInfo, error) {
	var (
		conn net.Conn
		err  error
	)
	if u.Scheme == "mqtts" {
		conn, err = tls.Dial("tcp", u.Host, &tls.Config{
			ServerName: u.Host,
		})
	} else {
		conn, err = net.Dial("tcp", u.Host)
	}
	if err != nil {
		return nil, nil, err
	}

	if info, err := handshake(conn, opt); err != nil {
		log.Debug("failed to handshake with server: ", err)
		conn.Close()
		return nil, nil, err
	} else {
		return conn, info, nil
	}
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
		break
	}
	return nil, errors.New("unexpected frame received on handhshake")
}
