package client

import (
	"errors"
	"io"
	"mst-cli/ipc/types"
	"reflect"
	"time"

	"github.com/anacrolix/torrent"
	"golang.org/x/exp/slog"
)

type RequestManager struct {
	torrentClient *torrent.Client
}

// TODO: why not a class to have better parameter handling
func (rm *RequestManager) HandleRequest(msg *types.Packet) *types.Packet {
	var res any

	beforeCommand := time.Now()

	switch p := msg.Payload.(type) {
	case types.AddTorrentRequest:
		res = rm.addTorrent(&p)
	case types.RemoveTorrentRequest:
		res = rm.removeTorrent(&p)
	case types.ListTorrentsRequest:
		res = rm.listTorrents(&p)
	case types.SelectFilesToDownloadRequest:
		res = rm.selectFilesToDownload(&p)
	case types.PrioritizeFilesRequest:
		res = rm.prioritizeFiles(&p)
	case types.SequentialDownloadRequest:
		res = rm.sequentialDownload(&p)
	}

	timeSpent := time.Since(beforeCommand)
	// TODO: See if reflect is not shit
	slog.Debug("command done", "name", reflect.TypeOf(msg.Payload).Name(), "duration", timeSpent)

	return &types.Packet{
		CommandID: msg.CommandID,
		Payload:   res,
	}
}

func (rm *RequestManager) findTorrent(torrents []*torrent.Torrent, torrentID string) *torrent.Torrent {
	for _, t := range torrents {
		if t.InfoHash().String() == torrentID {
			return t
		}
	}
	return nil
}

func (rm *RequestManager) addTorrent(request *types.AddTorrentRequest) (res types.AddTorrentResponse) {
	t, err := rm.torrentClient.AddTorrentFromFile(request.Path)
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

func (rm *RequestManager) removeTorrent(request *types.RemoveTorrentRequest) (res types.ResponsePayload) {
	requestedTorrent := rm.findTorrent(rm.torrentClient.Torrents(), request.ID)

	if requestedTorrent == nil {
		res.Err = errors.New("no torrent with ID " + request.ID)
		return
	}

	slog.Debug("removed torrent", "name", requestedTorrent.Name())
	requestedTorrent.Drop()
	return
}

func (rm *RequestManager) listTorrents(_ *types.ListTorrentsRequest) (res types.ListTorrentsResponse) {
	torrents := rm.torrentClient.Torrents()
	res.Torrents = make([]types.CondensedTorrent, len(torrents))

	for i, t := range torrents {
		files := make([]types.CondensedFile, len(t.Files()))
		for j, file := range t.Files() {
			files[j] = types.CondensedFile{
				Name:            file.DisplayPath(),
				BytesDownloaded: file.BytesCompleted(),
				TotalBytes:      file.Length(),
			}
		}

		res.Torrents[i] = types.CondensedTorrent{
			Name:            t.Name(),
			Files:           files,
			BytesDownloaded: t.BytesCompleted(),
			TotalBytes:      t.Length(),
			Seeding:         t.Seeding(),
		}
	}

	return
}

func (rm *RequestManager) selectFilesToDownload(request *types.SelectFilesToDownloadRequest) (res types.ResponsePayload) {
	requestedTorrent := rm.findTorrent(rm.torrentClient.Torrents(), request.TorrentID)

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

func (rm *RequestManager) prioritizeFiles(msg *types.PrioritizeFilesRequest) (res types.ResponsePayload) {
	// NOTE: Will probably fuck off for now
	return types.ResponsePayload{}
}

func (rm *RequestManager) sequentialDownload(request *types.SequentialDownloadRequest) (res types.ResponsePayload) {
	requestedTorrent := rm.findTorrent(rm.torrentClient.Torrents(), request.ID)

	if requestedTorrent == nil {
		res.Err = errors.New("no torrend with ID " + request.ID)
		return
	}

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

	// TODO: Store reader or goroutine in map[torrentID][fileName] to close it and check if it's already in seq
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
