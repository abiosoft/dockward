package main

import (
	"fmt"
	"strings"

	"github.com/abiosoft/dockward/balancer"
	"github.com/abiosoft/dockward/util"
)

func forwardToHost(args cliConf) error {
	endpoints := make(balancer.Endpoints, len(args.Endpoints))
	for i, endpoint := range args.Endpoints {
		endpoints[i] = balancer.ParseEndpoint(endpoint)
	}

	lb := balancer.New(args.HostPort, endpoints)

	go lb.ListenForEndpoints(balancer.EndpointPort)

	fmt.Println("Forwarding", args.HostPort, "to", strings.Join(args.Endpoints, ", "))
	return lb.Start(nil)
}

func forwardToDocker(args cliConf) {
	key, val := containerFilter(args)
	if key == "" || val == "" {
		exit(fmt.Errorf("Missing container parameters."))
	}

	endpoints, err := endpointsFromFilter(args.HostPort, key, val)
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
