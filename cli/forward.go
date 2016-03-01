package cli

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

	lb := balancer.New(args.HostPort, endpoints, args.Policy)

	go lb.ListenForEndpoints(balancer.EndpointPort)

	fmt.Println("Forwarding", args.HostPort, "to", strings.Join(endpoints.Addrs(), ", "))
	return lb.Start(nil)
}

func forwardToDocker(args cliConf) {
	endpoints, err := endpointsFromFilter(args.ContainerPort, args.Filter, args.FilterValue)
	exitIfErr(err)

	// if filter is not label, it has to exist.
	if args.Filter != string(labelFilter) && endpoints.Len() == 0 {
		exit(fmt.Errorf("Container with %s=%s not found.", args.Filter, args.FilterValue))
	}

	destinations := make([]string, len(endpoints))
	for i, e := range endpoints {
		destinations[i] = e.String()
	}

	endpointPort, err := util.RandomPort()
	exitIfErr(err)

	err = launchBalancerContainer(args.HostPort, endpointPort, args.Policy, destinations...)
	exitIfErr(err)

	if args.Filter == string(labelFilter) {
		go monitor(endpointPort, args.ContainerPort, args.FilterValue, args.DockerHost)
		fmt.Println("Forwarding", args.HostPort, "to", args.ContainerPort, "in containers with label="+args.FilterValue)
	} else {
		fmt.Println("Forwarding", args.HostPort, "to", args.ContainerPort, "in container", args.FilterValue)
	}
}
