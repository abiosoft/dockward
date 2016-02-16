package tcpforward

import (
	"fmt"
	"io"
	"log"
	"net"
	"sync"
)

func Forward(port, dest string) error {
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return err
	}
	for {
		conn, err := listener.Accept()
		fmt.Println("here")
		if err != nil {
			log.Println(err)
			if _, ok := err.(*net.OpError); ok {
				log.Println("Connection is closed.")
				break
			}
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
