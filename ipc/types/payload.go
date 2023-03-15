package types

import (
	"encoding/gob"
)

type ResponsePayload struct {
	Err string
}

type AddTorrentRequest struct {
	Path string
}
type AddTorrentResponse struct {
	ResponsePayload
	ID      string
	Torrent CondensedTorrent
}

type RemoveTorrentRequest struct {
	ID string
}

type ListTorrentsRequest struct{}
type ListTorrentsResponse struct {
	ResponsePayload
	Torrents []CondensedTorrent
}

type SelectFilesToDownloadRequest struct {
	TorrentID string
	FileIDs   []string
}

type FilePriorityName string

const (
	High   FilePriorityName = "high"
	Medium FilePriorityName = "mid"
	Low    FilePriorityName = "low"
)

type FilePriority struct {
	FileName string
	Priority FilePriorityName
}
type PrioritizeFilesRequest struct {
	ID    string
	Files []FilePriority
}

type SequentialDownloadRequest struct {
	ID       string
	FileName string
}

type CancelSequentialDownloadRequest struct {
	ID       string
	FileName string
}

// Registers every payload types
func RegisterEveryPayloadToGob() {
	gob.Register(AddTorrentRequest{})
	gob.Register(RemoveTorrentRequest{})
	gob.Register(ListTorrentsRequest{})
	gob.Register(SelectFilesToDownloadRequest{})
	gob.Register(FilePriority{})
	gob.Register(PrioritizeFilesRequest{})
	gob.Register(SequentialDownloadRequest{})
	gob.Register(CancelSequentialDownloadRequest{})

	gob.Register(ResponsePayload{})
	gob.Register(AddTorrentResponse{})
	gob.Register(ListTorrentsResponse{})
}
