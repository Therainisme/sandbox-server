package main

import (
	"flag"
	"sandbox-server/sandbox"
)

var dispatch = make(chan sandbox.Task, 100)

func main() {
	flag.Parse()

	go sandbox.Run(dispatch)
	RunWebsocket(7777)
}
