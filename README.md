# gqtt

*MQTT5* broker/client implementation by golang.

See OASIS's spec: http://docs.oasis-open.org/mqtt/mqtt/v5.0/cs02/mqtt-v5.0-cs02.html

## Requirement

- Go1.12 or later

## Installation

This packages uses `go mod`, so you need to enable it as `export GO111MODULE=on` in your environment.

## Usage

To debug message passing, set `export DEBUG=1` before start processes.

### Broker

Simple broker with accept hook message example:

```go
package main

import (
	"context"
	"github.com/ysugimoto/gqtt"
	"github.com/ysugimoto/gqtt/message"
)

func main() {
	server := gqtt.NewBroker(":9999")
	ctx := context.Background()
	// Start server inside goroutine
	go server.ListenAndServe(ctx)

	// Hooks of messages
	for evt := range server.MessageEvent {
		switch e := evt.(type) {
    // Client subscribed
		case *message.Subscribe:
			log.Println("Received SUBSCRIBE event: ", e.GetType())
		// Client connected
		case *message.Connect:
			log.Println("Received CONNECT event", e.GetType())
		// Client published message
		case *message.Publish:
			log.Println("Received PUBLISH event", e.GetType())
		}
	}
	<-ctx.Done()
}
```

### Client

Simple connect (with authentication) -&gt; subscribe -&gt; publish example.
The auth challenge during connect phase is new of MQTT5 spec.

```go
package main

import (
	"log"
	"os"
	"time"
	"context"

	"github.com/ysugimoto/gqtt"
	"github.com/ysugimoto/gqtt/message"
)

func main() {
	client := gqtt.NewClient("mqtt://localhost:9999")
	defer client.Disconnect()

	ctx := context.Background()

	// Connect with authenticate
	auth := gqtt.WithLoginAuth("admin", "admin")
	if err := client.Connect(ctx, auth); err != nil {
		log.Fatal(err)
	}

	// Subscribe topic
	if err := client.Subscribe("gqtt/example", message.QoS2); err != nil {
		log.Fatal(err)
	}

	ticker := time.NewTicker(3 * time.Second)

	for {
		select {
		case <-client.Closed:
			log.Println("connection closed")
			return
		case <-ctx.Done():
			log.Println("context canceled")
			return
		case msg := <-client.Message:
			log.Printf("published message received: %s\n", string(msg.Body))
		case <-ticker.C:
			log.Printf("message publish")
			if err := client.Publish("gqtt/example", []byte("Hello, MQTT5!"), gqtt.WithQoS(message.QoS2)); err != nil {
				return
			}
		}
	}
}
```

See more examples in detail.

## Features

This package now implements partial features. See following checks:

- [x] QoS0 message
- [x] QoS1 message (but not persistent storage, only store on memory)
- [x] QoS2 message (but not persistent storage, only store on memory)
- [x] Retain message (but not persistent storage, only store on memory)
- [x] Will message
- [x] Wildcard topics
- [x] User Property
- [ ] Connect redirection
- [ ] Request/Response feature
- [ ] Auth challenge (now experimental. Only basic/login auth with `admin/admin`)
- [ ] Distirbuted brokers

## LICENSE

MIT

## Author

Yoshiaki Sugimoto


This package is still under development. PR is welcome :-)
