package main

import "zues/server"

func main() {
	zuesServer := server.New(nil, "")

	// Start the server
	zuesServer.Start()
}

