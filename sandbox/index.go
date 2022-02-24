package sandbox

import (
	"fmt"
	"path/filepath"

	"github.com/docker/docker/client"
)

var cli *client.Client
var compileTaskList = make(chan task, 100)
var execTaskList = make(chan task, 100)
var mainPath = ""

func init() {
	client, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}
	cli = client
}

func Run(currentPath string, TaskList chan Task) {
	mainPath = currentPath

	CheckGCCImage()
	compilerContainerId := switchCompilerContainer()
	if compilerContainerId == "" {
		compilerContainerId = runCompilerContainer()
	}

	go listenCompileTaskList(compilerContainerId)
	go listenExecTaskList(compilerContainerId)

	go listenSandboxTaskList(TaskList)
}

func listenSandboxTaskList(taskList chan Task) {
	for task := range taskList {
		go handleSandboxTask(task)
	}
}

func handleSandboxTask(parcal Task) {
	dispathResult := &TaskResult{
		CResult: &CompileResult{},
		EResult: &ExecResult{},
	}

	// try to compile
	compileTask := task{
		filename: parcal.Filename,
		result:   make(chan taskResult),
	}
	compileTaskList <- compileTask

	// wait for compile result
	compileResponse := <-compileTask.result
	compileResult := &CompileResult{
		Msg:   compileResponse.out.String(),
		Error: compileResponse.err.String(),
	}

	if !IsExistFile(filepath.Join(mainPath, "workspace", parcal.Filename)) {
		if len(compileResult.Msg) > 0 {
			fmt.Println(compileResult.Msg)
		}
		if len(compileResult.Error) > 0 {
			fmt.Println(compileResult.Error)
		}
		dispathResult.CResult = compileResult
		parcal.Result <- dispathResult
		return
	}

	// try to exec
	execTask := task{
		filename: parcal.Filename,
		result:   make(chan taskResult),
	}
	execTaskList <- execTask

	// wait for exec result
	execResponse := <-execTask.result
	execResult := NewExecResult(execResponse.out.Bytes())

	println(execResult.Memory)
	println(execResult.UseTime)
	println(execResult.Output)
	println(execResult.Error)

	dispathResult.EResult = execResult
	parcal.Result <- dispathResult
}
