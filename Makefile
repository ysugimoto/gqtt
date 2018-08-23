.PHONY: test lint server client

test: lint
	DEBUG= go test ./...

lint:
	go list ./... | xargs golint
	go vet ./...

generate:
	go generate ./...

server: generate
	go run cmd/server/main.go

client: generate
	go run cmd/client/main.go

will-client: generate
	go run cmd/client/will.go
