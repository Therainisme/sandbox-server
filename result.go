package main

import (
	"encoding/json"
	"fmt"
	"log"
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

func NewExecResult(data []byte) *ExecResult {
	var r ExecResult
	err := json.Unmarshal(data, &r)
	if err != nil {
		log.Fatal(err.Error())
	}
	return &r
}

type CompileResult struct {
	Msg   string `json:"msg"`
	Error string `json:"error"`
}
