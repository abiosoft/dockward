package network

import (
	"github.com/abiosoft/dockward/util"
	docker "github.com/docker/engine-api/client"
	"github.com/docker/engine-api/types"
	"golang.org/x/net/context"
)

const namePrefix = "dockward_"
const nameSuffixLen = 10

// Network is a dockward network.
type Network struct {
	Name   string
	ID     string
	client *docker.Client
}

// Create creates a new network with a random name.
func Create(client *docker.Client) (*Network, error) {
	name := namePrefix + util.RandomChars(nameSuffixLen)
	return CreateWithName(client, name)
}

// CreateWithName creates a new network using name.
func CreateWithName(client *docker.Client, name string) (*Network, error) {
	n, err := client.NetworkCreate(context.Background(), name, types.NetworkCreate{Internal: true})
	if err != nil {
		return nil, err
	}
	return &Network{
		Name:   name,
		ID:     n.ID,
		client: client,
	}, nil
}

// ConnectContainer connects docker container with id to the network.
func (n *Network) ConnectContainer(id string) error {
	return n.client.NetworkConnect(context.Background(), n.ID, id, nil)
}

// DisconnectContainer disconnects docker container with id from the network.
func (n *Network) DisconnectContainer(id string) error {
	return n.client.NetworkDisconnect(context.Background(), n.ID, id, false)
}

// Stop disconnects all connected docker containers from the network and
// removes the network.
func (n *Network) Stop() error {
	info, err := n.client.NetworkInspect(context.Background(), n.ID)
	if err != nil {
		return err
	}
	// disconnect all containers from it
	for id := range info.Containers {
		if err := n.client.NetworkDisconnect(context.Background(), n.ID, id, true); err != nil {
			return err
		}
	}
	// Remove network
	return n.client.NetworkRemove(context.Background(), n.ID)
}
