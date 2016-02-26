package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/abiosoft/dockward/network"
	"github.com/abiosoft/dockward/tcpforward"
	docker "github.com/docker/engine-api/client"
	"github.com/docker/engine-api/types"
	"github.com/docker/engine-api/types/container"
	"github.com/docker/engine-api/types/filters"
	"github.com/docker/engine-api/types/strslice"
	"github.com/docker/go-connections/nat"
	"golang.org/x/net/context"
)

var dockerBin string
var dockerMode bool
var monitorMode bool
var client *docker.Client
var dockerNetwork *network.Network
var dockwardContainerId string

var errNetworkNotFound = errors.New("Error: Network not found. Consider restarting dockward.")

const NAME = "dockward"

func init() {
	if len(os.Args) > 1 && os.Args[1] == "docker" {
		dockerMode = true
		return
	} else if len(os.Args) > 1 && os.Args[1] == "monitor" {
		monitorMode = true
	}
	d, err := exec.LookPath("docker")
	if err != nil {
		exitWithErr(errors.New("Docker binary not found in PATH"))
	}
	dockerBin = d
	if err := run(args("version").Mute()); err != nil {
		exitWithErr(err)
	}
	if client, err = docker.NewEnvClient(); err != nil {
		exitWithErr(err)
	}
	if dockerNetwork, err = network.Create(client); err != nil {
		exitWithErr(err)
	}
}

func main() {
	testContainer()
}

const (
	Die       = "die"
	Start     = "start"
	Container = "container"
)

type Event struct {
	Status string `json:"status"`
	Type   string
	Id     string `json:"id"`
	Actor  struct {
		Attributes map[string]string
	}
}

func monitor() {
	filter := filters.NewArgs()
	filter.Add("label", "name=dock")
	containers, err := client.ContainerList(types.ContainerListOptions{Filter: filter})
	exitWithErr(err)
	fmt.Println(containers)

	exitWithErr(err)
	resp, err := client.Events(context.Background(), types.EventsOptions{})
	exitWithErr(err)
	decoder := json.NewDecoder(resp)
	var e Event
	for {
		err := decoder.Decode(&e)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			continue
		}
		fmt.Println(e)
	}
}

func testContainer() {
	if dockerMode {
		forwardToContainer()
		return
	} else if monitorMode {
		monitor()
		return
	}

	if len(os.Args) < 3 {
		exitWithErr(errors.New("Port and container id/name missing."))
	}

	port := os.Args[1]
	dest := os.Args[2]
	destPort := port
	if len(os.Args) > 3 {
		destPort = os.Args[3]
	}

	var ip string
	var err error
	if err = connectContainer(dest); err == nil {
		ip, err = ipFromContainer(dest)
	}
	if err != nil {
		exitWithErr(err)
	}
	dest = ip + ":" + destPort

	resp, err := client.ContainerCreate(
		&container.Config{
			Image: NAME,
			Cmd:   strslice.StrSlice{"docker", port, dest},
		},
		&container.HostConfig{
			PortBindings: nat.PortMap{
				nat.Port(port): []nat.PortBinding{
					nat.PortBinding{
						HostIP: "0.0.0.0", HostPort: port,
					},
				},
			},
		},
		nil,
		"",
	)

	exitWithErr(err)
	dockwardContainerId = resp.ID

	err = dockerNetwork.ConnectContainer(dockwardContainerId)
	exitWithErr(err)

	fmt.Println("Forwarding", port, "to", dest)

}

func ipFromContainer(name string) (string, error) {
	info, err := client.ContainerInspect(name)
	if err != nil {
		return "", err
	}
	if n, ok := info.NetworkSettings.Networks[dockerNetwork.Name]; ok {
		return n.IPAddress, nil
	}
	return "", errNetworkNotFound
}

func connectContainer(name string) error {
	return dockerNetwork.ConnectContainer(name)
}

func forwardToContainer() {
	if len(os.Args) < 4 {
		exitWithErr(errors.New("Port and remote address missing."))
	}
	port := os.Args[2]
	dest := os.Args[3]
	err := tcpforward.Forward(port, dest)
	exitWithErr(err)
}

func exitWithErr(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func run(options *options) error {
	cmd := exec.Command(dockerBin, options.args...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	if options.stdin != nil {
		cmd.Stdin = options.stdin
	}
	if options.stdout != nil {
		cmd.Stdout = options.stdout
	}
	if options.stderr != nil {
		cmd.Stderr = options.stderr
	}
	return cmd.Run()
}

type options struct {
	args   []string
	stdin  io.Reader
	stdout io.Writer
	stderr io.Writer
}

func (o *options) Mute() *options {
	o.stdout = ioutil.Discard
	return o
}

func args(a ...string) *options {
	return &options{args: a}
}
