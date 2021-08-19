package main

import (
	"fmt"
	"reflect"
	"runtime"
	"time"
)

func main() {
	On(test1)
	fmt.Print("\n")
	On(func() {})
	fmt.Print("\n")
	On(func() {})
	time.Sleep(time.Second * 2)
}
func test1() {

}
func test2() {

}

//On 监听事件 回调不可带参数
func On(task interface{}) {
	f := reflect.ValueOf(task)
	fmt.Print(runtime.FuncForPC(f.Pointer()).Name())
	fmt.Print("\n")
	fmt.Print(f.Pointer())

}
