package socket

import (
	"net"
	"reflect"
	"strings"
	"sync"

	"golang.org/x/exp/slog"
)

// Minimal generic unix socket server that will run ClientLogic on every connected client.
// It only opens a socket, accept new connections and calls a client handler
// There is no connection management apart from what is sent to the client handler
type Server struct {
	SocketPath  string
	ClientLogic func(net.Conn) error
	sock        net.Listener
	wg          sync.WaitGroup
}

func (l *Server) Start() error {
	sock, err := net.Listen("unix", l.SocketPath)
	if err != nil {
		return err
	}
	l.sock = sock
	l.wg = sync.WaitGroup{}
	slog.Info("server listening", "socket", l.SocketPath)
	return nil
}

// This function handles all the client connection after it got accepted
// It uses a ptr doneClientChannels as a parameter in case the server wants to close before the client,
// in which case every "close" watcher will be stored in this array. mu comes with the array to be thread safe
func (l *Server) clientConnectionLifetime(conn net.Conn, done chan int) {
	defer l.wg.Done()
	slog.Info("client connected")

	// Go routine that will listen to the "done" channel which is here to inform us when we need to close the conn
	// Could be either if the server is stopping or if we lost connection to the client
	go func() {
		defer l.wg.Done()
		<-done
		err := conn.Close()
		if err != nil {
			slog.Error("could not close client connection", err)
		}
	}()

	err := l.ClientLogic(conn)

	// If the connection is closed, it is not considered an error (as we probably closed it ourself)
	if err != nil && !l.errIsConnectionClosed(err) {
		slog.Error("client disconnected", err)
	} else {
		slog.Info("client disconnected")
	}

	// If no error or connection is not closed, let's notify our closer/done goroutine
	if err == nil || !l.errIsConnectionClosed(err) {
		done <- 0
		if err != nil {
			slog.Error("anormaly disconnected from client", err)
		}
	}
}

// net library asks us to check the error as so
func (l *Server) errIsConnectionClosed(err error) bool {
	return strings.Contains(err.Error(), "use of closed network connection")
}

func (l *Server) ManageClients() {
	doneClientChannels := make([]chan int, 0)

	for {
		conn, err := l.sock.Accept()
		if err != nil {
			if l.errIsConnectionClosed(err) {
				break
			}
			slog.Error("client connection error", err, "type", reflect.TypeOf(err))
			continue
		}

		done := make(chan int, 1)
		doneClientChannels = append(doneClientChannels, done)

		// For each client, two go routines will be launched
		// One for overall client logic and one that watches when to close the connection
		l.wg.Add(2)
		go l.clientConnectionLifetime(conn, done)
	}

	// Makes sure that every connection will be closed. No-op if it already has been done
	for _, done := range doneClientChannels {
		done <- 0
		close(done)
	}

	// Wait for every go routines that has been started by the server
	l.wg.Wait()
}

func (l *Server) Close() {
	l.sock.Close()
}
