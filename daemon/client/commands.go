package client

import (
	"encoding/gob"
	"mst-cli/ipc/types"
	"net"
)

func HandleCommands(msg *types.Packet, torrentClient string, conn net.Conn) {
	enc := gob.NewEncoder(conn)
	var res any

	switch msg.Payload.(type) {
	case types.AddTorrentRequest:
		res = AddTorrent(msg, torrentClient)
	case types.RemoveTorrentRequest:
		res = RemoveTorrent(msg, torrentClient)
	case types.ListTorrentsRequest:
		res = ListTorrents(msg, torrentClient)
	case types.SelectFilesToDownloadRequest:
		res = SelectFilesToDownload(msg, torrentClient)
	case types.PrioritizeFilesRequest:
		res = PrioritizeFiles(msg, torrentClient)
	case types.SequentialDownloadRequest:
		res = SequentialDownload(msg, torrentClient)
	}

	enc.Encode(res)
}

func AddTorrent(msg *types.Packet, torrentClient string) types.AddTorrentResponse {
	return types.AddTorrentResponse{}
}

func RemoveTorrent(msg *types.Packet, torrentClient string) types.ResponsePayload {

	return types.ResponsePayload{}
}

func ListTorrents(msg *types.Packet, torrentClient string) types.ListTorrentsResponse {
	return types.ListTorrentsResponse{}
}

func SelectFilesToDownload(msg *types.Packet, torrentClient string) types.ResponsePayload {
	return types.ResponsePayload{}
}

func PrioritizeFiles(msg *types.Packet, torrentClient string) types.ResponsePayload {
	return types.ResponsePayload{}
}

func SequentialDownload(msg *types.Packet, torrentClient string) types.ResponsePayload {
	return types.ResponsePayload{}
}
