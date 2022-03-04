package main

import (
	"bytes"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/tklauser/go-sysconf"
)

var name = flag.String("name", "", "program name")

var timeLimit = flag.Int64("time", 5000, "time limit(millisecond)")

var memoryLimit = flag.Int64("memory", 1024*64, "memory limit(kB)")

var clktck = getClktck()

func main() {
	flag.Parse()

	if *name == "" {
		log.Fatal("program name is empty")
	} else {
		output, rusage, err := execute(*name)
		var useTime int64 = 0
		var memory int64 = 0
		var errStr = ""
		if rusage != nil {
			// kill process haven't rusage
			useTime = rusage.Utime.Nano() + rusage.Stime.Nano()
			memory = rusage.Maxrss
		}
		if err != nil {
			errStr = err.Error()
		}

		result := ExecResult{
			Memory: memory,
			Time:   useTime,
			Output: output.String(),
			Error:  errStr,
		}
		result.print()
	}
}

func execute(name string) (*bytes.Buffer, *syscall.Rusage, error) {

	cmd := exec.Command("./workspace/" + name)
	var output bytes.Buffer
	cmd.Stdout = &output
	cmd.Stdin = os.Stdin

	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}

	Done := make(chan struct{})
	WaitFault := make(chan error)
	TLE := false

	// trace execute vmpeak
	// go func() {

	// 	for {
	// 		vmPeak, _ := strconv.ParseInt(getVmPeakByPid(cmd.Process.Pid), 10, 64)
	// 		if vmPeak > *memoryLimit {
	// 			cmd.Process.Kill()
	// 			MemoryLimitExceeded <- struct{}{}
	// 			break
	// 		}
	// 		println("vmPeak", vmPeak)
	// 		time.Sleep(time.Millisecond * 3000)
	// 	}
	// }()

	// trace execute time
	go func() {
		// ??????
		time.Sleep(time.Millisecond * time.Duration(*timeLimit) * 2)
		TLE = true
		cmd.Process.Kill()
	}()

	go func() {
		for {
			time.Sleep(time.Nanosecond * 1)
			useTime := getUseTime(cmd.Process.Pid)
			if useTime > *timeLimit {
				TLE = true
				cmd.Process.Kill()
				break
			}
		}
	}()

	// wait done
	go func() {
		if err := cmd.Wait(); err != nil && !TLE {
			WaitFault <- err
		} else {
			Done <- struct{}{}
		}
	}()

	select {
	case <-Done:
		rusage := cmd.ProcessState.SysUsage().(*syscall.Rusage)
		realTime := rusage.Utime.Nano() + rusage.Stime.Nano()
		// "realTime" convert to nanoseconds
		if TLE || realTime/1000000 > *timeLimit {
			return &output, rusage, ErrorTimeLimitExceeded
		}
		if rusage.Maxrss > *memoryLimit {
			return &output, rusage, ErrorMemoryLimitExceeded
		}
		return &output, rusage, nil

	case wf := <-WaitFault:
		rusage := cmd.ProcessState.SysUsage().(*syscall.Rusage)
		return &output, rusage, wf
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
	reg := regexp.MustCompile(`VmHWM:\s+\d+\s+kB`)
	numbReg := regexp.MustCompile(`\d+`)

	return numbReg.FindString(reg.FindString(res))
}

func getUseTime(pid int) int64 {
	f, err := os.Open("/proc/" + strconv.Itoa(pid) + "/stat")
	if err != nil {
		return -1
	}
	defer f.Close()

	data, err := ioutil.ReadAll(f)
	if err != nil {
		return -1
	}

	ss := strings.Split(string(data), " ")
	if len(ss) < 14 {
		return -1
	}

	utime, _ := strconv.ParseInt(ss[13], 10, 64)
	stime, _ := strconv.ParseInt(ss[14], 10, 64)

	return (utime + stime) * 1000 / clktck
}

func getClktck() int64 {
	tk, err := sysconf.Sysconf(sysconf.SC_CLK_TCK)
	if err != nil {
		panic(err)
	}
	return tk
}
