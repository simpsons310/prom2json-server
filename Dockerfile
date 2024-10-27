FROM golang:1.21-alpine AS build

WORKDIR /app

RUN apk update && apk add --no-cache git

ARG CGO_ENABLED=0

COPY go.* ./
RUN go mod download

COPY cmd/ cmd/
COPY internal/ internal/

RUN go build -v -o /app/p2jsvr ./cmd/server/main.go

FROM alpine:3.16.2

WORKDIR /app

RUN apk add --no-cache tzdata

COPY --from=build /app/p2jsvr /app/p2jsvr

EXPOSE 8080/tcp

CMD ["./p2jsvr", "$@"]