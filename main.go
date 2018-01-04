package main

import (
	"anonymous-messaging/server"
)

func main() {
	// args := os.Args

	mixServer := server.NewMixServer("Mix", "localhost", "9999", 0, 0, "./pki/database.db")
	mixServer.Start()
}
