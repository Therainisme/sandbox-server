package sandbox

import "bytes"

type task struct {
	filename string
	res      chan result
}

type result struct {
	out bytes.Buffer
	err bytes.Buffer
}
