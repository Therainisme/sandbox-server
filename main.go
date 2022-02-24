package main

import "sandbox/sandbox"

var dispatch = make(chan sandbox.Parcel, 100)

func main() {
	go sandbox.Run(getCurrentAbPath(), dispatch)
	RunWebsocket(7777)
}
