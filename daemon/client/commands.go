package client

import (
	"encoding/gob"
	"errors"
	"fmt"
	"io"
	"log"
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

func findTorrent(torrents []*torrent.Torrent, torrentID string) *torrent.Torrent {
	for _, t := range torrents {
		if t.InfoHash().String() == torrentID {
			return t
		}
	}
	return nil
}

func AddTorrent(msg *types.Packet, torrentClient *torrent.Client) (res types.AddTorrentResponse) {
	t, err := torrentClient.AddTorrentFromFile(msg.Payload.(types.AddTorrentRequest).Path)
	res.CommandID = msg.CommandID

	res.Err = err
	if err != nil {
		return res
	}

	<-t.GotInfo()
	t.DownloadAll()
	res.ID = t.InfoHash().String()[:16]
	return
}

func RemoveTorrent(msg *types.Packet, torrentClient *torrent.Client) (res types.ResponsePayload) {
	request := msg.Payload.(types.RemoveTorrentRequest)
	requestedTorrent := findTorrent(torrentClient.Torrents(), request.ID)
	res.CommandID = msg.CommandID

	if requestedTorrent == nil {
		res.Err = errors.New("no torrend with ID " + request.ID)
		return
	}

	requestedTorrent.Drop()
	return
}

func ListTorrents(msg *types.Packet, torrentClient *torrent.Client) (res types.ListTorrentsResponse) {
	torrents := torrentClient.Torrents()
	res.Torrents = make([]types.CondensedTorrent, len(torrents))
	res.CommandID = msg.CommandID

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

	return
}

func SelectFilesToDownload(msg *types.Packet, torrentClient *torrent.Client) (res types.ResponsePayload) {
	request := msg.Payload.(types.SelectFilesToDownloadRequest)
	requestedTorrent := findTorrent(torrentClient.Torrents(), request.TorrentID)
	res.CommandID = msg.CommandID

	if requestedTorrent == nil {
		res.Err = errors.New("no torrend with ID " + request.TorrentID)
		return
	}

	// NOTE: This is shit maybe i'll better it later
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
	return
}

func PrioritizeFiles(msg *types.Packet, torrentClient *torrent.Client) (res types.ResponsePayload) {
	// NOTE: Will probably fuck off for now
	return types.ResponsePayload{}
}

func SequentialDownload(msg *types.Packet, torrentClient *torrent.Client) (res types.ResponsePayload) {
	request := msg.Payload.(types.SequentialDownloadRequest)
	requestedTorrent := findTorrent(torrentClient.Torrents(), request.ID)
	res.CommandID = msg.CommandID

	if requestedTorrent == nil {
		res.Err = errors.New("no torrend with ID " + request.ID)
		return
	}

	// TODO : To test, and yes anacrolix, i'm going to use separate programs to stream it, fight me
	var requestedFile *torrent.File
	for _, f := range requestedTorrent.Files() {
		if f.DisplayPath() == request.FileName {
			requestedFile = f
			break
		}
	}

	reader := requestedFile.NewReader()
	go func() {
		log.Print("Starting sequential download for ", requestedTorrent.Name(), ":", requestedFile.FileInfo().Path)
		n, err := io.Copy(io.Discard, reader)
		if err != nil {
			log.Printf("caught seq error: %v", err)
		}
		log.Print("Finished seq download for ", requestedTorrent.Name(), ":", requestedFile.FileInfo().Path, " ", n, "bytes downloaded")
	}()

	// NOTE: Naive method below
	// NOTE: This will be fake sequential download, it will prioritize the 15 first percent of the download
	// NOTE: I can probably make it work to prioritize along the way we'll see about all this
	// for _, f := range requestedTorrent.Files() {
	// 	if f.DisplayPath() == request.FileName {
	// 		pieceBytesRatio := int64(requestedTorrent.NumPieces()) / requestedTorrent.Length()
	// 		firstPiece := f.Offset() * pieceBytesRatio
	// 		lastPiece := (f.Offset() + f.Length()) * pieceBytesRatio
	// 		for i := firstPiece; i <= lastPiece*15/100; i++ {
	// 			requestedTorrent.Piece(int(i)).SetPriority(torrent.PiecePriorityNow)
	// 		}
	// 	}
	// }
	return
}
