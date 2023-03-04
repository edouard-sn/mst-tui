package client

import (
	"errors"
	"io"
	"mst-cli/ipc/types"
	"reflect"
	"sync"
	"time"

	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/anacrolix/torrent/types/infohash"
	"golang.org/x/exp/slog"
)

// Stores a pointer to the File and a pointer to the Reader that reads sequentially the File, nil if not in sequential mode
type FileIndex struct {
	file      *torrent.File
	seqReader torrent.Reader
}

// TorrentRepository is used to handle request made to the daemon
// It stores a reference to the torrent client, and a torrent files index for easier lookup while processing requests
type TorrentRepository struct {
	tClient           *torrent.Client
	torrentFilesIndex map[metainfo.Hash]map[string]*FileIndex
	mu                *sync.Mutex
}

func NewTorrentRepository(c *torrent.Client) *TorrentRepository {
	return &TorrentRepository{
		tClient:           c,
		torrentFilesIndex: make(map[infohash.T]map[string]*FileIndex),
		mu:                &sync.Mutex{},
	}
}

func (rp *TorrentRepository) ProcessRequest(msg *types.Packet) (response *types.Packet, tellEverybody bool) {
	slog := slog.With("id", msg.CommandID)
	beforeCommand := time.Now()

	response = &types.Packet{CommandID: msg.CommandID}

	switch p := msg.Payload.(type) {
	case types.AddTorrentRequest:
		response.Payload = rp.addTorrent(&p)
		tellEverybody = true
	case types.RemoveTorrentRequest:
		response.Payload = rp.removeTorrent(&p)
	case types.ListTorrentsRequest:
		response.Payload = rp.listTorrents(&p)
	case types.SelectFilesToDownloadRequest:
		response.Payload = rp.selectFilesToDownload(&p)
	case types.PrioritizeFilesRequest:
		response.Payload = rp.prioritizeFiles(&p)
	case types.SequentialDownloadRequest:
		response.Payload = rp.sequentialDownload(&p)
	case types.CancelSequentialDownloadRequest:
		response.Payload = rp.cancelSequentialDownload(&p)
	default:
		slog.Warn("process request unimplemented payload type", "type", reflect.TypeOf(msg.Payload))
	}

	timeSpent := time.Since(beforeCommand)
	slog.Debug("command done", "name", reflect.TypeOf(msg.Payload).Name(), "duration", timeSpent)

	return response, tellEverybody
}

func (rp *TorrentRepository) getTorrentFromID(torrentID string) (res *torrent.Torrent, err error) {
	infohashID := infohash.T{}
	err = infohashID.FromHexString(torrentID)

	if err != nil {
		return nil, err
	}

	requestedTorrent, ok := rp.tClient.Torrent(infohashID)

	if !ok {
		return nil, errors.New("could not find torrent with id " + torrentID)
	}
	return requestedTorrent, nil
}

func (rp *TorrentRepository) getFileIndexMapFromID(torrentID string) (res map[string]*FileIndex, err error) {
	infohashID := infohash.T{}
	err = infohashID.FromHexString(torrentID)

	if err != nil {
		return nil, err
	}

	rp.mu.Lock()
	defer rp.mu.Unlock()

	fileIndexMap, ok := rp.torrentFilesIndex[infohashID]
	if !ok {
		return nil, errors.New("no torrent with id " + torrentID)
	}
	return fileIndexMap, nil
}

func (tp *TorrentRepository) torrentToCondensedTorrent(t *torrent.Torrent) types.CondensedTorrent {
	// TODO: Change Downloading fields when implementing pause
	torrentNotComplete := !t.Complete.Bool()

	files := make([]types.CondensedFile, len(t.Files()))
	for j, file := range t.Files() {
		files[j] = types.CondensedFile{
			Name:            file.DisplayPath(),
			BytesDownloaded: file.BytesCompleted(),
			TotalBytes:      file.Length(),
			Downloading:     torrentNotComplete && file.Priority() != torrent.PiecePriorityNone && file.BytesCompleted() != file.Length(),
		}
	}

	return types.CondensedTorrent{
		Name:            t.Name(),
		Files:           files,
		BytesDownloaded: t.BytesCompleted(),
		TotalBytes:      t.Length(),
		Seeding:         t.Seeding(),
		Downloading:     torrentNotComplete,
	}
}

func (rp *TorrentRepository) addTorrent(request *types.AddTorrentRequest) (res types.AddTorrentResponse) {
	t, err := rp.tClient.AddTorrentFromFile(request.Path)
	if err != nil {
		res.Err = err.Error()
		return res
	}

	tID := t.InfoHash()
	res.ID = tID.String()
	res.Torrent = rp.torrentToCondensedTorrent(t)

	// No defer Unlock() in case of that <-t.GotInfo() takes time
	rp.mu.Lock()
	if _, ok := rp.torrentFilesIndex[tID]; ok {
		rp.mu.Unlock()
		return
	}
	rp.torrentFilesIndex[tID] = make(map[string]*FileIndex)

	for _, file := range t.Files() {
		rp.torrentFilesIndex[tID][file.DisplayPath()] = &FileIndex{
			file:      file,
			seqReader: nil,
		}
	}
	rp.mu.Unlock()

	slog := slog.With("name", t.Name())
	slog.Debug("add torrent")

	<-t.GotInfo()
	slog.Debug("got torrent info")

	t.DownloadAll()
	return
}

func (rp *TorrentRepository) removeTorrent(request *types.RemoveTorrentRequest) (res types.ResponsePayload) {
	torrent, err := rp.getTorrentFromID(request.ID)
	if err != nil {
		res.Err = err.Error()
		return
	}

	torrent.Drop()
	slog.Debug("removed torrent", "name", torrent.Name())
	return
}

func (rp *TorrentRepository) listTorrents(_ *types.ListTorrentsRequest) (res types.ListTorrentsResponse) {
	torrents := rp.tClient.Torrents()
	res.Torrents = make([]types.CondensedTorrent, len(torrents))

	for i, t := range torrents {
		res.Torrents[i] = rp.torrentToCondensedTorrent(t)
	}
	return
}

func (rp *TorrentRepository) selectFilesToDownload(request *types.SelectFilesToDownloadRequest) (res types.ResponsePayload) {
	t, err := rp.getTorrentFromID(request.TorrentID)
	if err != nil {
		res.Err = err.Error()
		return
	}

	fileIndexMap, err := rp.getFileIndexMapFromID(request.TorrentID)
	if err != nil {
		res.Err = err.Error()
		return
	}

	for _, f := range t.Files() {
		f.SetPriority(torrent.PiecePriorityNone)
	}

	for _, fileName := range request.FileIDs {
		fileIndex, ok := fileIndexMap[fileName]

		if !ok || fileIndex == nil {
			slog.Error("fetching file ", errors.New(fileName))
			continue
		}

		f := fileIndex.file
		if f.Priority() == torrent.PiecePriorityNone {
			f.Download()
		}
	}
	return
}

func (rp *TorrentRepository) prioritizeFiles(request *types.PrioritizeFilesRequest) (res types.ResponsePayload) {
	fileIndexMap, err := rp.getFileIndexMapFromID(request.ID)
	if err != nil {
		res.Err = err.Error()
		return
	}

	for _, f := range request.Files {
		fileIndex, ok := fileIndexMap[f.FileName]

		if !ok || fileIndex == nil {
			slog.Error("fetching file ", errors.New(f.FileName))
			continue
		}

		file := fileIndex.file
		switch f.Priority {
		case types.High:
			file.SetPriority(torrent.PiecePriorityHigh)
		case types.Medium:
			file.SetPriority(torrent.PiecePriorityNext)
		case types.Low:
			file.SetPriority(torrent.PiecePriorityNormal)

		}
	}
	return
}

func (rp *TorrentRepository) sequentialDownload(request *types.SequentialDownloadRequest) (res types.ResponsePayload) {
	fileIndexMap, err := rp.getFileIndexMapFromID(request.ID)
	if err != nil {
		res.Err = err.Error()
		return
	}

	fileIndex, ok := fileIndexMap[request.FileName]

	if !ok || fileIndex == nil {
		res.Err = "no file with name " + request.FileName
		return
	}

	if fileIndex.seqReader != nil {
		res.Err = "already in sequential mode"
		return
	}

	reader := fileIndex.file.NewReader()
	fileIndex.seqReader = reader
	go func() {
		requestedFile := fileIndex.file

		slog := slog.With("name", fileIndex.file.Torrent().Name(), "file", requestedFile.DisplayPath())
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

func (rp *TorrentRepository) cancelSequentialDownload(request *types.CancelSequentialDownloadRequest) (res types.ResponsePayload) {
	fileIndexMap, err := rp.getFileIndexMapFromID(request.ID)
	if err != nil {
		res.Err = err.Error()
		return
	}

	fileIndex, ok := fileIndexMap[request.FileName]

	if !ok || fileIndex == nil {
		res.Err = "no file with name " + request.FileName
		return
	}

	if fileIndex.seqReader == nil {
		return
	}

	requestedFile := fileIndex.file

	slog := slog.With("name", fileIndex.file.Torrent().Name(), "file", requestedFile.DisplayPath())
	slog.Debug("cancel sequential download")

	if err := fileIndex.seqReader.Close(); err != nil {
		res.Err = err.Error()
	}
	return
}
