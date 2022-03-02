package sandbox

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/pkg/stdcopy"
)

func listenExecTaskList(compilerContainerId string) {
	for task := range execTaskList {
		go handleRunTask(task)
	}
}

func handleRunTask(task task) {
	ctx := context.Background()
	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image:        "therainisme/executor:1.0",
		Cmd:          []string{"sh", "-c", fmt.Sprintf("%srun -name %s", GetExecutorPath(), task.filename)},
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
		// StdinOnce:    true,
		OpenStdin: true,
		// Tty:          true,
	}, &container.HostConfig{
		Resources: container.Resources{
			Memory: 32 * 1024 * 1024,
		},
		Mounts: []mount.Mount{
			{
				Type:   mount.TypeBind,
				Source: GetHostWorkspace(),
				Target: GetExecutorWorkspacePath(),
			},
		},
	}, nil, nil, "run-"+task.filename)
	if err != nil {
		panic(err)
	}

	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		panic(err)
	}

	timeout, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	waiter, err := cli.ContainerAttach(ctx, resp.ID, types.ContainerAttachOptions{
		Stderr: true,
		Stdout: true,
		Stdin:  true,
		Stream: true,
		Logs:   true,
	})
	if err != nil {
		panic(err)
	}
	defer waiter.Close()
	// redirect stdin, stdout, stderr
	var taskout, taskerr bytes.Buffer
	readerDone := make(chan struct{})
	go func() {
		n, err := stdcopy.StdCopy(&taskout, &taskerr, waiter.Reader)
		if err != nil {
			log.Printf("read len: %d, stdcopy err: %s\n", n, err.Error())
		}
		readerDone <- struct{}{}
	}()
	go func() {
		n, err := waiter.Conn.Write([]byte(task.stdin + "\n"))
		if err != nil {
			log.Printf("write len: %d, waiter conn err: %s\n", n, err.Error())
		}
	}()

	statusCh, errCh := cli.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)

	select {
	case err := <-errCh:
		if err != nil {
			panic(err)
		}
	case <-timeout.Done():
		var execResult ExecResult
		execResult.Error = ErrorTimeLimitExceeded.Error()
		dataByte, _ := json.Marshal(&execResult)
		taskout.WriteString(string(dataByte))
	case status := <-statusCh:
		if status.StatusCode != 0 {
			log.Printf("docker container <id:%s> exit code: %d\n", resp.ID[:12], status.StatusCode)
			panic("sb")
		}
		if status.Error != nil {
			log.Printf("docker container <id:%s>exit err: %s\n", resp.ID[:12], status.Error.Message)
		}
	}

	task.result <- taskResult{taskout, taskerr}

	go func() {
		<-readerDone
		err = cli.ContainerRemove(ctx, resp.ID, types.ContainerRemoveOptions{
			Force: true,
		})
		if err != nil {
			panic(err)
		}
	}()
}
