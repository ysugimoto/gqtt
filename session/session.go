package session

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/ysugimoto/gqtt/internal/log"
	"github.com/ysugimoto/gqtt/message"
)

const MaxRetries = 5

type sessionData struct {
	messageType message.MessageType
	channel     chan interface{}
}

type Session struct {
	stack sync.Map
	ctx   context.Context
	conn  net.Conn
	mu    sync.Mutex
}

func New(conn net.Conn, ctx context.Context) *Session {
	return &Session{
		ctx:   ctx,
		conn:  conn,
		stack: sync.Map{},
	}
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
		return err
	}

	if n, err := s.conn.Write(buf); err != nil {
		log.Debug("failed to write packet: ", buf)
		return err
	} else if n != len(buf) {
		log.Debug("could not enough patck")
		return errors.New("could not write enough packet")
	}
	return nil
}

func (s *Session) Read(p []byte) (n int, err error) {
	return s.conn.Read(p)
}

func (s *Session) Start(ident uint16, meet message.MessageType, msg message.Encoder, retry int) (interface{}, error) {
	data := sessionData{
		messageType: meet,
		channel:     make(chan interface{}),
	}
	s.stack.Store(ident, data)
	ctx, timeout := context.WithTimeout(s.ctx, 10*time.Second)
	defer timeout()

	for {
		// Send message
		if err := s.Write(msg); err != nil {
			return nil, err
		}
		// wait or timeout
		select {
		case <-ctx.Done():
			retry--
			if retry < 0 {
				return nil, ctx.Err()
			}
			msg.Duplicate()
			return s.Start(ident, meet, msg, retry)
		case ack := <-data.channel:
			return ack, nil
		}
	}
}

func (s *Session) Meet(ident uint16, meet message.MessageType, msg interface{}) error {
	v, ok := s.stack.Load(ident)
	if !ok {
		return fmt.Errorf("session not found for ident: %d", ident)
	}
	defer s.stack.Delete(ident)
	data := v.(sessionData)
	if data.messageType != meet {
		return fmt.Errorf("session found, but unexpected message type")
	}
	data.channel <- msg
	return nil
}
