.PHONY: dev generate build run preprocess tidy

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
