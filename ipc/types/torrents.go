package types

type CondensedFile struct {
	Name            string
	BytesDownloaded int64
	TotalBytes      int64
	Downloading     bool
}

type CondensedTorrent struct {
	Name            string
	Files           []CondensedFile
	BytesDownloaded int64
	TotalBytes      int64
	Downloading     bool
	Seeding         bool
}
