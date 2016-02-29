package balancer

import (
	"math/rand"
	"sync/atomic"
	"time"
)

// Policy is a selection policy.
type Policy interface {
	Select(e Endpoints) Endpoint
}

// Random is a random policy.
type Random struct{}

func init() {
	rand.Seed(time.Now().UnixNano())
}

// Select satisfies Policy.
func (r Random) Select(e Endpoints) Endpoint {
	if len(e) == 0 {
		return Endpoint{}
	}
	return e[rand.Int()%len(e)]
}

type RoundRobin struct {
	count uint32
}

func (r *RoundRobin) Select(e Endpoints) Endpoint {
	if len(e) == 0 {
		return Endpoint{}
	}
	i := atomic.AddUint32(&r.count, 1) % uint32(len(e))
	return e[i]
}
