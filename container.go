package main

import (
	"errors"
	"fmt"

	"github.com/abiosoft/dockward/balancer"
	"github.com/docker/engine-api/types/container"
	"github.com/docker/engine-api/types/strslice"
	"github.com/docker/go-connections/nat"
)

const ImageName = "abiosoft/dockward"

var errNetworkNotFound = errors.New("Error: Network not found. Consider restarting dockward.")

func containerIp(id string) (string, error) {
	info, err := client.ContainerInspect(id)
	if err != nil {
		return "", err
	}
	if n, ok := info.NetworkSettings.Networks[dockwardNetwork.Name]; ok {
		return n.IPAddress, nil
	}
	return "", errNetworkNotFound
}

func connectContainer(id string) error {
	return dockwardNetwork.ConnectContainer(id)
}

func disconnectContainer(id string) error {
	return dockwardNetwork.DisconnectContainer(id)
}

func launchBalancerContainer(hostPort int, monitorPort int, policy string, destinations ...string) error {
	hPort := nat.Port(fmt.Sprintf("%d/tcp", hostPort))
	mPort := nat.Port(fmt.Sprintf("%d/tcp", balancer.EndpointPort))
	command := append(
		strslice.StrSlice{
			"--host",
			"--policy",
			policy,
			fmt.Sprint(hostPort),
		},
		strslice.StrSlice(destinations)...,
	)
	containerConf := &container.Config{
		Image: ImageName,
		Cmd:   command,
		ExposedPorts: map[nat.Port]struct{}{
			hPort: struct{}{},
			mPort: struct{}{},
		},
	}
	hostConf := &container.HostConfig{
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
	}

	resp, err := client.ContainerCreate(containerConf, hostConf, nil, "")
	if err != nil {
		return err
	}

	if err := connectContainer(resp.ID); err != nil {
		return err
	}

	if err := client.ContainerStart(resp.ID); err != nil {
		return err
	}

	addCleanUpFunc(func() {
		client.ContainerKill(resp.ID, "")
	})

	return err

}
