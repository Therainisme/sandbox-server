package main

import (
	"encoding/json"
	"fmt"
)

type ExecResult struct {
	Memory  int64  `json:"memory"`
	UseTime int64  `josn:"time"`
	Output  string `json:"output"`
	Error   error  `json:"error"`
}

func (r *ExecResult) print() {
	result, _ := json.Marshal(r)
	fmt.Println(string(result))
}