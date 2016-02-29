package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
)

const (
	AppName = "dockward"
	Version = "0.0.1"
	Usage   = `Usage: dockward command [options] [host endpoints...]

command: one of port|help|version
  port         Port to listen on. e.g. 80
  help         Show this help.
  version      Show version.

options:
  --name=""    Container name.
  --id=""      Container id.
  --label=""   Container label e.g. com.myorg.key=value.
  --host=false Host mode, forward to host endpoints instead of container.

host endpoints:
  Endpoints to forward to. Requires --host.

`
)

type cliConf struct {
	HostPort       int
	ContainerPort  int
	ContainerName  string
	ContainerId    string
	ContainerLabel string
	Host           bool
	Endpoints      []string
	Monitor        bool
	DockerHost     string

	containerFilter containerFilterType
}

func usageErr(err error) {
	exit(fmt.Errorf("%v\n\n%v", err, Usage))
}

func parseCli() cliConf {
	if len(os.Args) == 1 {
		usageErr(fmt.Errorf("Command missing"))
	}

	switch os.Args[1] {
	case "help":
		fmt.Println(Usage)
		exit(nil)
	case "version":
		fmt.Println("dockward version", Version)
		exit(nil)
	}
	hostPort, err := strconv.Atoi(os.Args[1])
	if err != nil {
		usageErr(err)
	}

	conf := cliConf{HostPort: hostPort}

	fs := flag.FlagSet{}
	fs.SetOutput(ioutil.Discard)

	fs.BoolVar(&conf.Host, "host", conf.Host, "")
	fs.StringVar(&conf.ContainerId, "id", conf.ContainerId, "")
	fs.StringVar(&conf.ContainerName, "name", conf.ContainerName, "")
	fs.StringVar(&conf.ContainerLabel, "label", conf.ContainerLabel, "")

	err = fs.Parse(os.Args[2:])
	exitIfErr(err)

	// if not host mode, require one container param.
	if !conf.Host {
		if conf.ContainerId == "" && conf.ContainerLabel == "" && conf.ContainerName == "" {
			usageErr(fmt.Errorf("One of container id, name or label is required."))
		}
		filters := 0
		if conf.ContainerId != "" {
			conf.containerFilter = idFilter
			filters++
		}
		if conf.ContainerName != "" {
			conf.containerFilter = nameFilter
			filters++
		}
		if conf.ContainerLabel != "" {
			conf.containerFilter = labelFilter
			filters++
		}
		if filters > 1 {
			usageErr(fmt.Errorf("Only one of container id, name or label is required"))
		}
	} else {
		// if host mode, load endpoints.
		if fs.NArg() > 0 {
			conf.Endpoints = fs.Args()
		}
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
