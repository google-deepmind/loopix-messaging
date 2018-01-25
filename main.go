package main

import (
	"flag"
	"anonymous-messaging/client"
	"anonymous-messaging/server"
	"anonymous-messaging/sphinx"
	"anonymous-messaging/pki"
	"fmt"
	"anonymous-messaging/publics"
)

func main() {

	typ := flag.String("typ", "", "A type of entity we want to run")
	id := flag.String("id", "", "Id of the entity we want to run")
	host := flag.String("host", "", "The host on which the entity is running")
	port := flag.String("port", "", "The port on which the entity is running")
	providerId := flag.String("provider", "", "The port on which the entity is running")
	flag.Parse()

	switch *typ {
	case "client":
		db := pki.OpenDatabase("pki/database.db", "sqlite3")
		row := db.QueryRow("SELECT Info FROM Providers WHERE ProviderId = ?", providerId)

		var results []byte
		err := row.Scan(&results)
		if err != nil {
			fmt.Println(err)
		}
		providerInfo, err := publics.MixPubsFromBytes(results)

		pubC, privC := sphinx.GenerateKeyPair()
		client := client.NewClient(*id, *host, *port, pubC, privC, "./pki/database.db", providerInfo)
		client.Start()
	case "mix":
		pubM, privM := sphinx.GenerateKeyPair()
		mixServer := server.NewMixServer(*id, *host, *port, pubM, privM, "./pki/database.db")
		mixServer.Start()
	case "provider":
		pubP, privP := sphinx.GenerateKeyPair()
		providerServer := server.NewProviderServer(*id, *host, *port, pubP, privP, "./pki/database.db")
		providerServer.Start()
	}
}
