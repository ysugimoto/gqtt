.PHONY: test lint server client

test: lint
	DEBUG= go test ./...

lint:
	go list ./... | xargs golint
	go vet ./...

server:
	go run cmd/server/main.go

client:
	go run cmd/client/main.go
