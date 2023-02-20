package main

import (
	"log"
	"mst-cli/daemon/client"
	"mst-cli/ipc/server"
	"mst-cli/ipc/types"

	"github.com/anacrolix/torrent"
)

func must(err error, msg string) {
	if err != nil {
		log.Fatal(msg, ":", err)
	}
}

func main() {
	// TODO: Change DataDir in cfg :(
	torrentClient, err := torrent.NewClient(torrent.NewDefaultClientConfig()) // NOTE: wawi look how many fucks we are giving for the config rn
	must(err, "couldn't initialize torrent client: "+err.Error())

	types.RegisterEveryPayloadToGob()

	server := &server.SocketServer{
		SocketPath:    "/tmp/mst.sock",
		ClientHandler: client.HandlerWithTorrentClientWrapper(torrentClient),
	}
	server.Start()
	server.ManageClients()

	must(server.Close(), "server couldn't close properly ("+server.SocketPath+")")
}
