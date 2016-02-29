package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

const (
	AppName = "dockward"
	Version = "0.0.1"
	Usage = `Usage: dockward [options...] <port> [<container port> <filter>] [endpoints...]
try 'dockward --help' for more.
`
	FullUsage   = `Usage: dockward [options...] <port> [<container port> <filter>] [endpoints...]

options:
  --host=false         Host mode, forward to host endpoints instead of container.
  --docker-host=""     Ip address of docker host if this machine is not docker host.
                       For monitoring and adding/removing filtered containers.
  -h, --help=false     Show this help.
  -v, --version=false  Show version.

port:
  port to listen on. e.g. 8080

container port:
  port to forward to inside in the container. Defaults to <port>.

filter:
  Containers' filter. Supports id, name, label.
  e.g. id=749bdeaf6920, name=amazing_leavitt, label=com.myorg.key=value.

endpoints:
  port, ip/host or ip/host:port to forward to. Requires --host.

`
)

type cliConf struct {
	HostPort      int
	ContainerPort int
	Filter        string
	FilterValue   string
	Host          bool
	Endpoints     []string
	DockerHost    string

	Help    bool
	Version bool
}

func usageErr(err error) {
	exit(fmt.Errorf("dockward: %v\n%v", err, Usage))
}

func parseCli() cliConf {
	conf := cliConf{}

	fs := flag.FlagSet{}
	fs.SetOutput(ioutil.Discard)

	fs.BoolVar(&conf.Host, "host", conf.Host, "")
	fs.StringVar(&conf.DockerHost, "docker-host", conf.DockerHost, "")
	fs.BoolVar(&conf.Help, "h", conf.Help, "")
	fs.BoolVar(&conf.Help, "help", conf.Help, "")
	fs.BoolVar(&conf.Version, "v", conf.Help, "")
	fs.BoolVar(&conf.Version, "version", conf.Help, "")

	err := fs.Parse(os.Args[1:])
	exitIfErr(err)

	if conf.Help {
		fmt.Println(FullUsage)
		exit(nil)
	}
	if conf.Version {
		fmt.Println("dockward version", Version)
		exit(nil)
	}

	switch fs.NArg() {
	case 0:
		usageErr(fmt.Errorf("port missing."))
	case 1:
		usageErr(fmt.Errorf("filter or endpoint missing."))
	}

	args := fs.Args()

	conf.HostPort, err = strconv.Atoi(args[0])
	if err != nil {
		usageErr(err)
	}

	// if not host mode, require one container param.
	if !conf.Host {
		var filterArg string
		if len(args) > 2 {
			conf.ContainerPort, err = strconv.Atoi(args[1])
			if err != nil {
				usageErr(err)
			}
			filterArg = args[2]
		} else {
			conf.ContainerPort = conf.HostPort
			filterArg = args[1]
		}

		str := strings.SplitN(filterArg, "=", 2)
		if len(str) != 2 {
			usageErr(fmt.Errorf("Invalid filter."))
		}

		conf.Filter, conf.FilterValue = str[0], str[1]

		switch conf.Filter {
		case "id", "name", "label":
		default:
			usageErr(fmt.Errorf("Invalid filter %s", conf.Filter))
		}

	} else {
		// if host mode, load endpoints.
		conf.Endpoints = args[1:]
	}

	return conf
}

func exitIfErr(err error) {
	if err != nil {
		exit(err)
	}
}

func exit(err error) {
	if err == nil {
		os.Exit(0)
	}
	fmt.Println(err)
	os.Exit(1)
}
