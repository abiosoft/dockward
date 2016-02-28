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

func disconnectContainer(name string) error {
	return dockwardNetwork.DisconnectContainer(name)
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
	hPort := nat.Port(fmt.Sprintf("%d/tcp", hostPort))
	mPort := nat.Port(fmt.Sprintf("%d/tcp", balancer.EndpointPort))
	resp, err := client.ContainerCreate(
		&container.Config{
			Image: AppName,
			Cmd:   append(strslice.StrSlice{fmt.Sprint(hostPort), "--host"}, strslice.StrSlice(dests)...),
			ExposedPorts: map[nat.Port]struct{}{
				hPort: struct{}{},
				mPort: struct{}{},
			},
		},
		&container.HostConfig{
			PortBindings: nat.PortMap{
				hPort: []nat.PortBinding{
					nat.PortBinding{
						HostIP: "0.0.0.0", HostPort: fmt.Sprint(hostPort),
					},
				},
				// endpoints update port
				mPort: []nat.PortBinding{
					nat.PortBinding{
						HostIP: "0.0.0.0", HostPort: fmt.Sprint(monitorPort),
					},
				},
			},
		}, nil, "")

	exitIfErr(err)
	dockwardContainerId = resp.ID

	err = dockwardNetwork.ConnectContainer(dockwardContainerId)
	exitIfErr(err)

	err = client.ContainerStart(dockwardContainerId)
	exitIfErr(err)

	addCleanUpFunc(func() {
		client.ContainerKill(dockwardContainerId, "")
	})

	return err

}
