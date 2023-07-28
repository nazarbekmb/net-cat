package main

import (
	"fmt"
	"os"

	"net-cat/server"
)

func main() {
	args := os.Args[1:]

	if len(args) == 0 {
		server.StartServer("8989")
	} else if len(args) == 1 {
		server.StartServer(args[0])
	} else {
		fmt.Println("[USAGE]: go run . $port <- (optional)")
		return
	}
}
