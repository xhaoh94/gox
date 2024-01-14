package game

import (
	"github.com/xhaoh94/gox/engine/helper/strhelper"
	"github.com/xhaoh94/gox/engine/network/codec"
	"github.com/xhaoh94/gox/engine/network/protoreg"
)

const (
	//Gate 网关服务
	Gate string = "gate"
	//Login 登录服务
	Login string = "login"
	//Scene 场景服务
	Scene string = "scene"
)

var (
	InteriorRelay uint32
)

func init() {
	InteriorRelay = strhelper.StringToHash("InteriorRelay")
	protoreg.BindCodec(InteriorRelay, codec.MsgPack)
}

type Interior_Relay struct {
	Roles   []uint32
	CMD     uint32
	Require []byte
}
