package main

import (
	"anonymous-messaging/server"
	"anonymous-messaging/publics"
)

func main() {
	// args := os.Args

	pubM, privM := publics.GenerateKeyPair()
	mixServer := server.NewMixServer("Mix", "localhost", "9999", pubM, privM, "./pki/database.db")
	mixServer.Start()
}
