package sandbox

import (
	"bytes"
	"context"
	"log"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/pkg/stdcopy"
)

func handleRunTask(task ExecTask) ExecTaskResult {
	ctx := context.Background()
	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image:        "ubuntu:20.04",
		Cmd:          []string{"sh", "-c", "/workspace/" + task.Filename},
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
		OpenStdin:    true,
		WorkingDir:   "/workspace/",
	}, &container.HostConfig{
		Resources: container.Resources{
			Memory: 96 * 1024 * 1024,
		},
		Mounts: []mount.Mount{
			{
				Type:   mount.TypeBind,
				Source: GetHostWorkspace(),
				Target: "/workspace/",
			},
		},
	}, nil, nil, "run-"+task.Filename)
	if err != nil {
		panic(err)
	}

	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		panic(err)
	}

	containerJSON, err := cli.ContainerInspect(ctx, resp.ID)
	if err != nil {
		panic(nil)
	}
	println(containerJSON.State.Pid)

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
		n, err := waiter.Conn.Write([]byte(task.Stdin + "\n"))
		if err != nil {
			log.Printf("write len: %d, waiter conn err: %s\n", n, err.Error())
		}
	}()

	statusCh, errCh := cli.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)

	var remove = true
	select {
	case err := <-errCh:
		if err != nil {
			panic(err)
		}
	case <-timeout.Done():
		panic("no handle")
	case status := <-statusCh:
		if status.StatusCode != 0 {
			log.Printf("docker container <id:%s> exit code: %d\n", resp.ID[:12], status.StatusCode)
			remove = false
		}
		if status.Error != nil {
			log.Printf("docker container <id:%s>exit err: %s\n", resp.ID[:12], status.Error.Message)
		}
	}

	go func() {
		<-readerDone
		if remove {
			err = cli.ContainerRemove(ctx, resp.ID, types.ContainerRemoveOptions{
				Force: true,
			})
			if err != nil {
				panic(err)
			}
		}
	}()

	return ExecTaskResult{
		Memory: 0,
		Time:   0,
		Output: taskout.String(),
		Error:  taskerr.String(),
	}
}
