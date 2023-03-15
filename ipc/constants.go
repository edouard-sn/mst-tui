package ipc

// Informative values
const (
	SocketPath = "/tmp/mst.sock"
)

// Error values
const (
	ErrAlreadyInTheRequestedState string = "already in the requested state"
	ErrFileNotFound               string = "file not found"
	ErrTorrentNotFound            string = "torrent not found"
	ErrInternal                   string = "server internal error: "
)
