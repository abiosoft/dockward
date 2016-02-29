package main

import (
	"github.com/abiosoft/dockward/balancer"
	"github.com/docker/engine-api/types"
	"github.com/docker/engine-api/types/filters"
)

type filterType string

const (
	idFilter    filterType = "id"
	nameFilter  filterType = "name"
	labelFilter filterType = "label"
)

func endpointsFromFilter(containerPort int, key, value string) (balancer.Endpoints, error) {
	filter := filters.NewArgs()
	filter.Add(key, value)
	containers, err := client.ContainerList(types.ContainerListOptions{Filter: filter})
	if err != nil {
		return nil, err
	}
	endpoints := make(balancer.Endpoints, len(containers))
	for i, c := range containers {
		if err := connectContainer(c.ID); err != nil {
			return nil, err
		}
		ip, err := containerIp(c.ID)
		if err != nil {
			return nil, err
		}
		endpoints[i] = balancer.Endpoint{
			Id:   c.ID,
			Ip:   ip,
			Port: containerPort,
		}
	}
	return endpoints, nil
}
