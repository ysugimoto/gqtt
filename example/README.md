# Examples of gqtt

This is example implementation for `gqtt`.

## How to use

### Broker

```
cd server
DEBUG=1 go run main.go
```

Server will start on `:9999`.

### Client

We provide some of client examples:

- `client/main.go` -> simple client
- `client/multi-client.go -> several client will connect to the broker
- `client/retain.go` -> client with `RETAIN` message
- `client/will.go` -> client connect with `Will` message

