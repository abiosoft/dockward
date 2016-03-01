package cli

import (
	"github.com/abiosoft/dockward/network"
	docker "github.com/docker/engine-api/client"
)

var (
	client              *docker.Client
	dockwardNetwork     *network.Network
)

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