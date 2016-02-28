package balancer

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"
)

// TODO make dynamic
const EndpointPort = 9923

type Message struct {
	Endpoint Endpoint
	Remove   bool
}

type Balancer struct {
	Port      int
	Endpoints Endpoints
	Policy    Policy
	sync.RWMutex
}

func New(port int, endpoints Endpoints) *Balancer {
	return &Balancer{
		Port:      port,
		Endpoints: endpoints,
	}
}

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
			continue
		}

		// choose using a policy
		endpoint := b.Select(b.Endpoints)

		// handle request
		go handle(conn, endpoint.Addr())
	}

	return nil
}

func (b *Balancer) Select(e Endpoints) Endpoint {
	b.RLock()
	defer b.RUnlock()

	if b.Policy == nil {
		return Random{}.Select(e)
	}
	return b.Policy.Select(e)
}

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
				b.Endpoints.Delete(message.Endpoint.Ip)
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
