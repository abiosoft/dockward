package balancer

import (
	"io"
	"log"
	"net"
	"sync"
)

func handleConn(conn net.Conn, dest string) {
	client, err := net.Dial("tcp", dest)
	if err != nil {
		log.Println(err)
		if err := conn.Close(); err != nil {
			log.Println(err)
		}
		return
	}
	var w sync.WaitGroup
	w.Add(2)
	go func() {
		defer conn.Close()
		defer client.Close()
		io.Copy(client, conn)
		w.Done()
	}()
	go func() {
		defer client.Close()
		defer conn.Close()
		io.Copy(conn, client)
		w.Done()
	}()
	w.Wait()
}
