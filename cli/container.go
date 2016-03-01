package cli

import (
	"errors"
	"fmt"

	"github.com/abiosoft/dockward/balancer"
	"github.com/docker/engine-api/types/container"
	"github.com/docker/engine-api/types/strslice"
	"github.com/docker/go-connections/nat"
)

var errNetworkNotFound = errors.New("Error: Network not found. Consider restarting dockward.")

const imageName = "abiosoft/dockward"
const imageTag = "latest"

func imageString() string {
	return imageName + ":" + imageTag
}

// containerIp retrieves the ip address of container with id on the dockward network.
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

// connectContainer connects container with id to the dockward network.
func connectContainer(id string) error {
	return dockwardNetwork.ConnectContainer(id)
}

// disconnectContainer disconnects container with id to the dockward network.
func disconnectContainer(id string) error {
	return dockwardNetwork.DisconnectContainer(id)
}

// launchBalancerContainer creates and launches a docker container on the dockward network
// to load balance requests to other containers.
func launchBalancerContainer(hostPort int, monitorPort int, policy string, destinations ...string) error {
	hPort := nat.Port(fmt.Sprintf("%d/tcp", hostPort))
	mPort := nat.Port(fmt.Sprintf("%d/tcp", balancer.EndpointPort))
	command := append(
		strslice.StrSlice{
			"--host",
			"--policy",
			policy,
			"--remote-client",
			fmt.Sprint(hostPort),
		},
		strslice.StrSlice(destinations)...,
	)
	containerConf := &container.Config{
		Image: imageString(),
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
