package main

import (
	"flag"
	"anonymous-messaging/client"
	"anonymous-messaging/server"
	"anonymous-messaging/sphinx"
	"anonymous-messaging/pki"
	"fmt"
	"anonymous-messaging/config"
)

func main() {

	fmt.Println("Starting!!")

	typ := flag.String("typ", "", "A type of entity we want to run")
	id := flag.String("id", "", "Id of the entity we want to run")
	host := flag.String("host", "", "The host on which the entity is running")
	port := flag.String("port", "", "The port on which the entity is running")
	providerId := flag.String("provider", "", "The port on which the entity is running")
	flag.Parse()

	switch *typ {
	case "client":
		db, err := pki.OpenDatabase("pki/database.db", "sqlite3")

		if err != nil{
			panic(err)
		}


		row := db.QueryRow("SELECT Config FROM Providers WHERE Id = ?", providerId)

		var results []byte
		err = row.Scan(&results)
		if err != nil {
			fmt.Println(err)
		}
		providerInfo, err := config.MixPubsFromBytes(results)

		pubC, privC, err := sphinx.GenerateKeyPair()
		if err != nil{
			panic(err)
		}

		client := client.NewClient(*id, *host, *port, pubC, privC, "./pki/database.db", providerInfo)
		client.Start()
	case "mix":
		pubM, privM, err := sphinx.GenerateKeyPair()
		if err != nil{
			panic(err)
		}

		mixServer := server.NewMixServer(*id, *host, *port, pubM, privM, "./pki/database.db")
		mixServer.Start()
	case "provider":
		pubP, privP, err := sphinx.GenerateKeyPair()
		if err != nil{
			panic(err)
		}

		providerServer := server.NewProviderServer(*id, *host, *port, pubP, privP, "./pki/database.db")
		providerServer.Start()
	}
}
