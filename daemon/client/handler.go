package client

import (
	"encoding/gob"
	"io"
	"mst-cli/daemon/client/notification"
	"mst-cli/ipc/types"
	"net"
	"reflect"

	"golang.org/x/exp/slog"
)

func handle(conn net.Conn, tr *TorrentRepository, nt *notification.Notifier) error {
	dec := gob.NewDecoder(conn)
	enc := gob.NewEncoder(conn)

	nt.AddConn(conn, enc)
	defer nt.RemoveConn(conn)

	for {
		message := &types.Packet{}
		if err := dec.Decode(message); err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}

		go func() {
			msg, tellEverybody := tr.ProcessRequest(message)

			if tellEverybody {
				nt.GetNotifyAllChannel() <- notification.NotifyAllRequest{Message: msg}
				return
			}

			err := enc.Encode(msg)
			if err != nil {
				slog.Error("gob encoding", err, "type", reflect.TypeOf(msg.Payload), "payload", msg.Payload)
			}
		}()
	}
}

func HandlerWrapper(tr *TorrentRepository, nt *notification.Notifier) func(net.Conn) error {
	return func(conn net.Conn) error {
		return handle(conn, tr, nt)
	}
}
