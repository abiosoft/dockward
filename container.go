package main

import (
	"fmt"

	"github.com/abiosoft/dockward/balancer"
	"github.com/docker/engine-api/types/container"
	"github.com/docker/engine-api/types/strslice"
	"github.com/docker/go-connections/nat"
)

func ipFromContainer(name string) (string, error) {
	info, err := client.ContainerInspect(name)
	if err != nil {
		return "", err
	}
	if n, ok := info.NetworkSettings.Networks[dockwardNetwork.Name]; ok {
		return n.IPAddress, nil
	}
	return "", errNetworkNotFound
}

func connectContainer(name string) error {
	return dockwardNetwork.ConnectContainer(name)
}

func forwardToBalancer(hostPort int, dests ...string) error {
	endpoints := make(balancer.Endpoints, len(dests))
	for i, dest := range dests {
		endpoints[i] = balancer.ParseEndpoint(dest)
	}

	lb := balancer.New(hostPort, endpoints)

	go lb.ListenForEndpoints(balancer.EndpointPort)

	return lb.Start(nil)
}

func createBalancerContainer(hostPort int, monitorPort int, dests ...string) error {
	resp, err := client.ContainerCreate(
		&container.Config{
			Image: AppName,
			Cmd:   append(strslice.StrSlice{fmt.Sprint(hostPort), "-host"}, strslice.StrSlice(dests)...),
		},
		&container.HostConfig{
			PortBindings: nat.PortMap{
				nat.Port(hostPort): []nat.PortBinding{
					nat.PortBinding{
						HostIP: "0.0.0.0", HostPort: fmt.Sprint(hostPort),
					},
				},
				// endpoints
				nat.Port(fmt.Sprint(monitorPort)): []nat.PortBinding{
					nat.PortBinding{
						HostIP: "0.0.0.0", HostPort: fmt.Sprint(balancer.EndpointPort),
					},
				},
			},
		}, nil, "")

	exitIfErr(err)
	dockwardContainerId = resp.ID

	err = dockwardNetwork.ConnectContainer(dockwardContainerId)
	exitIfErr(err)
	addCleanUpFunc(func() {
		client.ContainerKill(dockwardContainerId, "")
	})

	return err

}
