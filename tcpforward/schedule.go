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
	i := rand.Int() % (len(d) - 1)
	if i < 0 {
		i = 0
	}
	return d[i]
}
