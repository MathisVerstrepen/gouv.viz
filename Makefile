.PHONY: dev generate build run preprocess tidy test verify

dev:
	air

generate:
	templ generate -path ./web/components

build: generate
	go build -o ./bin/gouv-viz ./cmd/web

run: generate
	go run ./cmd/web

preprocess:
	go run ./cmd/preprocess

tidy:
	go mod tidy

test:
	go test ./...

verify: generate
	gofmt -w ./cmd ./internal ./web
	go mod tidy
	go test ./...
	go vet ./...
	go build ./...
