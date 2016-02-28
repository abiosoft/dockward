package balancer

import (
	"math/rand"
	"time"
)

// Policy is a selection policy.
type Policy interface {
	Select(e Endpoints) Endpoint
}

// Random is a random policy.
type Random struct{}

// Select satisfies Policy.
func (r Random) Select(e Endpoints) Endpoint {
	if len(e) == 0 {
		return Endpoint{}
	}
	return e[rand.Int()%len(e)]
}

func init() {
	rand.Seed(time.Now().UnixNano())
}
