package balancer

import (
	"math/rand"
	"time"
)

// Policy is a selection policy.
type Policy interface {
	Select(d Endpoints) Endpoint
}

// Random is a random policy.
type Random struct{}

// Select satisfies Policy.
func (r Random) Select(d Endpoints) Endpoint {
	if len(d) == 0 {
		return Endpoint{}
	}
	return d[rand.Int()%len(d)]
}

func init() {
	rand.Seed(time.Now().UnixNano())
}
