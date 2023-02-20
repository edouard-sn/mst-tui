package types

type CommandType uint8

const (
	AddTorrent CommandType = iota
	RemoveTorrent
	ListTorrents
	SelectFilesToDownload
	PrioritizeFiles
	SequentialDownload
)

type Packet struct {
	CommandId []byte
	Payload   any
}
