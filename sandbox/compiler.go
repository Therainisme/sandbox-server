package sandbox

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/pkg/stdcopy"
)

const compilerContainerName = "sandbox-gcc-compiler"

func switchCompilerContainer() (containerId string) {
	ctx := context.Background()
	containers, err := cli.ContainerList(ctx, types.ContainerListOptions{All: true})
	if err != nil {
		panic(err)
	}

	for _, container := range containers {
		if compilerContainerName == (container.Names[0])[1:] {
			if container.State == "exited" {
				if err := cli.ContainerRemove(ctx, container.ID, types.ContainerRemoveOptions{}); err != nil {
					panic(err)
				}
				return ""
			} else {
				return container.ID
			}
		}
	}
	return ""
}

func runCompilerContainer() (containerId string) {
	ctx := context.Background()

	// run bash and hang up
	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image:      "gcc:9.4.0",
		WorkingDir: "/workspace",
		Cmd:        []string{"sleep", "infinity"},
	}, &container.HostConfig{
		Mounts: []mount.Mount{
			{
				Type:   mount.TypeBind,
				Source: GetHostWorkspace(),
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

func handleCompileTask(filename string, compilerContainerId string) CompileTaskResult {
	ctx := context.Background()
	resp, _ := cli.ContainerExecCreate(ctx, compilerContainerId, types.ExecConfig{
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
		Tty:          true,
		WorkingDir:   "/workspace",
		Cmd:          []string{"sh", "-c", fmt.Sprintf("g++ -O2 -fdiagnostics-color=never -std=c++11 -fmax-errors=3 -lm -o %s %s.cpp", filename, filename)},
	})

	response, err := cli.ContainerExecAttach(ctx, resp.ID, types.ExecStartCheck{})
	if err != nil {
		panic(err)
	}

	done := make(chan struct{})
	timeout, cancal := context.WithTimeout(ctx, 5*time.Second)
	var taskout, taskerr bytes.Buffer
	go func() {
		stdcopy.StdCopy(&taskout, &taskerr, response.Reader)
		close(done)
		cancal()
	}()

	select {
	case <-timeout.Done():
		taskerr.WriteString(ErrorCompilerTimeLimitExceededError.Error())
	case <-done:
	}
	response.Close()

	return CompileTaskResult{
		Stdout: taskout.String(),
		Stderr: taskerr.String(),
	}
}

func IsExistFile(filename string) bool {
	_, err := os.Stat(filename)
	if err != nil && os.IsNotExist(err) {
		return false
	}

	return true
}
