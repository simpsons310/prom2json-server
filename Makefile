.PHONY: run build run-bin run-docker

DOCKER_PASSWORD_FILE ?= docker.password
DOCKER_IMAGE ?= simpsons310/prom2json-server

run:
	go run ./cmd/server/main.go

build:
	go build -v -o ./build/server/p2jsvr ./cmd/server/main.go

run-bin:
	make build
	./build/server/p2jsvr

run-docker:
	docker run -p 8080:8080 -it -v ./config.yaml:/app/config.yaml ${DOCKER_IMAGE}:latest

docker-bulid:
	docker build -t ${DOCKER_IMAGE}:$(shell cat VERSION) .
	docker build -t ${DOCKER_IMAGE}:latest .

docker-publish:
	docker login -u simpsons310 -p $(shell cat ${DOCKER_PASSWORD_FILE})
	docker push ${DOCKER_IMAGE}:$(shell cat VERSION)
	docker push ${DOCKER_IMAGE}:latest
