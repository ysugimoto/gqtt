.PHONY: test lint

test: lint
	DEBUG= go test ./...

lint:
	go list ./... | xargs golint
	go vet ./...
