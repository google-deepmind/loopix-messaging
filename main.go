package main

import (
	"anonymous-messaging/server"
)

func main() {

	mixServer := server.NewMixServer("Mix","localhost", "9999", 0, 0)
	mixServer.Start()
}

