package tcpforward

import "fmt"

type Endpoint struct {
	Id   string
	Ip   string
	Port int
}

func (ep Endpoint) Addr() string {
	return ep.Ip + ":" + fmt.Sprint(ep.Port)
}

type Endpoints []Endpoint

func (e Endpoints) Len() int {
	return len(e)
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
	part := (*e)[:pos]
	if pos < len(*e)-1 {
		part = append(part, (*e)[pos+1:]...)
	}
	*e = part
}
