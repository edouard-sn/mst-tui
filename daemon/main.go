package main

import (
	"mst-cli/daemon/client"
	"mst-cli/ipc/server"
	"mst-cli/ipc/types"
	"os"

	"github.com/anacrolix/torrent"
	"golang.org/x/exp/slog"
)

func main() {
	initLogger()
	torrentClient := initTorrentClient()

	types.RegisterEveryPayloadToGob()

	server := &server.SocketServer{
		SocketPath:    "/tmp/mst.sock",
		ClientHandler: client.HandlerWithTorrentClientWrapper(torrentClient),
	}
	must(server.Start(), "couldn't start server")
	defer server.Close()

	server.ManageClients()
}

func initLogger() {
	opts := slog.HandlerOptions{
		Level: slog.LevelDebug,
	}
	textHandler := opts.NewTextHandler(os.Stdout)
	slogger := slog.New(textHandler)
	slog.SetDefault(slogger)
}

func initTorrentClient() *torrent.Client {
	cfg := torrent.NewDefaultClientConfig()
	cfg.DataDir = os.TempDir()
	torrentClient, err := torrent.NewClient(cfg) // NOTE: wawi look how many fucks we are giving for the config rn
	must(err, "couldn't initialize torrent client:")
	return torrentClient
}

func must(err error, msg string) {
	if err != nil {
		slog.Error(msg, err)
		os.Exit(1)
	}
}
