package main

import (
	"bytes"
	"context"
	"path/filepath"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/pkg/stdcopy"
)

func handleRunTask(compilerContainerId string) {
	ctx := context.Background()
	for task := range runTask {
		resp, err := cli.ContainerCreate(ctx, &container.Config{
			Image:      "gcc",
			WorkingDir: "/workspace",
			Cmd:        []string{"timeout", "5", "sh", "-c", "./test" + task.filename},
		}, &container.HostConfig{
			Mounts: []mount.Mount{
				{
					Type:   mount.TypeBind,
					Source: filepath.Join(getCurrentAbPath(), "workspace"),
					Target: "/workspace",
				},
			},
		}, nil, nil, "run-"+task.filename)
		if err != nil {
			panic(err)
		}

		if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
			panic(err)
		}

		statusCh, errCh := cli.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)

		select {
		case err := <-errCh:
			if err != nil {
				panic(err)
			}
		case <-statusCh:
		}

		out, err := cli.ContainerLogs(ctx, resp.ID, types.ContainerLogsOptions{ShowStdout: true})
		if err != nil {
			panic(err)
		}

		var taskout, taskerr bytes.Buffer
		stdcopy.StdCopy(&taskout, &taskerr, out)
		task.res <- result{taskout, taskerr}
		out.Close()

		err = cli.ContainerRemove(ctx, resp.ID, types.ContainerRemoveOptions{})
		if err != nil {
			panic(err)
		}

	}
}
