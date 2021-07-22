package consts

type (
	SessionTag int
)

const (
	//Connector 连接者
	Connector SessionTag = 1
	//Accept 接收者
	Accept SessionTag = 2
)

const (
	//Linux Linux
	Linux string = "linux"
	//Windows Windows
	Windows string = "windows"
)
