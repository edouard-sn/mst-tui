package client

import (
	"encoding/gob"
	"io"
	"mst-cli/ipc/types"
	"net"

	"github.com/anacrolix/torrent"
)

func handle(torrentClient *torrent.Client, conn net.Conn) error {
	dec := gob.NewDecoder(conn)

	for {
		message := &types.Packet{}
		if err := dec.Decode(message); err != nil && err != io.EOF {
			return err
		}
		go HandleCommands(message, torrentClient, conn)
		// TODO: Maybe tick wait
	}

}

func HandlerWithTorrentClientWrapper(torrentClient *torrent.Client) func(conn net.Conn) error {
	return func(conn net.Conn) error {
		return handle(torrentClient, conn)
	}
}
