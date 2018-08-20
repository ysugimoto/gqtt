package broker

import (
	"bytes"
	"fmt"
	"net"

	"encoding/base64"

	"github.com/ysugimoto/gqtt/internal/log"
	"github.com/ysugimoto/gqtt/message"
)

const (
	basicAuthentication = "basic"
	loginAuthentication = "login"
)

func write(conn net.Conn, m message.Encoder) error {
	buf, err := m.Encode()
	if err != nil {
		log.Debug("failed to encode auth packet ", err)
		return err
	}
	if n, err := conn.Write(buf); err != nil {
		log.Debug("failed to write auth packet ", err)
		return err
	} else if len(buf) != n {
		log.Debug("failed to write enough packet")
		return fmt.Errorf("failed to write enough packet")
	}
	return nil
}

// Do basic authentication on AUTH phase
func doBasicAuth(conn net.Conn, cp *message.ConnectProperty) error {
	dec, err := base64.StdEncoding.DecodeString(string(cp.AuthenticationData))
	if err != nil {
		return err
	}
	spl := bytes.SplitN(dec, []byte(":"), 2)
	// TODO: auth username and password
	if string(spl[0]) != "admin" || string(spl[1]) != "admin" {
		return fmt.Errorf("authentication failed for supplied user/pass")
	}
	log.Debug("[BASIC] username/password matched. Authentication success")
	if err := write(conn, message.NewAuth(message.Success)); err != nil {
		return err
	}
	return nil
}

// Do user / password authentication on AUTH phase
func doLoginAuth(conn net.Conn, cp *message.ConnectProperty) error {
	user := string(cp.AuthenticationData)
	log.Debugf("[LOGIN] user: %s", user)
	auth := message.NewAuth(message.ContinueAuthentication)
	if err := write(conn, auth); err != nil {
		return err
	}
	frame, payload, err := message.ReceiveFrame(conn)
	if err != nil {
		return err
	} else if frame.Type != message.AUTH {
		return fmt.Errorf("unexpected packet type received: %s", frame.Type.String())
	}
	auth, err = message.ParseAuth(frame, payload)
	if err != nil {
		return err
	}
	if auth.Property == nil {
		return fmt.Errorf("auth challenge data is required")
	}
	pass := string(auth.Property.AuthenticationData)
	if user != "admin" || pass != "admin" {
		return fmt.Errorf("authentication failed for supplied user/pass")
	}
	log.Debug("[LOGIN] username/password matched. Authentication success")
	if err := write(conn, message.NewAuth(message.Success)); err != nil {
		return err
	}
	return nil
}
