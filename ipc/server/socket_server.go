package server

import (
	"log"
	"net"
	"os"
)

// Minimal unix socket server that will run ClientHandler on every connected client.
// This handler takes a channel as a parameter to communicate with the rest of your server.
type SocketServer struct {
	SocketPath    string
	ClientHandler func(net.Conn) error
	sock          net.Listener
}

func (l *SocketServer) Start() error {
	sock, err := net.Listen("unix", l.SocketPath)
	if err != nil {
		return err
	}
	l.sock = sock
	return nil
}

func (l *SocketServer) ManageClients() {
	for {
		conn, err := l.sock.Accept() // Blocking
		if err != nil {
			log.Printf("accept err: %v", err)
			continue
		}

		go func(conn net.Conn) { // Go routine for each client
			defer conn.Close()

			err := l.ClientHandler(conn)
			if err != nil {
				log.Printf("client err: %v", err)
			}
		}(conn)
	}
}

func (l *SocketServer) Close() error {
	return os.Remove(l.SocketPath)
}
