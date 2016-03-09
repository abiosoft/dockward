package cli

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/abiosoft/dockward/network"
	docker "github.com/docker/engine-api/client"
	"github.com/docker/engine-api/types"
	"golang.org/x/net/context"
)

var (
	client          *docker.Client
	dockwardNetwork *network.Network
)

// setupDocker connects to the docker daemon and creates a dockward network.
func setupDocker() error {
	var err error
	if client, err = docker.NewEnvClient(); err != nil {
		return err
	}
	if err := pullImage(); err != nil {
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

func pullImage() error {
	// check if image is already pulled
	l, err := client.ImageList(context.Background(), types.ImageListOptions{})
	for i := range l {
		for _, tag := range l[i].RepoTags {
			if tag == imageString() {
				return nil
			}
		}
	}

	// otherwise pull image
	fmt.Println("Required docker image not found. Attempting to pull.")
	options := types.ImagePullOptions{ImageID: imageName, Tag: imageTag}
	resp, err := client.ImagePull(context.Background(), options, nil)
	if err != nil {
		return err
	}

	type progress struct {
		Status   string `json:"status"`
		Id       string `json:"id"`
		Error    string `json:"error"`
		Progress string `json:"progress"`
	}

	decoder := json.NewDecoder(resp)
	for {
		var p progress
		err := decoder.Decode(&p)
		if err != nil && err != io.EOF {
			resp.Close()
			return err
		} else if err == io.EOF {
			fmt.Println()
			break
		}
		if p.Error != "" {
			return fmt.Errorf(p.Error)
		}
		if p.Progress == "" {
			fmt.Println(p.Status, p.Id)
		}
	}

	return nil
}
