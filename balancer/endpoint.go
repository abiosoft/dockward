package balancer

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/abiosoft/dockward/util"
)

type Endpoint struct {
	Id   string
	Ip   string
	Port int
}

func ParseEndpoint(addr string) Endpoint {
	// assume addr as host, port as 80
	ip, port, id := addr, 80, util.RandomChars(10)

	// if its valid port, assume as port, host as 127.0.0.1
	if p, err := strconv.Atoi(addr); err == nil {
		ip = "127.0.0.1"
		port = p
	}

	// attempt parse
	str := strings.Split(addr, ":")

	// valid host/port
	if len(str) > 1 {
		ip = str[0]
		port, _ = strconv.Atoi(str[1])
	}
	// valid id
	if len(str) > 2 {
		id = str[2]
	}

	return Endpoint{
		Id:   id,
		Ip:   ip,
		Port: port,
	}
}

func (ep Endpoint) Addr() string {
	return ep.Ip + ":" + fmt.Sprint(ep.Port)
}

func (ep Endpoint) String() string {
	return ep.Addr() + ":" + ep.Id
}

type Endpoints []Endpoint

func (e Endpoints) Len() int {
	return len(e)
}

func (e Endpoints) Addrs() []string {
	addrs := make([]string, e.Len())
	for i := range e {
		addrs[i] = e[i].Addr()
	}
	return addrs
}

func (e *Endpoints) Add(ep Endpoint) {
	for i, endpoint := range *e {
		if endpoint.Id == ep.Id {
			// already exists, replace instead.
			(*e)[i] = ep
			return
		}
	}
	*e = append(*e, ep)
}

func (e *Endpoints) Delete(id string) {
	pos := -1
	for i, ep := range *e {
		if ep.Id == id {
			pos = i
			break
		}
	}
	if pos == -1 {
		return
	}
	if pos == len(*e)-1 {
		*e = (*e)[:pos]
		return
	}
	*e = append((*e)[:pos], (*e)[pos+1:]...)
}
