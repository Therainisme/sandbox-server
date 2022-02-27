package main

import (
	"sandbox/sandbox"
)

var dispatch = make(chan sandbox.Task, 100)

func main() {
	go sandbox.Run(getCurrentAbPath(), dispatch)
	RunWebsocket(7777)
}
