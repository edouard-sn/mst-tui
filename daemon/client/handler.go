package client

import (
	"encoding/gob"
	"io"
	"mst-cli/ipc/types"
	"net"
)

func handler(torrentClient string, conn net.Conn) error {
	var message *types.Packet = nil

	dec := gob.NewDecoder(conn)

	for {
		if message == nil {
			message = &types.Packet{}
			if err := dec.Decode(message); err != nil && err != io.EOF {
				return err
			}
		} else {
			// handleCommand(conn)
		}
	}

}

func HandlerWithTorrentClientWrapper(torrentClient string) func(conn net.Conn) error {
	return func(conn net.Conn) error {
		return handler(torrentClient, conn)
	}

}
