package main

import (
	"errors"
	"fmt"
	"os"

	"strings"

	"github.com/abiosoft/dockward/network"
	docker "github.com/docker/engine-api/client"
)

var (
	client              *docker.Client
	dockwardNetwork     *network.Network
	dockwardContainerId string

	errNetworkNotFound = errors.New("Error: Network not found. Consider restarting dockward.")
)

func main() {
	args := parseCli()

	if args.Host {
		fmt.Println("Forwarding", args.HostPort, "to", strings.Join(args.Endpoints, ", "))
		forwardToHost(args.HostPort, args.Endpoints...)
		return
	}

	exitIfErr(setupDocker())

	forwardToDocker(args)

	<-trapInterrupts(nil)
}

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

func exitIfErr(err error) {
	if err != nil {
		exit(err)
	}
}

func exit(err error) {
	if err == nil {
		os.Exit(0)
	}
	fmt.Println(err)
	os.Exit(1)
}
