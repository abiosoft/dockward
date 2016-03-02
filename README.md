```sh
dockward 80 name=nginx
```

![Demonstration](https://github.com/abiosoft/dockward/blob/master/dockward.gif)

Typical dockward use cases include:
* Port forwarding to containers without published ports.
* Port forwarding to containers based on a filter.
* Testing stateless app scaled via docker-compose.

## Installation
### Binary
```sh
curl -LO https://github.com/abiosoft/dockward/releases/download/0.0.2/dockward_linux_amd64.tar.gz \
&& tar xvfz dockward_linux_amd64.tar.gz \
&& sudo mv dockward_linux_amd64 /usr/local/bin/dockward \
&& sudo chmod +x /usr/local/bin/dockward
```
### Source
```
go get github.com/abiosoft/dockward
```

## Requirements
Docker 1.10 and running docker daemon.

## Features
* Port forwarding to docker containers identifiable by id, name or label.
* Port forwarding to other host ports.
* Port forwarding to other endpoints.
* Auto load balancing with random or round robin policies.
* Auto endpoints management on load balancer.

## Usage
Note: If dockward is not running on Linux or docker host, you will access it via docker host ip (e.g. dockermachine ip). Except `--host` mode.

Forward port `8080` to port `80` in container with id `966a469e9716`.
```sh
dockward 8080 80 id=966a469e9716
```
Forward port `8080` to port `80` in containers with label `type=nginx`.
```sh
dockward 8080 80 label=type=nginx
```
Forward port `8080` to a local port `3000`.
```sh
dockward --host 8080 3000
```
Forward port `8080` to endpoints `127.0.0.1:3000` and `127.0.0.1:3001`.
```sh
dockward --host 8080 127.0.0.1:3000 127.0.0.1:3001
```
For more.
```sh
dockward --help
```

## Note
* dockward is still in experimental state, don't rely on the stability.
* dockward is intended for simple local development use cases. It may work for you outside of that.
* docker networks created are default settings i.e. bridge/overlay as the case may be. Nothing special.

## Why the name ?
Naming is hard, you know.

**Dock**erFor**ward**, port **forwarding** tool for **docker** containers.


