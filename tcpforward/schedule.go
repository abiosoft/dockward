package tcpforward

import "math/rand"

// Policy is a selection policy.
type Policy interface {
	Select(d Endpoints) Endpoint
}

// Random is a random policy.
type Random struct{}

// Select satisfies Policy.
func (r Random) Select(d Endpoints) Endpoint {
	return d[rand.Intn(len(d))]
}
