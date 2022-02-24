package sandbox

import "bytes"

type task struct {
	filename string
	res      chan taskResult
}

type taskResult struct {
	out bytes.Buffer
	err bytes.Buffer
}
