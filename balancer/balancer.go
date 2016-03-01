package balancer

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
)

// Port is the port to listen for updates to endpoints.
// TODO make dynamic
const EndpointPort = 9923

// Message is body for an endpoint update.
type Message struct {
	Endpoint Endpoint
	Remove   bool
}

// Balancer is a load balancer.
type Balancer struct {
	Port      int
	Endpoints Endpoints
	Policy    Policy
	sync.RWMutex
}

// New creates a new Balancer.
func New(port int, endpoints Endpoints, policy string) *Balancer {
	var p Policy

	switch strings.ToLower(policy) {
	case "round_robin":
		p = &RoundRobin{}
	case "random":
		p = Random{}
	default:
		log.Println("Defaulting to random policy. Unknown scheduling policy", policy+".")
		p = Random{}
	}

	return &Balancer{
		Port:      port,
		Endpoints: endpoints,
		Policy:    p,
	}
}

// Start starts b. This function blocks if start is successful.
func (b *Balancer) Start(stop chan struct{}) error {
	listener, err := net.Listen("tcp", ":"+fmt.Sprint(b.Port))
	if err != nil {
		return err
	}

	// close on signal
	go func() {
		<-stop
		listener.Close()
	}()

	// load balanced port forwarding
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println(err)
			if _, ok := err.(*net.OpError); ok {
				log.Println("Connection is closed for port:", b.Port)
				break
			}
		}

		// no endpoints
		if b.Endpoints.Len() == 0 {
			log.Println("No endpoints")
			if err := conn.Close(); err != nil {
				log.Println(err)
			}
			continue
		}

		// choose using a policy
		endpoint := b.Select(b.Endpoints)

		// handle request
		go handleConn(conn, endpoint.Addr())
	}

	return nil
}

// Select selects an endpoint using the current scheduling policy.
func (b *Balancer) Select(e Endpoints) Endpoint {
	b.RLock()
	defer b.RUnlock()

	if b.Policy == nil {
		return Random{}.Select(e)
	}
	return b.Policy.Select(e)
}

// ListenForEndpoints listens for updates endpoints.
func (b *Balancer) ListenForEndpoints(port int) {
	handler := http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			var message Message
			err := json.NewDecoder(r.Body).Decode(&message)
			if r.Method != "POST" || err != nil {
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				return
			}

			b.Lock()
			if message.Remove {
				b.Endpoints.Delete(message.Endpoint.Id)
			} else {
				b.Endpoints.Add(message.Endpoint)
			}
			b.Unlock()

			w.WriteHeader(200)
		})

	err := http.ListenAndServe(":"+fmt.Sprint(port), handler)

	// should not get here
	log.Println(err)
}
