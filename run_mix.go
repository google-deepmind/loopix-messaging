package main

import (
	"anonymous-messaging/server"
	"anonymous-messaging/publics"
)

func main() {
	// args := os.Args

	// read in key pair or generate a new one
	pubM, privM := publics.GenerateKeyPair()
	mixServer := server.NewMixServer("Mix", "localhost", "9999", pubM, privM, "./pki/database.db")
	mixServer.Start()
}
