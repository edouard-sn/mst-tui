package types

type CondensedTorrent struct {
	Name            string
	Path            string
	FileNames       []string
	BytesDownloaded int64
	TotalBytes      int64
	// TODO: wawi continue this type
}

// TODO: File type
