package util

import (
	"math/rand"
	"net"
	"time"
)

// RandomPort returns a random available port on the host.
func RandomPort() (int, error) {
	l, err := net.Listen("tcp", "")
	if err != nil {
		return 0, err
	}
	if err := l.Close(); err != nil {
		return 0, nil
	}
	return l.Addr().(*net.TCPAddr).Port, nil
}

const alphaNum = "abcdefghijklmnopqrstuvwxyz0123456789"

// RandomChars returns random characters of length n.
func RandomChars(n int) string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	s := ""
	for i := 0; i < n; i++ {
		index := r.Int() % (len(alphaNum) - 1)
		s += alphaNum[index : index+1]
	}
	return s
}
