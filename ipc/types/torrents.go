package types

type CondensedTorrent struct {
	Name            string
	Path            string
	FileNames       []string
	BytesDownloaded uint32
	BytesToDownload uint32
	// TODO: wawi continue this type
}

// TODO: File type
