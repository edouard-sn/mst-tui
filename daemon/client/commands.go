package client

import (
	"encoding/gob"
	"mst-cli/ipc/types"
	"net"

	"github.com/anacrolix/torrent"
)

func HandleCommands(msg *types.Packet, torrentClient *torrent.Client, conn net.Conn) {
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

func AddTorrent(msg *types.Packet, torrentClient *torrent.Client) types.AddTorrentResponse {
	t, err := torrentClient.AddTorrentFromFile(msg.Payload.(types.AddTorrentRequest).Path)

	res := types.AddTorrentResponse{}
	res.Err = err
	if err != nil {
		return res
	}

	res.ID = t.InfoHash().String()
	return res
}

func RemoveTorrent(msg *types.Packet, torrentClient *torrent.Client) types.ResponsePayload {
	request := msg.Payload.(types.RemoveTorrentRequest)
	torrents := torrentClient.Torrents()

	for _, t := range torrents {
		if t.InfoHash().String() == request.ID {
			t.Drop()
		}
	}

	return types.ResponsePayload{
		Err: nil,
	}
}

func ListTorrents(msg *types.Packet, torrentClient *torrent.Client) types.ListTorrentsResponse {
	// torrents := torrentClient.Torrents()

	return types.ListTorrentsResponse{}
}

func SelectFilesToDownload(msg *types.Packet, torrentClient *torrent.Client) types.ResponsePayload {
	return types.ResponsePayload{}
}

func PrioritizeFiles(msg *types.Packet, torrentClient *torrent.Client) types.ResponsePayload {
	return types.ResponsePayload{}
}

func SequentialDownload(msg *types.Packet, torrentClient *torrent.Client) types.ResponsePayload {
	return types.ResponsePayload{}
}
