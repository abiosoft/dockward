package main

import (
	"github.com/abiosoft/dockward/balancer"
	"github.com/docker/engine-api/types"
	"github.com/docker/engine-api/types/filters"
)

type containerFilterType int8

const (
	idFilter containerFilterType = iota + 1
	nameFilter
	labelFilter
)

func containerFilter(args cliConf) (key, value string) {
	switch args.containerFilter {
	case idFilter:
		return "id", args.ContainerId
	case nameFilter:
		return "name", args.ContainerName
	case labelFilter:
		return "label", args.ContainerLabel
	}
	return "", ""
}

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
		ip, err := ipFromContainer(c.ID)
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
