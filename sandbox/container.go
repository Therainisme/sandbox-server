package sandbox

import (
	"bytes"
	"context"
	"fmt"

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
			Memory:     64 * 1024 * 1024,
			MemorySwap: 64 * 1024 * 1024,
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

	waiter, err := cli.ContainerAttach(ctx, resp.ID, types.ContainerAttachOptions{
		Stderr: true,
		Stdout: true,
		Stdin:  true,
		Stream: true,
	})
	if err != nil {
		panic(err)
	}
	defer waiter.Close()

	// redirect stdin, stdout, stderr
	var taskout, taskerr bytes.Buffer
	go stdcopy.StdCopy(&taskout, &taskerr, waiter.Reader)
	waiter.Conn.Write([]byte(task.stdin + "\n"))

	statusCh, errCh := cli.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)

	select {
	case err := <-errCh:
		if err != nil {
			panic(err)
		}
	case <-statusCh:
	}

	task.result <- taskResult{taskout, taskerr}

	err = cli.ContainerRemove(ctx, resp.ID, types.ContainerRemoveOptions{})
	if err != nil {
		panic(err)
	}
}
