package main

import (
	"fmt"
	"reflect"
)

type (
	TT struct {
		Name string
		Age  int
	}
)

func main() {
	// ExecuteCmd("run", "sv/main.go", "-appConf", "app_1.yaml")
	t := &TT{}
	tt(t)
	fmt.Printf("%v", t)
}
func tt(t any) {
	newT := &TT{}
	newT.Name = "ccc"
	newT.Age = 18
	v1 := reflect.ValueOf(t).Elem()
	v2 := reflect.ValueOf(newT).Elem()
	for i := 0; i < v2.NumField(); i++ {
		fieldInfo := v2.Type().Field(i)
		v1.FieldByName(fieldInfo.Name).Set(v2.Field(i))
	}
}

// func ExecuteCmd(args ...string) {
// 	cmd := exec.Command("go", args...)
// 	var stdout, stderr bytes.Buffreflect
// 	cmd.Stdout = &stdout // 标准输出
// 	cmd.Stderr = &stderr // 标准错误
// 	err := cmd.Run()
// 	if len(stderr.Bytes()) > 0 {
// 		log.Printf("stderr:%s\n", string(stderr.Bytes()))
// 	}
// 	if len(stdout.Bytes()) > 0 {
// 		log.Printf("stdout:%s\n", string(stdout.Bytes()))
// 	}

// 	if err != nil {
// 		log.Fatalf("cmd.Run() failed with %s\n", err)
// 	}

// }
