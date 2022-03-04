package sandbox

import "errors"

var (
	ErrorTimeLimitExceeded              = errors.New("time limit exceeded")
	ErrorMemoryLimitExceeded            = errors.New("memory limit exceeded")
	ErrorSegmentationFault              = errors.New("segmentation fault")
	ErrorCompilerError                  = errors.New("compiler error")
	ErrorCompilerTimeLimitExceededError = errors.New("compiler time limit exceeded")
)

const (
	CompileErrorType = "compile error"
	ExecErrorType    = "exec error"
)
