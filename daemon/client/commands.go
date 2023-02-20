package client

import (
	"encoding/gob"
	"fmt"
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

	err := enc.Encode(res)
	if err != nil {
		fmt.Printf("are you fucking kidding me, gob.Encode failed: %v", err)
	}
}

func AddTorrent(msg *types.Packet, torrentClient *torrent.Client) (res types.AddTorrentResponse) {
	t, err := torrentClient.AddTorrentFromFile(msg.Payload.(types.AddTorrentRequest).Path)

	res.Err = err
	if err != nil {
		return res
	}
	t.DownloadAll()
	res.ID = t.InfoHash().String()
	return res
}

func RemoveTorrent(msg *types.Packet, torrentClient *torrent.Client) types.ResponsePayload {
	request := msg.Payload.(types.RemoveTorrentRequest)
	torrents := torrentClient.Torrents()

	for _, t := range torrents {
		if t.InfoHash().String() == request.ID {
			t.Drop()
			break
		}
	}

	return types.ResponsePayload{
		Err: nil,
	}
}

func ListTorrents(msg *types.Packet, torrentClient *torrent.Client) (res types.ListTorrentsResponse) {
	torrents := torrentClient.Torrents()
	res.Torrents = make([]types.CondensedTorrent, len(torrents))

	for i, t := range torrents {

		filePaths := make([]string, len(t.Files()))
		for j, file := range t.Files() {
			filePaths[j] = file.DisplayPath() // NOTE: Si dieu le veut c'est pas mal
		}

		receivedBytes := t.Stats().BytesReadData
		res.Torrents[i] = types.CondensedTorrent{
			Name:            t.Name(),
			FileNames:       filePaths, // good enough
			BytesDownloaded: (&receivedBytes).Int64(),
			TotalBytes:      t.Length(),
		}
	}

	return res
}

func SelectFilesToDownload(msg *types.Packet, torrentClient *torrent.Client) (res types.ResponsePayload) {
	request := msg.Payload.(types.SelectFilesToDownloadRequest)
	var requestedTorrent *torrent.Torrent
	for _, t := range torrentClient.Torrents() {
		if t.InfoHash().String() == request.TorrentID {
			requestedTorrent = t
			break
		}
	}

OUTER:
	for _, f := range requestedTorrent.Files() {
		for _, fileName := range request.FileIDs {
			if f.Priority() == torrent.PiecePriorityNone && f.DisplayPath() == fileName {
				f.Download()
				continue OUTER
			}
		}
		f.SetPriority(torrent.PiecePriorityNone)
	}
	return types.ResponsePayload{}
}

func PrioritizeFiles(msg *types.Packet, torrentClient *torrent.Client) (res types.ResponsePayload) {
	// NOTE: Will probably fuck off for now
	return types.ResponsePayload{}
}

// NOTE: This will be fake sequential download, it will prioritize the 15 first percent of the download
// NOTE: I can probably make it work to prioritize along the way we'll see about all this
func SequentialDownload(msg *types.Packet, torrentClient *torrent.Client) (res types.ResponsePayload) {
	request := msg.Payload.(types.SequentialDownloadRequest)
	var requestedTorrent *torrent.Torrent
	fileName := "lol"

	for _, t := range torrentClient.Torrents() {
		if t.InfoHash().String() == request.ID {
			requestedTorrent = t
			break
		}
	}

	for _, f := range requestedTorrent.Files() {
		if f.DisplayPath() == fileName {
			pieceBytesRatio := int64(requestedTorrent.NumPieces()) / requestedTorrent.Length()
			firstPiece := f.Offset() * pieceBytesRatio
			lastPiece := (f.Offset() + f.Length()) * pieceBytesRatio

			for i := firstPiece; i <= lastPiece*15/100; i++ {
				requestedTorrent.Piece(int(i)).SetPriority(torrent.PiecePriorityNow)
			}
		}
	}

	return types.ResponsePayload{
		Err: nil,
	}
}
