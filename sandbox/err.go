package sandbox

import "errors"

var (
	ErrorTimeLimitExceeded   = errors.New("time limit exceeded")
	ErrorMemoryLimitExceeded = errors.New("memory limit exceeded")
	ErrorSegmentationFault   = errors.New("segmentation fault")
)
