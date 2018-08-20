package client

import (
	"crypto/tls"
	"encoding/base64"
	"errors"
	"net"
	"net/url"
	"time"

	"github.com/satori/go.uuid"
	"github.com/ysugimoto/gqtt/internal/log"
	"github.com/ysugimoto/gqtt/message"
)

func makeConnectionProperty(opts []ClientOption) *message.ConnectProperty {
	log.Debugf("option data %+v", opts)
	var exists bool
	p := &message.ConnectProperty{}
	for _, o := range opts {
		log.Debugf("option data %+v", o)
		switch o.name {
		case nameBasicAuth:
			p.AuthenticationMethod = "basic"
			v := o.value.(map[string]string)
			d := base64.StdEncoding.EncodeToString([]byte(v["user"] + ":" + v["pass"]))
			p.AuthenticationData = []byte(d)
			exists = true
		case nameLoginAuth:
			p.AuthenticationMethod = "login"
			v := o.value.(map[string]string)
			p.AuthenticationData = []byte(v["user"])
			p.ChallengeData = v
			exists = true
		}
	}
	if !exists {
		return nil
	}
	return p
}

func connect(u string, opts []ClientOption) (net.Conn, *ServerInfo, error) {
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

	if info, err = handshake(conn, opts); err != nil {
		log.Debug("failed to handshake with server: ", err)
		conn.Close()
		return nil, nil, err
	}
	return conn, info, nil
}

func handshake(conn net.Conn, opts []ClientOption) (*ServerInfo, error) {
	// if opt == nil {
	// 	opt = NewClientOption()
	// }
	connect := message.NewConnect()
	connect.ClientId = uuid.NewV4().String()
	if len(opts) > 0 {
		connect.Property = makeConnectionProperty(opts)
	}

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

	for {
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
			auth, err := message.ParseAuth(frame, payload)
			if err != nil {
				log.Debug("packet parse error for AUTH: ", err)
				return nil, err
			}
			switch auth.ReasonCode {
			case message.Success:
				// Authenticate succeed on broker
				continue
			case message.ContinueAuthentication:
				if err := authenticate(conn, connect.Property); err != nil {
					return nil, err
				}
			}
		default:
			return nil, errors.New("unexpected frame received on handhshake")
		}
	}
}

func authenticate(conn net.Conn, prop *message.ConnectProperty) error {
	switch prop.AuthenticationMethod {
	case "login":
		auth := message.NewAuth(message.ContinueAuthentication)
		cd := prop.ChallengeData.(map[string]string)
		auth.Property = &message.AuthProperty{
			AuthenticationMethod: "login",
			AuthenticationData:   []byte(cd["pass"]),
		}
		packet, err := auth.Encode()
		if err != nil {
			log.Debug("failed to encode AUTH packet: ", err)
			return err
		}
		if _, err := conn.Write(packet); err != nil {
			log.Debug("failed to write packet: ", err)
			return err
		}
	default:
		return errors.New("unexpected auth method: " + prop.AuthenticationMethod)
	}
	return nil
}
