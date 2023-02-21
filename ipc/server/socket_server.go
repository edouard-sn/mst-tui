package server

import (
	"net"
	"os"

	"golang.org/x/exp/slog"
)

// Minimal generic unix socket server that will run ClientHandler on every connected client.
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

// TODO: Way of
func (l *SocketServer) ManageClients() {
	for {
		conn, err := l.sock.Accept() // Blocking
		if err != nil {
			slog.Error("client connection error", err)
			continue
		}

		// TODO: think if id for each client is a good idea
		go func(conn net.Conn) { // Go routine for each client
			defer conn.Close()

			err := l.ClientHandler(conn)
			if err != nil {
				slog.Error("client disconnected", err)
			} else {
				slog.Info("client disconnected")
			}
		}(conn)
	}
}

func (l *SocketServer) Close() {
	err := os.Remove(l.SocketPath)
	if err != nil {
		slog.Error("can't remove sock", err)
	}
}
