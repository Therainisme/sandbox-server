package sandbox

import (
	"flag"
	"path/filepath"
)

var workspace = flag.String("workspace", "", "workspace path")

func GetHostWorkspace() string {
	return *workspace
}

func GetRelativeWorkspace() string {
	return filepath.Join("./", "workspace")
}

func GetExecutorPath() string {
	return "/sandbox-server/executor/"
}

func GetExecutorWorkspacePath() string {
	return filepath.Join(GetExecutorPath(), "workspace")
}
