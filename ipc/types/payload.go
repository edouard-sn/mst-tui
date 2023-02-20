package types

import "encoding/gob"

type ResponsePayload struct {
	Err error
}

type AddTorrentRequest struct {
	Path string
}
type AddTorrentResponse struct {
	ResponsePayload
	ID string
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

type PrioritizeFilesRequest struct {
	ID        string
	fileNames []string
}

type SequentialDownloadRequest struct {
	ID string
}

func RegisterEveryPayloadToGob() {
	gob.Register(AddTorrentRequest{})
	gob.Register(RemoveTorrentRequest{})
	gob.Register(ListTorrentsRequest{})
	gob.Register(SelectFilesToDownloadRequest{})
	gob.Register(PrioritizeFilesRequest{})
	gob.Register(SequentialDownloadRequest{})

	gob.Register(ResponsePayload{})
	gob.Register(AddTorrentResponse{})
	gob.Register(ListTorrentsResponse{})
}
