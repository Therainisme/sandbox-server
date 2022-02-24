package sandbox

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

var CurrentPath = ""

func Run(currentPath string, dispatch chan Parcel) {
	CurrentPath = currentPath

	CheckGCCImage()
	compilerContainerId := switchCompilerContainer()
	if compilerContainerId == "" {
		compilerContainerId = runCompilerContainer()
	}
	go handleCompileTask(compilerContainerId)
	go handleRunTask(compilerContainerId)

	for parcal := range dispatch {
		dispathResult := &DispatchResult{
			CResult: &CompileResult{},
			EResult: &ExecResult{},
		}

		compileMessage := make(chan taskResult)
		compileTask <- task{filename: parcal.Filename, res: compileMessage}
		compileResponse := <-compileMessage
		compileResult := &CompileResult{
			Msg:   compileResponse.out.String(),
			Error: compileResponse.err.String(),
		}

		if !IsExistFile(filepath.Join(CurrentPath, "workspace", parcal.Filename)) {
			if len(compileResult.Msg) > 0 {
				fmt.Println(compileResult.Msg)
			}
			if len(compileResult.Error) > 0 {
				fmt.Println(compileResult.Error)
			}
			dispathResult.CResult = compileResult
			parcal.Response <- dispathResult
			continue
		}

		execMessage := make(chan taskResult)
		execTask <- task{filename: parcal.Filename, res: execMessage}
		execResponse := <-execMessage

		execResult := NewExecResult(execResponse.out.Bytes())

		println(execResult.Memory)
		println(execResult.UseTime)
		println(execResult.Output)
		println(execResult.Error)

		dispathResult.EResult = execResult
		parcal.Response <- dispathResult
	}
}
