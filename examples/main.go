package main

import (
	"bytes"
	"log"
	"os/exec"
)

type (
	TT struct {
		Name string
		Age  int
	}
)

func main() {
	ExecuteCmd("run", "sv/main.go", "-appConf", "sv/app_1.yaml")
}

func ExecuteCmd(args ...string) {
	cmd := exec.Command("go", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout // 标准输出
	cmd.Stderr = &stderr // 标准错误
	err := cmd.Run()
	if len(stderr.Bytes()) > 0 {
		log.Printf("stderr:%s\n", string(stderr.Bytes()))
	}
	if len(stdout.Bytes()) > 0 {
		log.Printf("stdout:%s\n", string(stdout.Bytes()))
	}

	if err != nil {
		log.Fatalf("cmd.Run() failed with %s\n", err)
	}

}
