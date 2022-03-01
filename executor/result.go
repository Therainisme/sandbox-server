package main

import (
	"encoding/json"
	"fmt"
)

type ExecResult struct {
	Memory  int64  `json:"memory"`
	UseTime int64  `json:"time"`
	Output  string `json:"output"`
	Error   string `json:"error"`
}

func (r *ExecResult) print() {
	result, _ := json.Marshal(r)
	fmt.Println(string(result))
}
