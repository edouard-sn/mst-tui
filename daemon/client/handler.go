package client

import (
	"encoding/gob"
	"io"
	"mst-cli/daemon/client/notification"
	"mst-cli/ipc/types"
	"net"
	"sync"
)

func handle(conn net.Conn, tr *TorrentRepository, nt *notification.Notifier) error {
	dec := gob.NewDecoder(conn)
	enc := gob.NewEncoder(conn)
	wg := sync.WaitGroup{}

	nt.AddConn(conn, enc)
	defer nt.RemoveConn(conn)
	defer wg.Wait()

	for {
		message := &types.Packet{}
		if err := dec.Decode(message); err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			response, tellEverybody := tr.ProcessRequest(message)

			if tellEverybody {
				nt.GetNotifyAllChannel() <- notification.NotifyAllRequest{Message: response}
			} else { // This sounds stupid
				nt.GetNotifyChannel() <- notification.NotifyRequest{Message: response, Conn: conn}
			}
		}()
	}
}

func HandlerWrapper(tr *TorrentRepository, nt *notification.Notifier) func(net.Conn) error {
	return func(conn net.Conn) error {
		return handle(conn, tr, nt)
	}
}
