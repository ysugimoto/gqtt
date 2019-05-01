package client

import (
	"net"
	"time"

	"crypto/tls"
	"encoding/base64"
	"net/url"

	"github.com/pkg/errors"
	"github.com/satori/go.uuid"
	"github.com/ysugimoto/gqtt/internal/log"
	"github.com/ysugimoto/gqtt/message"
)

func makeConnectionMessage(opts []ClientOption) *message.Connect {
	connect := message.NewConnect()
	connect.ClientId = uuid.NewV4().String()

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
		case nameWill:
			v := o.value.(map[string]interface{})
			connect.FlagWill = true
			connect.WillQoS = v["qos"].(message.QoSLevel)
			connect.WillRetain = v["retain"].(bool)
			connect.WillTopic = v["topic"].(string)
			connect.WillPayload = v["payload"].(string)
			if v["property"] != nil {
				connect.WillProperty = v["property"].(*message.WillProperty)
			}
		}
	}
	if exists {
		connect.Property = p
	}
	return connect
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
			return nil, nil, errors.Wrap(err, "failed to dial TCP")
		}
	case "mqtts":
		conf := &tls.Config{
			ServerName: parsed.Host,
		}
		if conn, err = tls.Dial("tcp", parsed.Host, conf); err != nil {
			return nil, nil, errors.Wrap(err, "failed to dial TLS")
		}
	default:
		return nil, nil, errors.New("connection protocol must start with mqtt(s)://")
	}

	if info, err = handshake(conn, opts); err != nil {
		log.Debug("failed to handshake with server: ", err)
		conn.Close()
		return nil, nil, errors.Wrap(err, "failed to handshake with server")
	}
	return conn, info, nil
}

func handshake(conn net.Conn, opts []ClientOption) (*ServerInfo, error) {
	c := makeConnectionMessage(opts)
	if err := message.WriteFrame(conn, c); err != nil {
		log.Debug("failed to write CONNECT packet: ", err)
		return nil, errors.Wrap(err, "failed to write CONNECT packet")
	}

	log.Debug("Client wrote connect packet, wait for CONNACK...")
	// Set read timeout
	conn.SetReadDeadline(time.Now().Add(10 * time.Second))
	defer conn.SetReadDeadline(time.Time{})

	for {
		frame, payload, err := message.ReceiveFrame(conn)
		if err != nil {
			log.Debug("failed to read CONNACK packet: ", err)
			return nil, errors.Wrap(err, "failed to read CONNACK packet")
		}
		switch frame.Type {
		case message.CONNACK:
			ack, err := message.ParseConnAck(frame, payload)
			if err != nil {
				log.Debug("packet parse error for CONNACK: ", err)
				return nil, errors.Wrap(err, "packet parse error for CONNACK")
			} else if ack.ReasonCode != message.Success {
				log.Debug("CONNACK doesn't reply success code: ", ack.ReasonCode)
				return nil, errors.New("CONNACK doesn't reply success code")
			} else {
				log.Debug("CONNACK received, clientId is ", c.ClientId)
				return ack.Property, nil
			}
		case message.AUTH:
			auth, err := message.ParseAuth(frame, payload)
			if err != nil {
				log.Debug("packet parse error for AUTH: ", err)
				return nil, errors.Wrap(err, "packet parse error for AUTH")
			}
			switch auth.ReasonCode {
			case message.Success:
				log.Debug("Authentication success")
				// Authenticate succeed on broker
				time.Sleep(10 * time.Millisecond)
				continue
			case message.ContinueAuthentication:
				if err := authenticate(conn, c.Property); err != nil {
					return nil, errors.Wrap(err, "failed to authenticate")
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
		if err := message.WriteFrame(conn, auth); err != nil {
			log.Debug("failed to send AUTH packet: ", err)
			return errors.Wrap(err, "failed to send AUTH packet")
		}
	default:
		return errors.New("unexpected auth method: " + prop.AuthenticationMethod)
	}
	return nil
}
