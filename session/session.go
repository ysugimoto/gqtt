package session

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/ysugimoto/gqtt/internal/log"
	"github.com/ysugimoto/gqtt/message"
)

const MaxRetries = 5

type sessionData struct {
	messageType message.MessageType
	channel     chan interface{}
}

type Session struct {
	stack         sync.Map
	storedMessage sync.Map
	ctx           context.Context
	conn          net.Conn
	mu            sync.Mutex
	isRunning     bool
}

func New(conn net.Conn, ctx context.Context) *Session {
	return &Session{
		ctx:           ctx,
		conn:          conn,
		stack:         sync.Map{},
		storedMessage: sync.Map{},
	}
}

func (s *Session) recoverError(err error) error {
	if s.isRunning {
		log.Debugf("write error occured, but keep connection due to session is running: %s", err.Error())
		return nil
	}
	return errors.Wrap(err, "failed to recover error")
}

func (s *Session) Write(msg message.Encoder) error {
	s.mu.Lock()
	defer func() {
		// This is a trick for sending mutex blocked message queue properly.
		// Wait a tiny microseconds before unlock mutex, it makes socket accepts to write next packet again.
		time.Sleep(100 * time.Microsecond)
		s.mu.Unlock()
	}()

	buf, err := msg.Encode()
	if err != nil {
		log.Debug("failed to encode message ", err)
		return s.recoverError(errors.Wrap(err, "failed to encode message"))
	}

	if n, err := s.conn.Write(buf); err != nil {
		log.Debug("failed to write packet: ", buf)
		return s.recoverError(errors.Wrap(err, "failed to write packet"))
	} else if n != len(buf) {
		log.Debug("could not enough patck")
		return s.recoverError(errors.New("could not write enough packet"))
	}

	log.Debug("-->> ", msg.GetType())
	return nil
}

func (s *Session) Start(ident uint16, meet message.MessageType, msg message.Encoder, retry int) (interface{}, error) {
	data := sessionData{
		messageType: meet,
		channel:     make(chan interface{}),
	}
	// sync.Map is safety for threading, but not relaiable on too fast store/delete.
	// So we guard again by mutex
	s.stack.Store(ident, data)
	// time.Sleep( * time.Second)

	log.Debug("session stored for ident: ", ident)
	ctx, timeout := context.WithTimeout(s.ctx, 10*time.Second)
	defer func() {
		timeout()
		s.isRunning = false
	}()
	s.isRunning = true

	// Send message
	if err := message.WriteFrame(s.conn, msg); err != nil {
		return nil, errors.Wrap(err, "failed to send message")
	}
	// wait or timeout
	select {
	case <-ctx.Done():
		retry--
		if retry < 0 {
			return nil, errors.Wrap(ctx.Err(), "max retry times exceeded")
		}
		log.Debugf("Session retry for type: %s\n", meet)
		time.Sleep(3 * time.Second)
		msg.Duplicate()
		return s.Start(ident, meet, msg, retry)
	case ack := <-data.channel:
		return ack, nil
	}
}

func (s *Session) Meet(ident uint16, meet message.MessageType, msg interface{}) error {
	v, ok := s.stack.Load(ident)
	if !ok {
		return errors.New("session not found for ident: " + fmt.Sprint(ident))
	}
	defer s.stack.Delete(ident)
	log.Debugf("stack deleted for ident: %d", ident)
	data := v.(sessionData)
	if data.messageType != meet {
		return fmt.Errorf("session found, but unexpected message type: %s", data.messageType.String())
	}
	data.channel <- msg
	return nil
}

func (s *Session) StoreMessage(pb *message.Publish) {
	s.storedMessage.Store(pb.PacketId, pb)
	log.Debug("[session] message stored")
}

func (s *Session) LoadMessage(packetId uint16) (*message.Publish, bool) {
	v, ok := s.storedMessage.Load(packetId)
	if !ok {
		log.Debug("[session] message load failed")
		return nil, false
	}
	log.Debug("[session] message load success")
	pb, ok := v.(*message.Publish)
	return pb, ok
}

func (s *Session) DeleteMessage(packetId uint16) {
	log.Debug("[session] message delete")
	if _, ok := s.storedMessage.Load(packetId); ok {
		s.storedMessage.Delete(packetId)
	}
}
