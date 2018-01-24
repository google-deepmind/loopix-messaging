package main

import (
	"flag"
	"anonymous-messaging/publics"
	"anonymous-messaging/client"
	"anonymous-messaging/server"
)

func main() {

	typ := flag.String("typ", "", "A type of entity we want to run")
	id := flag.String("id", "", "Id of the entity we want to run")
	host := flag.String("host", "", "The host on which the entity is running")
	port := flag.String("port", "", "The port on which the entity is running")
	flag.Parse()

	switch *typ {
	case "client":
		pubC, privC := publics.GenerateKeyPair()
		client := client.NewClient(*id, *host, *port, pubC, privC, "./pki/database.db")
		client.Start()
	case "mix":
		pubM, privM := publics.GenerateKeyPair()
		mixServer := server.NewMixServer(*id, *host, *port, pubM, privM, "./pki/database.db")
		mixServer.Start()
	case "provider":
		pubP, privP := publics.GenerateKeyPair()
		providerServer := server.NewProviderServer(*id, *host, *port, pubP, privP, "./pki/database.db")
		providerServer.Start()
	}
}
