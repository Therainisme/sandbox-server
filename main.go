package main

import (
	"encoding/json"
	"flag"
	"sandbox-server/sandbox"
)

var dispatch = make(chan sandbox.Task, 100)

func main() {
	flag.Parse()

	go sandbox.Run(dispatch)
	// RunWebsocket(7777)
	localMain()
}

func localMain() {

	generatorName := "run"

	task := sandbox.Task{
		Filename: generatorName,
		Stdin:    "HelloWorld!",
		Result:   make(chan *sandbox.TaskResult),
	}
	dispatch <- task

	// wait for sandbox result
	res := <-task.Result
	bytes, _ := json.Marshal(res)

	println(string(bytes))
}
