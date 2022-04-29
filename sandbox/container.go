package sandbox

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/pkg/stdcopy"
)

var executorToken = make(chan struct{}, 10)

func handleRunTask(task ExecTask) ExecTaskResult {
	executorToken <- struct{}{}
	defer func() {
		<-executorToken
	}()

	ctx := context.Background()
	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image: "therainisme/executor:2.0",
		Cmd: []string{
			"./run", "-name", task.Filename,
		},
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
		OpenStdin:    true,
	}, &container.HostConfig{
		Resources: container.Resources{
			Memory: 96 * 1024 * 1024,
		},
		Mounts: []mount.Mount{
			{
				Type:   mount.TypeBind,
				Source: GetHostWorkspace(),
				Target: GetExecutorWorkspacePath(),
			},
		},
	}, nil, nil, "run-"+task.Filename)
	if err != nil {
		panic(err)
	}

	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		panic(err)
	}

	timeout, timeoutCancel := context.WithTimeout(ctx, 6*time.Second)
	defer timeoutCancel()

	waiter, err := cli.ContainerAttach(timeout, resp.ID, types.ContainerAttachOptions{
		Stdin:  true,
		Stream: true,
	})
	if err != nil {
		panic(err)
	}
	defer waiter.Close()

	go func() {
		n, err := waiter.Conn.Write([]byte(task.Stdin + "\n"))
		if err != nil {
			log.Printf("write len: %d, waiter conn err: %s\n", n, err.Error())
		}
		waiter.CloseWrite()
	}()

	statusCh, errCh := cli.ContainerWait(timeout, resp.ID, container.WaitConditionNotRunning)

	var remove = true
	var result ExecTaskResult
	select {

	case <-timeout.Done():
		result.Error = ErrorTimeLimitExceeded.Error()

	case err := <-errCh:
		if err != nil {
			panic(err)
		}

	case status := <-statusCh:
		if status.StatusCode != 0 {
			log.Printf("docker container <id:%s> exit code: %d\n", resp.ID[:12], status.StatusCode)
			remove = false
		}
		if status.Error != nil {
			log.Printf("docker container <id:%s>exit err: %s\n", resp.ID[:12], status.Error.Message)
		}

		var taskout, taskerr bytes.Buffer
		out, err := cli.ContainerLogs(timeout, resp.ID, types.ContainerLogsOptions{ShowStdout: true, Follow: true})
		if err != nil {
			panic(err)
		}
		stdcopy.StdCopy(&taskout, &taskerr, out)

		err = json.Unmarshal(taskout.Bytes(), &result)
		if err != nil {
			panic(err)
		}
	}

	if remove {
		err = cli.ContainerRemove(ctx, resp.ID, types.ContainerRemoveOptions{
			Force: true,
		})
		if err != nil {
			panic(err)
		}
	}

	return result
}
