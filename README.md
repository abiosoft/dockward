# dockward 

Port forwarding tool for Docker containers. 

Dockward can be used for
* Port forwarding to containers without published ports.
* Port forwarding to containers based on a filter.
* Load balancing between multiple containers.

## Requirements
Docker and docker deamon running.

## Installation

### Binaries
Download binary for your platform on the [releases](https://github.com/abiosoft/dockward/releases) page.

### Source
Requires Go.
```sh
$ go get github.com/abiosoft/dockward
```

## Usage
Note: If dockward is not running on Linux or docker host, you will access it via docker host ip (e.g. dockermachine ip). Except `--host` mode.

Forward port `8080` to port `80` in container `amazing_leavitt`.
```sh
$ dockward 8080 80 name=amazing_levitt
```
Forward port `8080` to port `80` in containers with label `type=nginx`.
```sh
$ dockward 8080 80 label=type=nginx
```
Forward port `8080` to a local port `3000`.
```sh
$ dockward --host 8080 3000
```
Forward port `8080` to endpoints `127.0.0.1:3000` and `127.0.0.1:3001`.
```sh
$ dockward --host 8080 127.0.0.1:3000 127.0.0.1:3001
```
For more.
```
$ dockward --help
```

## Limitations
* Dockward is intended for simple local development use cases. It may work for you outside of that.
* Docker networks created are default settings i.e. bridge/overlay as the case may be. Nothing special.

## Why this tool ?
I wrote this to help with troubleshooting while developing.

My 2 most common use cases:

* Reaching a running container without published ports.
* Testing stateless app scaled via docker-compose.

## Why the name ?
Naming is hard, you know.

**Dock**erFor**ward**, port **forwarding** tool for **docker** containers.

## Demonstration Video
TBA

## License
Apache 2