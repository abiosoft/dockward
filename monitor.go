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
	"github.com/docker/engine-api/types/filters"
	"golang.org/x/net/context"
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

func monitor(endpointPort int, label string) {
	filter := filters.NewArgs()
	filter.Add("label", label)
	resp, err := client.Events(context.Background(), types.EventsOptions{Filters: filter})
	exitIfErr(err)

	decoder := json.NewDecoder(resp)
	var e Event

eventLoop:
	for {
		err := decoder.Decode(&e)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			continue
		}
		if e.Type != Container {
			log.Println("Not container event, ignoring...")
			continue
		}

		msg := balancer.Message{Endpoint: balancer.Endpoint{Id: e.Id}}
		switch e.Status {
		case Die:
			msg.Remove = true
		case Start:
			msg.Remove = false
		default:
			continue eventLoop
		}

		url := "http://127.0.0.1:" + fmt.Sprint(endpointPort)
		body := bytes.NewBuffer(nil)
		if err := json.NewEncoder(body).Encode(msg); err != nil {
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
