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

type cliArgs struct {
	HostPort       int
	ContainerName  string
	ContainerId    string
	ContainerLabel string
	Host           bool
	Endpoints      []string
}

func usageErr(err error) {
	exit(fmt.Errorf("%v\n\n%v", err, Usage))
}

func parseCli() cliArgs {
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

	args := cliArgs{HostPort: hostPort}

	fs := flag.FlagSet{}
	fs.SetOutput(ioutil.Discard)

	fs.BoolVar(&args.Host, "host", args.Host, "")
	fs.StringVar(&args.ContainerId, "id", args.ContainerId, "")
	fs.StringVar(&args.ContainerName, "name", args.ContainerName, "")
	fs.StringVar(&args.ContainerLabel, "label", args.ContainerLabel, "")

	err = fs.Parse(os.Args[2:])
	if err != nil {
		exit(err)
	}

	// if not host mode, require a container param.
	if !args.Host {
		if args.ContainerId == "" && args.ContainerLabel == "" && args.ContainerName == "" {
			exit(fmt.Errorf("One of container id, name or label is required."))
		}
	}

	if fs.NArg() > 0 {
		args.Endpoints = fs.Args()
	}
	return args
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
