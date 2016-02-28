package network

import (
	"github.com/abiosoft/dockward/util"
	docker "github.com/docker/engine-api/client"
	"github.com/docker/engine-api/types"
)

const namePrefix = "dockward_"
const nameSuffixLen = 10

type Network struct {
	Name   string
	Id     string
	client *docker.Client
}


// Create creates a new network with a random name.
func Create(client *docker.Client) (*Network, error) {
	name := namePrefix + util.RandChars(nameSuffixLen)
	return CreateWithName(client, name)
}

// Create creates a new network using name.
func CreateWithName(client *docker.Client, name string) (*Network, error) {
	n, err := client.NetworkCreate(types.NetworkCreate{Name: name})
	if err != nil {
		return nil, err
	}
	return &Network{
		Name:   name,
		Id:     n.ID,
		client: client,
	}, nil
}

func (n *Network) ConnectContainer(id string) error {
	return n.client.NetworkConnect(n.Id, id, nil)
}

func (n *Network) Stop() error {
	info, err := n.client.NetworkInspect(n.Id)
	if err != nil {
		return err
	}
	// disconnect all containers from it
	for id, _ := range info.Containers {
		if err := n.client.NetworkDisconnect(n.Id, id, true); err != nil {
			return err
		}
	}
	// Remove network
	return n.client.NetworkRemove(n.Id)
}
