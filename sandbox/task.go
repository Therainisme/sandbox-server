package sandbox

import "bytes"

type task struct {
	filename string
	result   chan taskResult
}

type taskResult struct {
	out bytes.Buffer
	err bytes.Buffer
}
