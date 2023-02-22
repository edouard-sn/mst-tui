package main

import (
	"mst-cli/daemon/client"
	"mst-cli/ipc/socket"
	"mst-cli/ipc/types"
	"os"

	"github.com/anacrolix/torrent"
	"golang.org/x/exp/slog"
)

func main() {
	initLogger()

	tClient := createTorrentClient()
	server := &socket.Server{
		SocketPath:    "/tmp/mst.sock",
		ClientHandler: client.HandlerWithTorrentClientWrapper(tClient),
	}

	must(server.Start(), "couldn't start server")
	defer server.Close()

	types.RegisterEveryPayloadToGob()
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

func createTorrentClient() *torrent.Client {
	cfg := torrent.NewDefaultClientConfig()
	cfg.DataDir = os.TempDir()
	cfg.HeaderObfuscationPolicy.RequirePreferred = true // NOTE: Make it configurable later
	torrentClient, err := torrent.NewClient(cfg)
	must(err, "couldn't initialize torrent client")
	return torrentClient
}

func must(err error, msg string) {
	if err != nil {
		slog.Error(msg, err)
		os.Exit(1)
	}
}
