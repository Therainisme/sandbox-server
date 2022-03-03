package sandbox

// export Task
type Task struct {
	Filename string
	Stdin    string
	Result   chan *TaskResult
}

type TaskResult struct {
	Memory    int64  `json:"memory"`
	Time      int64  `json:"time"`
	Output    string `json:"output"`
	Error     string `json:"error"`
	ErrorType string `json:"errorType"`
}

type CompileTaskResult struct {
	Stdout string
	Stderr string
}

type ExecTask struct {
	Filename string
	Stdin    string
}

type ExecTaskResult struct {
	Memory int64  `json:"memory"`
	Time   int64  `json:"time"`
	Output string `json:"output"`
	Error  string `json:"error"`
}
