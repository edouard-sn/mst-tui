package main

import (
	"log"
	"mst-cli/daemon/client"
	"mst-cli/ipc/server"
	"mst-cli/ipc/types"
)

func must(err error, msg string) {
	if err != nil {
		log.Fatal(msg, ":", err)
	}
}

func main() {
	types.RegisterEveryPayloadToGob()

	server := &server.SocketServer{
		SocketPath:    "/tmp/mst.sock",
		ClientHandler: client.HandlerWithTorrentClientWrapper("placeholder"),
	}
	server.Start()
	server.ManageClients()

	must(server.Close(), "server couldn't close properly ("+server.SocketPath+")")
}
