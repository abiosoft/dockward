package cli

import (
	"github.com/abiosoft/dockward/balancer"
	"github.com/docker/engine-api/types"
	"github.com/docker/engine-api/types/filters"
	"golang.org/x/net/context"
)

type filterType string

const (
	idFilter    filterType = "id"
	nameFilter  filterType = "name"
	labelFilter filterType = "label"
)

// endpointsFromFilter searches for containers with filter key and value, then create
// endpoints from them.
func endpointsFromFilter(containerPort int, key, value string) (balancer.Endpoints, error) {
	filter := filters.NewArgs()
	filter.Add(key, value)
	containers, err := client.ContainerList(context.Background(), types.ContainerListOptions{Filter: filter})
	if err != nil {
		return nil, err
	}
	endpoints := make(balancer.Endpoints, len(containers))
	for i, c := range containers {
		if err := connectContainer(c.ID); err != nil {
			return nil, err
		}
		ip, err := containerIP(c.ID)
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
