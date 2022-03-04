package sandbox

import (
	"log"

	"github.com/docker/docker/client"
)

var cli *client.Client
var compilerContainerId = ""

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

	compilerContainerId = switchCompilerContainer()
	if compilerContainerId == "" {
		compilerContainerId = runCompilerContainer()
	}
	go listenSandboxTaskList(TaskList)
}

func listenSandboxTaskList(taskList chan Task) {
	for task := range taskList {
		go handleSandboxTask(task)
	}
}

func handleSandboxTask(parcal Task) {

	// try to compile

	// wait for compile result
	compileResult, ok := handleCompileTask(parcal.Filename, compilerContainerId)
	if !ok {
		log.Println(compileResult.Stdout)
		parcal.Result <- &TaskResult{
			Error:     compileResult.Stdout,
			ErrorType: CompileErrorType,
		}
		return
	}

	// try to exec
	execResult := handleRunTask(ExecTask{
		Filename: parcal.Filename,
		Stdin:    parcal.Stdin,
	})

	// wait for exec result

	log.Printf("--------------------------------\n")
	log.Printf("memory: %d\n", execResult.Memory)
	log.Printf("time: %d\n", execResult.Time)
	log.Printf("output: %s\n", execResult.Output)
	log.Printf("error: %s\n", execResult.Error)
	log.Printf("================================\n")

	taskResult := &TaskResult{
		Memory:    execResult.Memory,
		Time:      execResult.Time,
		Output:    execResult.Output,
		Error:     execResult.Error,
		ErrorType: "",
	}
	if execResult.Error != "" {
		taskResult.ErrorType = ExecErrorType
	}
	parcal.Result <- taskResult
}
