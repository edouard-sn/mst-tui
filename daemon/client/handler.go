package client

import (
	"encoding/gob"
	"io"
	"mst-cli/ipc/types"
	"net"

	"github.com/anacrolix/torrent"
	"golang.org/x/exp/slog"
)

func handle(torrentClient *torrent.Client, conn net.Conn) error {
	dec := gob.NewDecoder(conn)
	enc := gob.NewEncoder(conn)
	rm := &RequestManager{
		torrentClient: torrentClient,
	}

	for {
		message := &types.Packet{}
		if err := dec.Decode(message); err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		go func() {
			res := rm.HandleRequest(message)
			err := enc.Encode(res)
			if err != nil {
				slog.Error("gob encode", err)
			}
		}()
	}

}

func HandlerWithTorrentClientWrapper(torrentClient *torrent.Client) func(conn net.Conn) error {
	return func(conn net.Conn) error {
		return handle(torrentClient, conn)
	}
}
