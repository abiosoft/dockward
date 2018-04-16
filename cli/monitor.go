package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/abiosoft/dockward/balancer"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/events"
)

const (
	statusDie     = "die"
	statusStart   = "start"
	typeContainer = "container"
)

type event struct {
	Status string `json:"status"`
	Type   string
	ID     string `json:"id"`
	Actor  struct {
		Attributes map[string]string
	}
}

// monitor monitors docker containers and add/remove from port forwarding
// endpoints as required.
func monitor(endpointPort, containerPort int, label, dockerHost string) {
	resp, err := client.Events(context.Background(), types.EventsOptions{})

eventLoop:
	for {
		select {
		case m := <-resp:
			handleMessage(m, endpointPort, containerPort, label, dockerHost)
		case err := <-err:
			log.Println(err)
			break eventLoop
		}
	}

}

func handleMessage(e events.Message, endpointPort, containerPort int, label, dockerHost string) error {
	if e.Type != typeContainer {
		return nil
	}
	if !validContainer(e.ID, label) {
		return nil
	}

	msg := balancer.Message{
		Endpoint: balancer.Endpoint{
			Id:   e.ID,
			Port: containerPort,
		},
	}

	switch e.Status {
	case statusDie:
		msg.Remove = true
		err := disconnectContainer(e.ID)
		if err != nil {
			return err
		}
	case statusStart:
		err := connectContainer(e.ID)
		if err != nil {
			return err
		}
		ip, err := containerIP(e.ID)
		if err != nil {
			return err
		}
		msg.Endpoint.Ip = ip
	default:
		return nil
	}

	go updateContainerEndpoints(msg, dockerHost, endpointPort)
	return nil

}

// updateContainerEndpoints updates the endpoints on the load balancer.
func updateContainerEndpoints(msg balancer.Message, dockerHost string, endpointPort int) {
	url := fmt.Sprintf("http://127.0.0.1:%d", endpointPort)
	if dockerHost != "" {
		url = fmt.Sprintf("http://%s:%d", dockerHost, endpointPort)
	}
	body := bytes.NewBuffer(nil)
	if err := json.NewEncoder(body).Encode(&msg); err != nil {
		log.Println(err)
		return
	}
	resp, err := http.Post(url, "application/json", body)
	if err != nil {
		log.Println(err)
		log.Println("Set --docker-host flag to fix this.")
		return
	}
	if resp.StatusCode != 200 {
		log.Println("Failed:", resp.Status)
	} else {
		if msg.Remove {
			log.Println("Removed", msg.Endpoint.Id, msg.Endpoint.Addr())
		} else {
			log.Println("Added", msg.Endpoint.Id, msg.Endpoint.Addr())
		}
	}
}

// validContainer validates if the container can be added/removed from endpoints.
func validContainer(name string, label string) bool {
	info, err := client.ContainerInspect(context.Background(), name)
	if err != nil {
		log.Println(err)
		return false
	}
	kv := strings.SplitN(label, "=", 2)
	if len(kv) != 2 {
		return false
	}
	v, ok := info.Config.Labels[kv[0]]
	return ok && v == kv[1]
}
