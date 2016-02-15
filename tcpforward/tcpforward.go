package tcpforward

import (
	"io"
	"log"
	"net"
	"sync"
)

var verbose = false

func Verbose(v bool) {
	verbose = v
}

func Forward(port, dest string) error {
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return err
	}
	listener.Addr().String()
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println(err)
		}
		go handle(conn, dest)
	}
	return nil
}

func handle(conn net.Conn, dest string) {
	client, err := net.Dial("tcp", dest)
	if err != nil {
		log.Println(err)
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
