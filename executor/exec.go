package main

import (
	"bytes"
	"flag"
	"log"
	"os/exec"
	"regexp"
	"strconv"
	"syscall"
	"time"
)

var name = flag.String("name", "", "program name")

var timeLimit = flag.Int64("time", 5000, "time limit(millisecond)")

var memoryLimit = flag.Int64("memory", 1024*64, "memory limit(kB)")

func main() {
	flag.Parse()

	if *name == "" {
		log.Fatal("program name is empty")
	} else {
		output, rusage, err := execute(*name)
		result := ExecResult{
			Memory:  rusage.Maxrss,
			UseTime: rusage.Utime.Nano() + rusage.Stime.Nano(),
			Output:  output.String(),
			Error:   err,
		}
		result.print()
	}
}

func execute(name string) (*bytes.Buffer, *syscall.Rusage, error) {

	cmd := exec.Command("./workspace/" + name)
	var output bytes.Buffer
	cmd.Stdout = &output

	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}

	Done := make(chan struct{})
	MemoryLimitExceeded := make(chan struct{})
	TimeLimitExceeded := make(chan struct{})
	SegmentationFault := make(chan struct{})

	// trace execute vmpeak
	go func() {

		for {
			vmPeak, _ := strconv.ParseInt(getVmPeakByPid(cmd.Process.Pid), 10, 64)
			if vmPeak > *memoryLimit {
				MemoryLimitExceeded <- struct{}{}
				break
			}
			time.Sleep(time.Millisecond * 300)
		}
	}()

	// trace execute time
	go func() {
		// ??????
		for {
			time.Sleep(time.Millisecond * time.Duration(*timeLimit))
			cmd.Process.Kill()
			TimeLimitExceeded <- struct{}{}
		}
	}()

	// wait done
	go func() {
		if err := cmd.Wait(); err != nil {
			SegmentationFault <- struct{}{}
		} else {
			Done <- struct{}{}
		}
	}()

	select {
	case <-Done:
		rusage := cmd.ProcessState.SysUsage().(*syscall.Rusage)
		realTime := rusage.Utime.Nano() + rusage.Stime.Nano()
		// "timeLimit" convert to nanoseconds
		if realTime > *timeLimit*1000000 {
			return &output, nil, ErrorTimeLimitExceeded
		}
		return &output, rusage, nil

	case <-MemoryLimitExceeded:
		return &output, nil, ErrorMemoryLimitExceeded

	case <-TimeLimitExceeded:
		return &output, nil, ErrorTimeLimitExceeded

	case <-SegmentationFault:
		return &output, nil, ErrorSegmentationFault
	}
}

// return kB
func getVmPeakByPid(id int) string {
	cmd := exec.Command("cat", "/proc/"+strconv.Itoa(id)+"/status")
	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}

	res := out.String()
	reg := regexp.MustCompile(`VmPeak:\s+\d+\s+kB`)
	numbReg := regexp.MustCompile(`\d+`)

	return numbReg.FindString(reg.FindString(res))
}
