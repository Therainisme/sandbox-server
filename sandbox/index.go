package sandbox

import (
	"log"
	"path/filepath"
	"strings"

	"github.com/docker/docker/client"
)

var cli *client.Client
var compileTaskList = make(chan task, 100)
var execTaskList = make(chan task, 100)

func init() {
	client, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}
	cli = client
}

func Run(TaskList chan Task) {
	if *workspace == "" {
		panic("workspace is empty")
	}

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
		Msg:   strings.ReplaceAll(compileResponse.out.String(), parcal.Filename, ""),
		Error: strings.ReplaceAll(compileResponse.err.String(), parcal.Filename, ""),
	}

	if !IsExistFile(filepath.Join(GetRelativeWorkspace(), parcal.Filename)) {
		if len(compileResult.Msg) > 0 {
			log.Printf("compile msg: %s\n", compileResult.Msg)
		}
		if len(compileResult.Error) > 0 {
			log.Printf("compile err: %s\n", compileResult.Error)
		}
		dispathResult.CResult = compileResult
		parcal.Result <- dispathResult
		return
	}

	// try to exec
	execTask := task{
		filename: parcal.Filename,
		stdin:    parcal.Stdin,
		result:   make(chan taskResult),
	}
	execTaskList <- execTask

	// wait for exec result
	execResponse := <-execTask.result

	execResult := NewExecResult(execResponse.out.Bytes())

	log.Printf("memory: %d\n", execResult.Memory)
	log.Printf("time: %d\n", execResult.UseTime)
	log.Printf("output: %s\n", execResult.Output)
	log.Printf("error: %s\n", execResult.Error)

	dispathResult.EResult = execResult
	parcal.Result <- dispathResult
}
