.PHONY: run build run-bin

run:
	go run ./cmd/server/main.go

build:
	go build -v -o ./build/server/p2jsvr ./cmd/server/main.go

run-bin:
	make build
	./build/server/p2jsvr