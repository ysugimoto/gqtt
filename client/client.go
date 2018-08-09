package client

import (
	"bufio"
	"context"
	"errors"
	"github.com/ysugimoto/gqtt/internal/log"
	"github.com/ysugimoto/gqtt/message"
	"net"
	"net/url"
	"sync"
	"time"
)

type ClientOption = message.ConnectProperty
type ServerInfo = message.ConnAckProperty

type session struct {
	Type    message.MessageType
	Channel chan interface{}
}

func NewClientOption() *ClientOption {
	return &ClientOption{}
}

type Client struct {
	pid          uint16
	url          string
	conn         net.Conn
	ctx          context.Context
	terminate    context.CancelFunc
	incoming     chan *message.Packet
	sessions     map[uint16]session
	outgoing     chan []byte
	writer       *bufio.Writer
	pingInterval *time.Ticker

	Closed  chan struct{}
	Message chan *message.Publish

	once       sync.Once
	ServerInfo *ServerInfo
	mu         sync.Mutex
}

func NewClient(u string) *Client {
	return &Client{
		url: u,
	}
}

func (c *Client) Connect(ctx context.Context) error {
	return c.ConnectWithOption(ctx, nil)
}

func (c *Client) ConnectWithOption(ctx context.Context, opt *ClientOption) error {
	parsed, err := url.Parse(c.url)
	if err != nil {
		return err
	} else if parsed.Scheme != "mqtt" && parsed.Scheme != "mqtts" {
		return errors.New("connection protocol must start with mqtt(s)://")
	}

	conn, si, err := connect(parsed, opt)
	if err != nil {
		return err
	}

	log.Debug("connection established!")

	c.conn = conn
	c.ServerInfo = si
	c.ctx, c.terminate = context.WithCancel(ctx)
	c.incoming = make(chan *message.Packet)
	c.outgoing = make(chan []byte)
	c.Closed = make(chan struct{})
	c.sessions = make(map[uint16]session)
	c.Message = make(chan *message.Publish)

	// TODO: set duration from option
	go func() {
		c.pingInterval = time.NewTicker(5 * time.Second)
		for {
			<-c.pingInterval.C
			log.Debug("ping interval is coming. send ping to server")
			ping, _ := message.NewPingReq().Encode()
			c.outgoing <- ping
		}
	}()

	go func() {
		defer c.terminate()
		for {
			frame, payload, err := message.ReceiveFrame(c.conn)
			if err != nil {
				log.Debug("failed to receive message: ", err)
				if nerr, ok := err.(net.Error); ok {
					if nerr.Temporary() {
						log.Debug("buffer is temporary, backoff")
						continue
					}
				}
				c.terminate()
				return
			}
			log.Debugf("frame received: %+v | %+v\n", frame.Type, payload)
			c.incoming <- message.NewPacket(frame, payload)
		}
	}()
	go c.mainLoop()

	return nil
}

func (c *Client) Disconnect() {
	c.once.Do(func() {
		log.Debug("============================ Client closing =======================")
		c.pingInterval.Stop()
		close(c.outgoing)
		close(c.incoming)

		dc, err := message.NewDisconnect(message.NormalDisconnection).Encode()
		if err != nil {
			log.Debug("failed to encode DISCONNECT message: ", err)
		} else {
			c.sendPacket(dc)
			log.Debug("Send DISCONNECT")
		}
		log.Debug("Closing connection")
		c.conn.Close()
		log.Debug("connection closed, send channel")
		c.Closed <- struct{}{}
		log.Debug("channel sent")
	})
}

func (c *Client) sendPacket(m []byte) error {
	if c.writer == nil {
		c.writer = bufio.NewWriter(c.conn)
	}

	c.mu.Lock()
	defer func() {
		time.Sleep(100 * time.Microsecond)
		c.mu.Unlock()
	}()

	if _, err := c.writer.Write(m); err != nil {
		log.Debug("failed to write packet: ", m)
		return err
	} else if err := c.writer.Flush(); err != nil {
		log.Debug("failed to flush packet: ", m)
		return err
	}
	return nil
}

func (c *Client) mainLoop() {
	for {
		select {
		case <-c.ctx.Done():
			log.Debugf("terminated")
			c.Disconnect()
			return
		case packet := <-c.outgoing:
			if err := c.sendPacket(packet); err != nil {
				log.Debug("failed to write packet: ", err)
				c.terminate()
			}
		case packet := <-c.incoming:
			switch packet.Frame.Type {
			case message.PINGRESP:
				pr, err := message.ParsePingReq(packet.Frame, packet.Payload)
				if err != nil {
					log.Debug("malformed packet: failed to decode to PINGREQ packet: ", err)
					continue
				}
				log.Debugf("PINGRESP received: %+v\n", pr)
			case message.SUBACK:
				ack, err := message.ParseSubAck(packet.Frame, packet.Payload)
				if err != nil {
					log.Debug("malformed packet: failed to decode to SUBACK packet: ", err)
					continue
				}
				sess, ok := c.sessions[ack.PacketId]
				if !ok || sess.Type != message.SUBACK {
					log.Debug("malformed packet: unexpected packet identifier received")
					continue
				}
				sess.Channel <- ack
				delete(c.sessions, ack.PacketId)
			case message.PUBACK:
				ack, err := message.ParsePubAck(packet.Frame, packet.Payload)
				if err != nil {
					log.Debug("malformed packet: failed to decode to PUBACK packet: ", err)
					continue
				}
				sess, ok := c.sessions[ack.PacketId]
				if !ok || sess.Type != message.PUBACK {
					log.Debug("malformed packet: unexpected packet identifier received")
					continue
				}
				sess.Channel <- ack
				delete(c.sessions, ack.PacketId)
			case message.PUBLISH:
				log.Debug("PUBLISH message received")
				pb, err := message.ParsePublish(packet.Frame, packet.Payload)
				if err != nil {
					log.Debug("malformed packet: failed to decode to PUBLISH packet: ", err)
					continue
				}
				c.Message <- pb
			case message.PUBREC:
				ack, err := message.ParsePubRec(packet.Frame, packet.Payload)
				if err != nil {
					log.Debug("malformed packet: failed to decode to PUBREC packet: ", err)
					continue
				}
				sess, ok := c.sessions[ack.PacketId]
				if !ok || sess.Type != message.PUBREC {
					log.Debug("malformed packet: unexpected packet identifier received")
					continue
				}
				sess.Channel <- ack
				delete(c.sessions, ack.PacketId)
			case message.PUBCOMP:
				ack, err := message.ParsePubComp(packet.Frame, packet.Payload)
				if err != nil {
					log.Debug("malformed packet: failed to decode to PUBCOMP packet: ", err)
					continue
				}
				sess, ok := c.sessions[ack.PacketId]
				if !ok || sess.Type != message.PUBCOMP {
					log.Debug("malformed packet: unexpected packet identifier received")
					continue
				}
				sess.Channel <- ack
				delete(c.sessions, ack.PacketId)
			}
		}
	}
}

func (c *Client) makePacketId() uint16 {
	if c.pid == 0xFFFF {
		c.pid = 0
	}
	c.pid++
	return c.pid
}

func (c *Client) runSession(packetId uint16, mt message.MessageType, e message.Encoder) (interface{}, error) {
	recv := make(chan interface{})
	c.sessions[packetId] = session{
		Type:    mt,
		Channel: recv,
	}
	ctx, timeout := context.WithTimeout(c.ctx, 10*time.Second)
	defer timeout()

	// Send encoded packet
	if buf, err := e.Encode(); err != nil {
		return nil, err
	} else {
		c.outgoing <- buf
	}
	// wait or timeout
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case ack := <-recv:
		return ack, nil
	}
	return nil, errors.New("unpecxted logic")
}

func (c *Client) Subscribe(topic string, qos message.QoSLevel) error {
	packetId := c.makePacketId()

	ss := message.NewSubscribe()
	ss.PacketId = packetId
	// TODO: enable to set Retain handling, NoLocal flag
	ss.AddTopic(message.SubscribeTopic{
		TopicName: topic,
		QoS:       qos,
	})

	if ack, err := c.runSession(packetId, message.SUBACK, ss); err != nil {
		log.Debug("failed to finish session: ", err)
		return err
	} else if _, ok := ack.(*message.SubAck); !ok {
		log.Debug("unexpected ack received")
		return errors.New("unpected ack received")
	}
	return nil
}

func (c *Client) Publish(topic string, qos message.QoSLevel, body []byte) error {
	switch qos {
	case message.QoS0:
		// If OoS is zero, we don't need packet identifier and any acknowledgment
		p := message.NewPublish(0, message.WithQoS(qos))
		p.TopicName = topic
		p.Body = body
		if buf, err := p.Encode(); err != nil {
			log.Debug("failed to encode publish with QoS0 ", err)
			return err
		} else {
			c.outgoing <- buf
		}
	case message.QoS1:
		packetId := c.makePacketId()
		p := message.NewPublish(packetId, message.WithQoS(qos))
		p.TopicName = topic
		p.Body = body
		if ack, err := c.runSession(packetId, message.PUBACK, p); err != nil {
			log.Debug("failed to publish session for OoS1: ", err)
			return err
		} else if _, ok := ack.(*message.PubAck); !ok {
			log.Debug("failed to type conversion for OoS1")
			return errors.New("failed to type conversion for OoS1")
		}
		// TODO: Need to save and delete message for QoS1
	case message.QoS2:
		packetId := c.makePacketId()
		p := message.NewPublish(packetId, message.WithQoS(qos))
		p.TopicName = topic
		p.Body = body
		if ack, err := c.runSession(packetId, message.PUBREC, p); err != nil {
			log.Debug("failed to publish session for OoS2: ", err)
			return err
		} else if _, ok := ack.(*message.PubRec); !ok {
			log.Debug("failed to type conversion fto PUBREC or OoS2")
			return errors.New("failed to type conversion fto PUBREC or OoS2")
		}
		// On QoS2, need to send more packet for PUBREL
		pl := message.NewPubRel(packetId)
		if ack, err := c.runSession(packetId, message.PUBCOMP, pl); err != nil {
			log.Debug("failed to pubrel session for OoS2: ", err)
			return err
		} else if _, ok := ack.(*message.PubComp); !ok {
			log.Debug("failed to type conversion to PUBCOMP for OoS2")
			return errors.New("failed to type conversion to PUBCOMP for OoS2")
		}
	}
	return nil
}

/*
import (
	"fmt"
	//import the Paho Go MQTT library
	MQTT "github.com/eclipse/paho.mqtt.golang"
	"log"
	"os"
	"time"
)

//define a function for the default message handler
var f MQTT.MessageHandler = func(client MQTT.Client, msg MQTT.Message) {
	fmt.Printf("TOPIC: %s\n", msg.Topic())
	fmt.Printf("MSG: %s\n", msg.Payload())
}

func main() {
	MQTT.DEBUG = log.New(os.Stdout, "", 0)
	//create a ClientOptions struct setting the broker address, clientid, turn
	//off trace output and set the default message handler
	opts := MQTT.NewClientOptions().AddBroker("tcp://localhost:9999")
	opts.SetClientID("go-simple")
	opts.SetDefaultPublishHandler(f)

	//create and start a client using the above ClientOptions
	c := MQTT.NewClient(opts)
	if token := c.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	//subscribe to the topic /go-mqtt/sample and request messages to be delivered
	//at a maximum qos of zero, wait for the receipt to confirm the subscription
	if token := c.Subscribe("go-mqtt/sample", 0, nil); token.Wait() && token.Error() != nil {
		fmt.Println(token.Error())
		os.Exit(1)
	}

	//Publish 5 messages to /go-mqtt/sample at qos 1 and wait for the receipt
	//from the server after sending each message
	for i := 0; i < 5; i++ {
		text := fmt.Sprintf("this is msg #%d!", i)
		token := c.Publish("go-mqtt/sample", 0, false, text)
		token.Wait()
	}

	time.Sleep(3 * time.Second)

	//unsubscribe from /go-mqtt/sample
	if token := c.Unsubscribe("go-mqtt/sample"); token.Wait() && token.Error() != nil {
		fmt.Println(token.Error())
		os.Exit(1)
	}

	c.Disconnect(250)
}
*/
