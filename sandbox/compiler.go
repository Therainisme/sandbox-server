package sandbox

import (
	"bytes"
	"context"
	"fmt"
	"os"
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

func switchCompilerContainer() (containerId string) {
	ctx := context.Background()
	containers, err := cli.ContainerList(ctx, types.ContainerListOptions{})
	if err != nil {
		panic(err)
	}

	for _, container := range containers {
		if compilerContainerName == (container.Names[0])[1:] {
			return container.ID
		}
	}
	return ""
}

func runCompilerContainer() (containerId string) {
	ctx := context.Background()

	// run bash and hang up
	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image:      "gcc",
		WorkingDir: "/workspace",
		Tty:        true,
	}, &container.HostConfig{
		Mounts: []mount.Mount{
			{
				Type:   mount.TypeBind,
				Source: filepath.Join(mainPath, "workspace"),
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

func listenCompileTaskList(compilerContainerId string) {
	for task := range compileTaskList {
		go handleCompileTask(task, compilerContainerId)
	}
}

func handleCompileTask(task task, compilerContainerId string) {
	ctx := context.Background()
	resp, _ := cli.ContainerExecCreate(ctx, compilerContainerId, types.ExecConfig{
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
		Tty:          true,
		WorkingDir:   "/workspace",
		Cmd:          []string{"timeout", "5", "sh", "-c", fmt.Sprintf("g++ -O2 -fdiagnostics-color=never -std=c++11 -fmax-errors=3 -lm -o %s %s.cpp", task.filename, task.filename)},
	})

	response, err := cli.ContainerExecAttach(context.Background(), resp.ID, types.ExecStartCheck{})
	if err != nil {
		panic(err)
	}

	var taskout, taskerr bytes.Buffer
	stdcopy.StdCopy(&taskout, &taskerr, response.Reader)
	task.result <- taskResult{taskout, taskerr}

	response.Close()
}

func IsExistFile(filename string) bool {
	_, err := os.Stat(filename)
	if err != nil && os.IsNotExist(err) {
		return false
	}

	return true
}
