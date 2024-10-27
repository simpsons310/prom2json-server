# Prom2Json Server

An HTTP server converts Prometheus request pull metrics from plain text into JSON format based on [prom2json](https://github.com/prometheus/prom2json) tool

## Introduction

There are some agent tools that scrape metrics for Prometheus and expose API-returned metrics in a plain format, with HTTP requests returning responses as text/plain. However, sometimes we need HTTP requests to return responses in JSON format. In my case, when integrating `node_exporter` with [Grafana Alloy](https://grafana.com/docs/alloy/latest/), this was not possible because Alloy only accepts HTTP responses in JSON format.

Fortunately, the Prometheus team provides a tool called [prom2json](https://github.com/prometheus/prom2json) to convert metrics data from plain text to JSON, but it can only be run as a command for each request.

To address this, I created this application, which acts as an HTTP server to convert Prometheus metrics data retrieved from agents from plain text to JSON. I hope this helps someone! ٩(＾◡＾)۶

## Usage

- Add configuration file

Copy example configuration file to have a schema

```bash
cp config.yaml.example config.yaml
```

- Building with Go

```bash
# with go
go build -v -o ./build/server/p2jsvr ./cmd/server/main.go
```

Binary file automatics detect a configuration file `config.yaml` in the directory of where command executed. However, you can specify configuration file path by flag `--config-path`

- Running

```bash
p2jsvr

# With configuration path
p2jsvr --config-file [path_to_config_file]
```

- Or you can use CMake

```bash
# build
make build

# build & run binary
make run-bin
```

Binary file built is in directory `./build/server/p2jsvr`

- Running with Docker

```bash
docker pull simpsons310/prom2json-server

docker run -it --port 8080:8080 -v [path_to_config_file]:/app/config.yaml simpsons310/prom2json-server
```

## Configuration

This application use a `YAML` file to configure how application will work. The table below show attributes of the configuration file

| attribute                     | description                                                                                      | type   | default                       |
|-------------------------------|--------------------------------------------------------------------------------------------------|--------|-------------------------------|
| server                        | Server configuration                                                                             | object |                               |
| server.port                   | HTTP server port                                                                                 | int    | 8080                          |
| logging                       | Log configuration                                                                                | object |                               |
| logging.time_format           | Log time format. Following by [golang time format](https://pkg.go.dev/time#pkg-constants)        | string | 2006-01-02T15:04:05Z07:00     |
| logging.level                 | Log level. `0 - debug`, `1 - info`, `2 - warn`, `3 - error`, `4 - fatal`, `5 - panic`            | int    | 0                             |
| logging.json_format           | Print log with JSON format                                                                       | bool   | false                         |
| logging.with_caller           | Print log with caller                                                                            | bool   | false                         |
| metrics                       | Metrics configuration                                                                            | array  |                               |
| metrics[].handler_path        | URI path HTTP server will expose                                                                 | string | /metrics                      |
| metrics[].url                 | URL which data scraper agent expose                                                              | string | http://localhost:9001/metrics |
| metrics[].response_timeout    | Timeout server will wait to agent return response                                                | int    | 10                            |
| metrics[].tls_cert            | Path to TLS certificate, in case agent HTTP endpoint requires                                    | string |                               |
| metrics[].tls_key             | Path to TLS certificate key, in case agent HTTP endpoint requires                                | string |                               |
| metrics[].tls_insecure_verify | In case agent HTTP URL is an `HTTPS` but no certificate requires, set this field value to `true` | bool   | false                         |

- Example of `config.yaml`

```yaml
server:
  port: 8080

logging:
  time_format: "2006-01-02T15:04:05Z07:00"
  level: 0
  json_format: false
  with_caller: true

metrics:
  - handler_path: /metrics
    url: http://localhost:9001/metrics
    response_timeout: 10
    tls_cert: ''
    tls_key: ''
    tls_insecure_verify: false
```