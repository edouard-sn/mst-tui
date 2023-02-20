package client

import (
	"encoding/gob"
	"io"
	"mst-cli/ipc/types"
	"net"
)

func HandleCommand(cmd *types.Packet, c *chan interface{}, conn net.Conn) error {
	switch cmd.Payload.(type) {
	case types.AddTorrentRequest:
	case types.RemoveTorrentRequest:
	case types.ListTorrentsRequest:
	case types.SelectFilesToDownloadRequest:
	case types.PrioritizeFilesRequest:
	case types.SequentialDownloadRequest:
	}
	return nil
}

func Handler(c *chan interface{}, conn net.Conn) error {
	var message *types.Packet = nil

	dec := gob.NewDecoder(conn)

	for {
		if message == nil {
			message = &types.Packet{}
			if err := dec.Decode(message); err != nil && err != io.EOF {
				return err
			}
		} else {
			HandleCommand(message, c, conn)
		}
	}
}
