package broker

import (
	"bytes"
	"net"

	"encoding/base64"

	"github.com/pkg/errors"
	"github.com/ysugimoto/gqtt/internal/log"
	"github.com/ysugimoto/gqtt/message"
)

const (
	basicAuthentication = "basic"
	loginAuthentication = "login"
)

// Do basic authentication on AUTH phase
func doBasicAuth(conn net.Conn, cp *message.ConnectProperty) error {
	dec, err := base64.StdEncoding.DecodeString(string(cp.AuthenticationData))
	if err != nil {
		return errors.Wrap(err, "failed to decode basic auth string")
	}
	spl := bytes.SplitN(dec, []byte(":"), 2)
	// TODO: auth username and password
	if string(spl[0]) != "admin" || string(spl[1]) != "admin" {
		return errors.New("authentication failed for supplied user/pass")
	}
	log.Debug("[BASIC] username/password matched. Authentication success")
	if err := message.WriteFrame(conn, message.NewAuth(message.Success)); err != nil {
		return errors.Wrap(err, "failed to write auth challenge frame")
	}
	return nil
}

// Do user / password authentication on AUTH phase
func doLoginAuth(conn net.Conn, cp *message.ConnectProperty) error {
	user := string(cp.AuthenticationData)
	log.Debugf("[LOGIN] user: %s", user)
	auth := message.NewAuth(message.ContinueAuthentication)
	if err := message.WriteFrame(conn, auth); err != nil {
		return errors.Wrap(err, "failed to write auth challenge frame")
	}
	frame, payload, err := message.ReceiveFrame(conn)
	if err != nil {
		return errors.Wrap(err, "failed to receive frame")
	} else if frame.Type != message.AUTH {
		return errors.New("unexpected packet type received: " + frame.Type.String())
	}
	auth, err = message.ParseAuth(frame, payload)
	if err != nil {
		return errors.Wrap(err, "failed to parse as Auth packet")
	}
	if auth.Property == nil {
		return errors.New("auth challenge data is required")
	}
	pass := string(auth.Property.AuthenticationData)
	if user != "admin" || pass != "admin" {
		return errors.New("authentication failed for supplied user/pass")
	}
	log.Debug("[LOGIN] username/password matched. Authentication success")
	if err := message.WriteFrame(conn, message.NewAuth(message.Success)); err != nil {
		return errors.Wrap(err, "failed to write auth success packet")
	}
	return nil
}
