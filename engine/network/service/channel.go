package service

import (
	"io"
	"sync"

	"github.com/xhaoh94/gox/app"
	"github.com/xhaoh94/gox/engine/codec"
	"github.com/xhaoh94/gox/engine/xlog"
)

type (
	//Channel 通信信道
	Channel struct {
		rfn        func([]byte)
		cfn        func()
		wfn        func([]byte)
		writeMutex sync.Mutex
		remoteAddr string
		localAddr  string

		Wg    sync.WaitGroup
		IsRun bool
	}
)

//RemoteAddr 获取连接地址
func (c *Channel) RemoteAddr() string {
	return c.remoteAddr
}

//LocalAddr 获取本地地址
func (c *Channel) LocalAddr() string {
	return c.localAddr
}

//Send 发送数据
func (c *Channel) Send(data []byte) {
	defer c.writeMutex.Unlock()
	c.writeMutex.Lock()
	if !c.IsRun {
		return
	}
	if app.SendMsgMaxLen > 0 { //分片发送
		DLen := len(data)
		pos := 0
		var endPos int
		for c.IsRun && pos < DLen {
			endPos = pos + app.SendMsgMaxLen
			if DLen < endPos {
				endPos = DLen
			}
			c.wfn(data[pos:endPos])
			pos = endPos
		}
	} else {
		c.wfn(data)
	}
}

//OnStop 停止信道
func (c *Channel) OnStop() {
	if c.cfn != nil {
		c.cfn()
	}
	c.localAddr = ""
	c.remoteAddr = ""
	c.rfn = nil
	c.wfn = nil
	c.cfn = nil
}

//SetCallBackFn 设置回调
func (c *Channel) SetCallBackFn(rfn func([]byte), cfn func()) {
	c.rfn = rfn
	c.cfn = cfn
}

//Init 初始化
func (c *Channel) Init(wfn func([]byte), remoteAddr string, localAddr string) {
	c.wfn = wfn
	c.remoteAddr = remoteAddr
	c.localAddr = localAddr
}

//Read
func (c *Channel) Read(r io.Reader, stopFunc func()) {

	// header, err := ioutil.ReadAll(io.LimitReader(r, 2))
	// if err != nil {
	// 	stopFunc() //超时断开链接
	// 	return
	// }
	// msgLen := util.Bytes2Uint16(header)
	// if int(msgLen) > consts.ReadMsgMaxLen {
	// 	xlog.Errorf("read msg size exceed local:[%s] remote:[%s]", c.LAddr, c.RAddr)
	// 	stopFunc() //超过读取最大限制
	// 	return
	// }
	// if msgLen == 0 {
	// 	xlog.Errorf("read msgLen=0 local:[%s] remote:[%s]", c.LAddr, c.RAddr)
	// 	stopFunc() //空数据
	// 	return
	// }
	// buf, err := ioutil.ReadAll(io.LimitReader(r, int64(msgLen)))
	// if err != nil {
	// 	stopFunc() //超时断开链接
	// 	return
	// }
	// if len(buf) < int(msgLen) {
	// 	return
	// }
	// c.session.OnRead(buf)

	header := make([]byte, 2)
	_, err := io.ReadFull(r, header)
	if err != nil {
		stopFunc() //超时断开链接
		return
	}
	msgLen := codec.BytesToUint16(header)
	if msgLen == 0 {
		xlog.Error("read msgLen=0 local:[%s] remote:[%s]", c.localAddr, c.remoteAddr)
		stopFunc() //空数据
		return
	}

	if int(msgLen) > app.ReadMsgMaxLen {
		xlog.Error("read msg size exceed local:[%s] remote:[%s]", c.localAddr, c.remoteAddr)
		stopFunc() //超过读取最大限制
		return
	}

	buf := make([]byte, msgLen)
	_, err = io.ReadFull(r, buf)

	if err != nil {
		stopFunc() //超时断开链接
		return
	}
	if c.rfn != nil {
		c.rfn(buf)
	}
}
