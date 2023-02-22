package socket

import (
	"net"
	"os"

	"golang.org/x/exp/slog"
)

// Minimal generic unix socket server that will run ClientHandler on every connected client.
type Server struct {
	SocketPath    string
	ClientHandler func(net.Conn) error
	sock          net.Listener
}

func (l *Server) Start() error {
	sock, err := net.Listen("unix", l.SocketPath)
	if err != nil {
		return err
	}
	l.sock = sock
	return nil
}

func (l *Server) ManageClients() {
	for {
		conn, err := l.sock.Accept() // Blocking
		if err != nil {
			slog.Error("client connection error", err)
			continue
		}

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

func (l *Server) Close() {
	err := os.Remove(l.SocketPath)
	if err != nil {
		slog.Error("can't remove sock", err)
	}
}
