package main

import (
	"context"

	"github.com/xhaoh94/gox/engine/network/protoreg"
	"github.com/xhaoh94/gox/engine/types"
	"github.com/xhaoh94/gox/examples/netpack"
	"github.com/xhaoh94/gox/examples/pb"
)

func main() {
	// ExecuteCmd("run", "sv/main.go", "-appConf", "app_1.yaml")
	protoreg.Register(1002, Test)
	protoreg.Register(1000, Test1)
}
func Test(ctx context.Context, session types.ISession, msg *pb.A) {
	session.Send(100, &pb.B{Id: "test", Etype: 1, Position: &pb.Vector3{X: 0, Y: 1, Z: 2}})
}
func Test1(ctx context.Context, session types.ISession, req any) {
	session.Send(100, &pb.B{Id: "test", Etype: 1, Position: &pb.Vector3{X: 0, Y: 1, Z: 2}})
}
func RspToken(ctx context.Context, req *netpack.G2L_Login) *netpack.L2G_Login {
	return &netpack.L2G_Login{}
}

// func ExecuteCmd(args ...string) {
// 	cmd := exec.Command("go", args...)
// 	var stdout, stderr bytes.Buffer
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
