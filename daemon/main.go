package main

import (
	"mst-cli/daemon/client"
	"mst-cli/daemon/client/notification"
	"mst-cli/ipc/socket"
	"mst-cli/ipc/types"
	"os"

	"github.com/anacrolix/torrent"
	"golang.org/x/exp/slog"
)

func main() {
	initLogger()

	tClient := createTorrentClient()
	tRepo := client.NewTorrentRepository(tClient)
	ntf := notification.NewNotifier()

	go ntf.Listen()
	defer ntf.Done()

	server := &socket.Server{
		SocketPath:    "/tmp/mst.sock",
		ClientHandler: client.HandlerWrapper(tRepo, ntf),
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
	cfg.DataDir = os.TempDir()                          // NOTE: make it configurable later
	cfg.HeaderObfuscationPolicy.RequirePreferred = true // NOTE: Make it configurable later
	torrentClient, err := torrent.NewClient(cfg)
	must(err, "couldn't initialize torrent client")
	return torrentClient
}

func must(err error, msg string, args ...any) {
	if err != nil {
		slog.Error(msg, err, args...)
		os.Exit(1)
	}
}
