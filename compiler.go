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

const compilerContainerName = "gcc-compiler"

func stopCompilerContainer() {
	ctx := context.Background()
	containers, err := cli.ContainerList(ctx, types.ContainerListOptions{})
	if err != nil {
		panic(err)
	}

	for _, container := range containers {
		if compilerContainerName == (container.Names[0])[1:] {
			if err := cli.ContainerStop(ctx, container.ID, nil); err != nil {
				panic(err)
			}
			if err := cli.ContainerRemove(ctx, container.ID, types.ContainerRemoveOptions{}); err != nil {
				panic(err)
			}
		}
	}
}

func runCompilerContainer() (containerId string) {
	ctx := context.Background()

	// todo pull image
	// run bash and hang up
	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image:      "gcc",
		WorkingDir: "/workspace",
		Tty:        true,
	}, &container.HostConfig{
		Mounts: []mount.Mount{
			{
				Type:   mount.TypeBind,
				Source: filepath.Join(getCurrentAbPath(), "workspace"),
				Target: "/workspace",
			},
		},
	}, nil, nil, compilerContainerName)
	if err != nil {
		panic(err)
	}

	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		panic(err)
	}

	return resp.ID
}

func handleCompileTask(compilerContainerId string) {
	ctx := context.Background()
	for task := range compileTask {
		resp, _ := cli.ContainerExecCreate(ctx, compilerContainerId, types.ExecConfig{
			AttachStdin:  true,
			AttachStdout: true,
			AttachStderr: true,
			Tty:          true,
			WorkingDir:   "/workspace",
			Cmd:          []string{"timeout", "5", "sh", "-c", "g++ -o test" + task.filename + " " + task.filename + ".cpp"},
		})

		response, err := cli.ContainerExecAttach(context.Background(), resp.ID, types.ExecStartCheck{})
		if err != nil {
			panic(err)
		}

		var taskout, taskerr bytes.Buffer
		stdcopy.StdCopy(&taskout, &taskerr, response.Reader)
		task.res <- result{taskout, taskerr}

		response.Close()
	}
}