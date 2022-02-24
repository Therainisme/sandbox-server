package sandbox

import (
	"bytes"
	"context"
	"fmt"
	"path/filepath"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/pkg/stdcopy"
)

const ExecutorPath = "/com.therainisme/sandbox/executor/"

func handleRunTask(compilerContainerId string) {
	ctx := context.Background()
	for task := range execTask {
		resp, err := cli.ContainerCreate(ctx, &container.Config{
			Image:        "executor:v1",
			Cmd:          []string{"sh", "-c", fmt.Sprintf("%srun -name %s", ExecutorPath, task.filename)},
			AttachStdin:  true,
			AttachStdout: true,
			AttachStderr: true,
		}, &container.HostConfig{
			Resources: container.Resources{
				Memory:     64 * 1024 * 1024,
				MemorySwap: 64 * 1024 * 1024,
			},
			Mounts: []mount.Mount{
				{
					Type:   mount.TypeBind,
					Source: filepath.Join(CurrentPath, "workspace"),
					Target: filepath.Join(ExecutorPath, "workspace"),
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

		out, err := cli.ContainerLogs(ctx, resp.ID, types.ContainerLogsOptions{ShowStdout: true, ShowStderr: true})
		if err != nil {
			panic(err)
		}

		var taskout, taskerr bytes.Buffer
		stdcopy.StdCopy(&taskout, &taskerr, out)
		task.res <- taskResult{taskout, taskerr}
		out.Close()

		err = cli.ContainerRemove(ctx, resp.ID, types.ContainerRemoveOptions{})
		if err != nil {
			panic(err)
		}
	}
}
