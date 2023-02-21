package client

import (
	"encoding/gob"
	"errors"
	"io"
	"mst-cli/ipc/types"
	"net"
	"reflect"

	"github.com/anacrolix/torrent"
	"golang.org/x/exp/slog"
)

// TODO: why not a class to have better parameter handling
func HandleCommands(msg *types.Packet, torrentClient *torrent.Client, conn net.Conn, enc *gob.Encoder) {
	var res any

	// TODO: See if reflect is not shit
	slog.Debug("command recieved", "name", reflect.TypeOf(msg.Payload).Name())

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

	// NOTE: Should we send the packet in the "HandleCommand" function?
	packet := &types.Packet{
		CommandID: msg.CommandID,
		Payload:   res,
	}
	err := enc.Encode(packet)
	if err != nil {
		slog.Error("gob encode", err)
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

	res.Err = err
	if err != nil {
		return res
	}
	slog := slog.With("name", t.Name())
	slog.Debug("add torrent")
	<-t.GotInfo()
	slog.Debug("got torrent info")
	t.DownloadAll()

	res.ID = t.InfoHash().String()
	return
}

func RemoveTorrent(msg *types.Packet, torrentClient *torrent.Client) (res types.ResponsePayload) {
	request := msg.Payload.(types.RemoveTorrentRequest)
	requestedTorrent := findTorrent(torrentClient.Torrents(), request.ID)

	if requestedTorrent == nil {
		res.Err = errors.New("no torrend with ID " + request.ID)
		return
	}

	slog.Debug("removed torrent", "name", requestedTorrent.Name())
	requestedTorrent.Drop()
	return
}

func ListTorrents(msg *types.Packet, torrentClient *torrent.Client) (res types.ListTorrentsResponse) {
	torrents := torrentClient.Torrents()
	res.Torrents = make([]types.CondensedTorrent, len(torrents))

	for i, t := range torrents {
		files := make([]types.CondensedFile, len(t.Files()))
		for j, file := range t.Files() {
			files[j] = types.CondensedFile{
				Name:            file.DisplayPath(),
				BytesDownloaded: file.BytesCompleted(),
				TotalBytes:      file.Length(),
			}
			// NOTE: Si dieu le veut c'est pas mal
			slog.Debug("list torrent", "name", t.Name(), "file", files[j])
		}

		res.Torrents[i] = types.CondensedTorrent{
			Name:            t.Name(),
			Files:           files, // good enough
			BytesDownloaded: t.BytesCompleted(),
			TotalBytes:      t.Length(),
		}
	}

	return
}

func SelectFilesToDownload(msg *types.Packet, torrentClient *torrent.Client) (res types.ResponsePayload) {
	request := msg.Payload.(types.SelectFilesToDownloadRequest)
	requestedTorrent := findTorrent(torrentClient.Torrents(), request.TorrentID)

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

	if requestedFile == nil {
		res.Err = errors.New("no file with name " + request.FileName)
		return
	}

	// Store reader or goroutine in map[torrentID][fileName] to close it and check if it's already in seq
	reader := requestedFile.NewReader()
	go func() {
		slog := slog.With("name", requestedTorrent.Name(), "file", requestedFile.DisplayPath())
		slog.Debug("sequential download")
		n, err := io.Copy(io.Discard, reader)
		if err != nil {
			slog.Error("sequential download", err)
		} else {
			slog.Debug("finished sequential download", "bytes-sequential", n, "bytes-total", requestedFile.Length())
		}
	}()
	return
}
