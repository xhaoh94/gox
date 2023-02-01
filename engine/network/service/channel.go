package service

import (
	"io"
	"sync"

	"github.com/xhaoh94/gox/engine/types"
)

type (
	//Channel 通信信道
	Channel struct {
		Session *Session
		// rfn        func([]byte)
		// cfn        func()
		wfn        func([]byte)
		writeMutex sync.Mutex
		remoteAddr string
		localAddr  string
		// endian     binary.ByteOrder

		Wg    sync.WaitGroup
		IsRun bool
	}
)

// RemoteAddr 获取连接地址
func (c *Channel) RemoteAddr() string {
	return c.remoteAddr
}

// LocalAddr 获取本地地址
func (c *Channel) LocalAddr() string {
	return c.localAddr
}

// Send 发送数据
func (c *Channel) Send(data []byte) {
	defer c.writeMutex.Unlock()
	c.writeMutex.Lock()
	if !c.IsRun {
		return
	}
	sendMax := c.Session.AppConf().Network.SendMsgMaxLen
	if sendMax > 0 { //分片发送
		DLen := len(data)
		pos := 0
		var endPos int
		for c.IsRun && pos < DLen {
			endPos = pos + sendMax
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

// OnStop 停止信道
func (c *Channel) OnStop() {
	if c.Session != nil {
		c.Session.release()
		c.Session = nil
	}
	c.localAddr = ""
	c.remoteAddr = ""
	c.wfn = nil
}

func (c *Channel) SetSession(session types.ISession) {
	c.Session = session.(*Session)
}

// Init 初始化
func (c *Channel) Init(wfn func([]byte), remoteAddr string, localAddr string) {
	c.wfn = wfn
	c.remoteAddr = remoteAddr
	c.localAddr = localAddr
}

// Read
func (c *Channel) Read(r io.Reader) bool {

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
	if c.Session != nil {
		return c.Session.parseReader(r)
	}
	return true
}
