# promdiscovery

This is a daemon that scrape Prometheus targets from docker-compose into `<file_sd_config>`

## How it works

It daemon subscribe on dockerd events. On received container action `start` or `die`
that scrape labels of all available containers. Prometheus targets extracted by key from config.
Scrape target created like `container_name:value_from_label`.
`value_from_label` - must containing `port/uri` for scrapping metrics. 

Example configuration:
```docker-compose
version: '3'

services:
  prometheus:
    image: prom/prometheus:v2.15.2
    restart: unless-stopped
    expose:
      - 9090
    command:
      - '--config.file=/prometheus.yml'
    volumes:
      - '$PWD/dockerdata/prometheus.yml:/prometheus.yml'
      - '$PWD/dockerdata/discovered.json:/discovered.json'

  promdiscovery:
    image: sgoroshko/prometheus-discovery
    restart: unless-stopped
    command:
      - '--output=/discovered.json'
      - '--key=metrics'
    volumes:
      - '$PWD/dockerdata/discovered.json:/discovered.json'
```

Deploy new container:
```bash
docker run -d --name testName --label metrics=9090/metrics image:tag
```

discovered.json will be:
```json
[
  {
    "targets": [ "testName:9090/metrics" ]
  }
]
```

## Configuration options

```
NAME:
   promdiscovery - scrape prometheus targets for docker-compose

USAGE:
   promdiscovery [global options] command [command options] [arguments...]

VERSION:
   0.0.1

COMMANDS:
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --debug             debug mode (default: false)
   --dockerHost value  docker host (default: "unix:///var/run/docker.sock")
   --key value         scrape key (default: "metrics")
   --output value      output filename (default: "discovered.json")
   --help, -h          show help (default: false)
   --version, -v       print the version (default: false)
```

## Develop

```
$ gmake

Usage: make <TARGETS> ... <OPTIONS>

  help    print this message
  clean   remove output binary
  deps    download dependencies
  fmt     running "go fmt" on sources packages
  vet     running "go vet" on sources packages
  tests   running "go test" on sources packages
  build   compile packages and dependencies
  image   build docker image

By default print this message.

```