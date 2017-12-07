package main

import (
	"fmt"
	node "anonymous-messaging/node"
	client "anonymous-messaging/client"
)


func main() {
	fmt.Println("Hello world")
	m := node.Mix{"Id", "google.com", "127.0.0.1", 45, 23}
	fmt.Println(m)
	c := client.Client{}
	fmt.Println(c)
	fmt.Println(c.EncodeMessage("Message"))
	fmt.Println(c.DecodeMessage("EnMessage"))


	fmt.Println(m.ProcessPacket("HelloPacket"))
	m.SendLoopMessage()
	m.LogInfo("Log message")

	p := new (node.Provider)
	p.Id = "ProviderId"
	p.Host = "ProviderAddr"
	p.PubKey = 45
	fmt.Println(p.ProcessPacket("packet for provider"))
	p.StorePacket("packet for storing")

}
