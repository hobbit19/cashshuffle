package server

import (
	"fmt"
	"net"
)

// Start brings up the TCP server.
func Start(port int, cert string, key string, poolSize int) (err error) {
	var listener net.Listener

	if tlsEnabled(cert, key) {
		listener, err = createTLSListener(port, cert, key)
		if err != nil {
			return err
		}
	} else {
		listener, err = net.Listen("tcp", fmt.Sprintf(":%d", port))
		if err != nil {
			return err
		}
	}

	defer listener.Close()

	t := &tracker{
		poolSize: poolSize,
	}
	t.init()

	signedChan := make(chan *signedConn)
	go startSignedChan(signedChan)

	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}

		go handleConnection(conn, signedChan, t)
	}
}

func handleConnection(conn net.Conn, c chan *signedConn, tracker *tracker) {
	defer conn.Close()
	defer tracker.remove(conn)

	processMessages(conn, c, tracker)

	return
}
