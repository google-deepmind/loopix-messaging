package main

import (
	svr "anonymous-messaging/server"
	"fmt"
	"os"
)

func f(from string) {
	for i := 0; i < 3; i++ {
		fmt.Println(from, ":", i)
	}
}

func main() {

	args := os.Args[1:]
	if args == nil {
		fmt.Println("No arguments")
	} else {
		host := args[0]
		port := args[1]
		mix := svr.NewMixServer("Mix1", host, port, 0, 0)
		mix.Run()
	}
}