package sandbox

import "bytes"

type task struct {
	filename string
	stdin    string
	result   chan taskResult
}

type taskResult struct {
	out bytes.Buffer
	err bytes.Buffer
}
