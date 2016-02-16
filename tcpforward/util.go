package tcpforward

import "net"

func randomPort() (int, error) {
	l, err := net.Listen("tcp", "")
	if err != nil {
		return 0, err
	}
	if err := l.Close(); err != nil {
		return 0, nil
	}
	return l.Addr().(*net.TCPAddr).Port, nil
}
