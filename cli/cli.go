package cli

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

const (
	appName    = "dockward"
	appVersion = "0.0.4"
	appUsage   = `Usage: dockward [options...] <port> [<container port> <filter>] [endpoints...]
try 'dockward --help' for more.
`
	appFullUsage = `Usage: dockward [options...] <port> [<container port> <filter>] [endpoints...]

options:
  --host=false         Host mode, forward to host endpoints instead of container.
  --docker-host=""     Ip address of docker host if this machine is not docker host.
                       For monitoring and adding/removing filtered containers.
  --policy="random"    Load balancer scheduling policy. One of "random", "round_robin".
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
	Policy        string

	Help         bool
	Version      bool
	remoteClient bool
}

func parseCli() cliConf {
	conf := cliConf{Policy: "random"}

	fs := flag.FlagSet{}
	fs.SetOutput(ioutil.Discard)

	fs.BoolVar(&conf.Host, "host", conf.Host, "")
	fs.StringVar(&conf.DockerHost, "docker-host", conf.DockerHost, "")
	fs.BoolVar(&conf.Help, "h", conf.Help, "")
	fs.BoolVar(&conf.Help, "help", conf.Help, "")
	fs.BoolVar(&conf.Version, "v", conf.Version, "")
	fs.BoolVar(&conf.Version, "version", conf.Version, "")
	fs.StringVar(&conf.Policy, "policy", conf.Policy, "")
	fs.BoolVar(&conf.remoteClient, "remote-client", conf.remoteClient, "")

	err := fs.Parse(os.Args[1:])
	exitIfErr(err)

	if conf.Help {
		fmt.Println(appFullUsage)
		exit(nil)
	}
	if conf.Version {
		fmt.Println(appName, "version", appVersion)
		exit(nil)
	}

	switch fs.NArg() {
	case 0:
		usageErr(errors.New("port missing"))
	case 1:
		if !conf.Host {
			// docker mode
			usageErr(errors.New("filter missing"))
		} else if !conf.remoteClient {
			// host mode
			usageErr(errors.New("endpoint(s) missing"))
		}

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

func usageErr(err error) {
	exit(fmt.Errorf("dockward: %v\n%v", err, appUsage))
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
