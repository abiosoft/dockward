package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/abiosoft/dockward/balancer"
	"github.com/docker/engine-api/types"
	"golang.org/x/net/context"
	"strings"
)

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

func monitor(endpointPort int, containerPort int, label string) {
	resp, err := client.Events(context.Background(), types.EventsOptions{})
	exitIfErr(err)

	decoder := json.NewDecoder(resp)

eventLoop:
	for {
		var e Event
		err := decoder.Decode(&e)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			continue
		}
		if e.Type != Container {
			log.Println("Not container event, ignoring...")
			continue
		}
		if !validContainer(e.Id, label) {
			log.Println("Container do not have label", label, "ignoring...")
			continue
		}

		msg := balancer.Message{
			Endpoint: balancer.Endpoint{
				Id:   e.Id,
				Port: containerPort,
			},
		}
		switch e.Status {
		case Die:
			msg.Remove = true
			err = dockwardNetwork.DisconnectContainer(e.Id)
			if err != nil {
				log.Println(err)
				continue eventLoop
			}
		case Start:
			err = connectContainer(e.Id)
			if err != nil {
				log.Println(err)
				continue eventLoop
			}
			ip, err := ipFromContainer(e.Id)
			if err != nil {
				log.Println(err)
				continue
			}
			msg.Endpoint.Ip = ip
		default:
			continue eventLoop
		}

		url := "http://127.0.0.1:" + fmt.Sprint(endpointPort)
		body := bytes.NewBuffer(nil)
		fmt.Println(msg)
		if err := json.NewEncoder(body).Encode(&msg); err != nil {
			log.Println(err)
			continue
		}
		resp, err := http.Post(url, "application/json", body)
		if err != nil {
			log.Println(err)
			continue
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
}

func validContainer(name string, label string) bool {
	info, err := client.ContainerInspect(name)
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
