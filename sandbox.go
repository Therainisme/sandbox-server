package main

import (
	"fmt"
	"path/filepath"

	"github.com/docker/docker/client"
)

var cli *client.Client
var compileTask = make(chan task, 100)
var execTask = make(chan task, 100)

func init() {
	client, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}
	cli = client
}

func main() {
	CheckGCCImage()
	compilerContainerId := switchCompilerContainer()
	if compilerContainerId == "" {
		compilerContainerId = runCompilerContainer()
	}
	go handleCompileTask(compilerContainerId)
	go handleRunTask(compilerContainerId)

	for {
		var filename string
		fmt.Scanln(&filename)
		compileMessage := make(chan result)
		compileTask <- task{filename: filename, res: compileMessage}
		compileResponse := <-compileMessage
		compileResult := &CompileResult{
			Msg:   compileResponse.out.String(),
			Error: compileResponse.err.String(),
		}

		if !IsExistFile(filepath.Join(getCurrentAbPath(), "workspace", filename)) {
			if len(compileResult.Msg) > 0 {
				fmt.Println(compileResult.Msg)
			}
			if len(compileResult.Error) > 0 {
				fmt.Println(compileResult.Error)
			}
			continue
		}

		execMessage := make(chan result)
		execTask <- task{filename: filename, res: execMessage}
		execResponse := <-execMessage

		execResult := NewExecResult(execResponse.out.Bytes())
		println(execResult.Memory)
		println(execResult.UseTime)
		println(execResult.Output)
		println(execResult.Error)
	}
}
