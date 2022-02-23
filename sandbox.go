package main

import (
	"fmt"

	"github.com/docker/docker/client"
)

var cli *client.Client
var compileTask = make(chan task, 100)
var runTask = make(chan task, 100)

func init() {
	client, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}
	cli = client
}

func main() {
	CheckGCCImage()
	stopCompilerContainer()
	compilerContainerId := runCompilerContainer()
	go handleCompileTask(compilerContainerId)
	go handleRunTask(compilerContainerId)

	for {
		var filename string
		fmt.Scanln(&filename)
		compileMessage := make(chan result)
		compileTask <- task{filename: filename, res: compileMessage}
		compileResult := <-compileMessage
		println("compile message:\n" + compileResult.out.String())

		runMessage := make(chan result)
		runTask <- task{filename: filename, res: runMessage}
		runResult := <-runMessage
		println("run message:\n" + runResult.out.String())
	}
}
