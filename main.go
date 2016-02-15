package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/abiosoft/dockward/tcpforward"
)

var dockerBin string
var dockerMode bool

const NAME = "dockward"

func init() {
	if len(os.Args) > 1 && os.Args[1] == "docker" {
		dockerMode = true
		return
	}
	d, err := exec.LookPath("docker")
	if err != nil {
		exitWithErr(errors.New("Docker binary not found in PATH"))
	}
	dockerBin = d
	if err := run(args("version").Mute()); err != nil {
		exitWithErr(err)
	}
}

func main() {
	testContainer()
}

func testContainer() {
	if dockerMode {
		forwardToContainer()
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

	ip, err := ipFromContainer(dest)
	if ip == "" {
		if err = connectContainer(dest); err == nil {
			ip, err = ipFromContainer(dest)
		}
	}
	if err != nil {
		exitWithErr(err)
	} else if ip == "" {
		exitWithErr(errors.New("Error communicating with container"))
	}
	dest = ip + ":" + destPort
	opt := &options{
		args: []string{"run", "-it", "-p", port + ":" + port, "--net", NAME, NAME, "docker", port, dest},
	}

	fmt.Println("Forwarding", port, "to", dest)
	err = run(opt)
	exitWithErr(err)

}

func ipFromContainer(name string) (string, error) {

	buf := bytes.NewBuffer(nil)
	opt := &options{
		args:   []string{"inspect", name},
		stdout: buf,
	}
	err := run(opt)
	if err != nil {
		return "", err
	}
	var container []struct {
		NetworkSettings struct {
			Networks struct {
				Dockward struct {
					IPAddress string
				} `json:"dockward"`
			}
		}
	}

	err = json.NewDecoder(buf).Decode(&container)
	if err != nil {
		return "", err
	}
	return container[0].NetworkSettings.Networks.Dockward.IPAddress, nil
}

func connectContainer(name string) error {
	//network ls -q -f name=dockward
	b := bytes.NewBuffer(nil)
	opts := &options{
		stdout: b,
		args:   []string{"network", "ls", "-q", "-f", "name=" + NAME},
	}
	if err := run(opts); err != nil {
		return err
	} else if b.Len() == 0 {
		if err := run(args("network", "create", NAME).Mute()); err != nil {
			return err
		}
	}
	return run(args("network", "connect", NAME, name))
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
