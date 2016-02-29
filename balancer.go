package main

import (
	"fmt"
	"github.com/abiosoft/dockward/balancer"
	"github.com/abiosoft/dockward/util"
	"strings"
)

func forwardToHost(args cliArgs) error {
	fmt.Println("Forwarding", args.HostPort, "to", strings.Join(args.Endpoints, ", "))
	endpoints := make(balancer.Endpoints, len(args.Endpoints))
	for i, endpoint := range args.Endpoints {
		endpoints[i] = balancer.ParseEndpoint(endpoint)
	}

	lb := balancer.New(args.HostPort, endpoints)

	go lb.ListenForEndpoints(balancer.EndpointPort)

	return lb.Start(nil)
}

func forwardToDocker(args cliArgs) {
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

	err = launchBalancerContainer(args.HostPort, endpointPort, dests...)
	exitIfErr(err)

	if args.ContainerLabel == "" {
		fmt.Println("Forwarding", args.HostPort, "to container", args.ContainerName+args.ContainerId)
		return

	}

	go monitor(endpointPort, args.HostPort, args.ContainerLabel)
	fmt.Println("Forwarding", args.HostPort, "to containers with label="+args.ContainerLabel)
}
