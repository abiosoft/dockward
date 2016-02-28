package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/abiosoft/dockward/balancer"
	"github.com/abiosoft/dockward/network"
	"github.com/abiosoft/dockward/util"
	docker "github.com/docker/engine-api/client"
	"strings"
)

var (
	client              *docker.Client
	dockwardNetwork     *network.Network
	dockwardContainerId string

	errNetworkNotFound = errors.New("Error: Network not found. Consider restarting dockward.")
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

func main() {
	args := parseCli()

	if args.Host {
		fmt.Println("Forwarding", args.HostPort, "to", strings.Join(args.Endpoints, ", "))
		forwardToBalancer(args.HostPort, args.Endpoints...)
		return
	}

	exitIfErr(setupDocker())

	var endpoints balancer.Endpoints
	var err error
	if args.ContainerLabel != "" {
		endpoints, err = endpointsFromLabel(args.HostPort, args.ContainerLabel)
	} else if args.ContainerName != "" {
		endpoints, err = endpointsFromName(args.HostPort, args.ContainerName)
	} else if args.ContainerId != "" {
		endpoints, err = endpointsFromId(args.HostPort, args.ContainerId)
	} else {
		err = fmt.Errorf("Missing container parameters")
	}
	exitIfErr(err)

	dests := make([]string, len(endpoints))
	for i, e := range endpoints {
		dests[i] = e.String()
	}

	endpointPort, err := util.RandomPort()
	exitIfErr(err)

	err = createBalancerContainer(args.HostPort, endpointPort, dests...)
	exitIfErr(err)
	if args.ContainerLabel != "" {
		go monitor(endpointPort, args.HostPort, args.ContainerLabel)
		fmt.Println("Forwarding", args.HostPort, "to containers with label="+args.ContainerLabel)
	} else {
		fmt.Println("Forwarding", args.HostPort, "to container", args.ContainerName+args.ContainerId)
	}

	<-trapInterrupts(nil)
}

func setupDocker() error {
	var err error
	if client, err = docker.NewEnvClient(); err != nil {
		return err
	}
	if dockwardNetwork, err = network.Create(client); err != nil {
		return err
	}
	addCleanUpFunc(func() {
		dockwardNetwork.Stop()
	})
	return nil
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
