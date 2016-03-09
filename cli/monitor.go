package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/abiosoft/dockward/balancer"
	"github.com/docker/engine-api/types"
	"golang.org/x/net/context"
)

const (
	StatusDie     = "die"
	StatusStart   = "start"
	TypeContainer = "container"
)

type Event struct {
	Status string `json:"status"`
	Type   string
	Id     string `json:"id"`
	Actor  struct {
		Attributes map[string]string
	}
}

// monitor monitors docker containers and add/remove from port forwarding
// endpoints as required.
func monitor(endpointPort int, containerPort int, label, dockerHost string) {
	resp, err := client.Events(context.Background(), types.EventsOptions{})
	exitIfErr(err)

	decoder := json.NewDecoder(resp)

eventLoop:
	for {
		var e Event
		if err := decoder.Decode(&e); err != nil {
			log.Println(os.Stderr, err)
			continue
		}
		if e.Type != TypeContainer {
			continue
		}
		if !validContainer(e.Id, label) {
			continue
		}

		msg := balancer.Message{
			Endpoint: balancer.Endpoint{
				Id:   e.Id,
				Port: containerPort,
			},
		}
		switch e.Status {
		case StatusDie:
			msg.Remove = true
			err = disconnectContainer(e.Id)
			if err != nil {
				log.Println(err)
				continue eventLoop
			}
		case StatusStart:
			err = connectContainer(e.Id)
			if err != nil {
				log.Println(err)
				continue eventLoop
			}
			ip, err := containerIp(e.Id)
			if err != nil {
				log.Println(err)
				continue
			}
			msg.Endpoint.Ip = ip
		default:
			continue eventLoop
		}

		updateContainerEndpoints(msg, dockerHost, endpointPort)
	}
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
