package service

import (
	"io"
	"sync"

	"github.com/xhaoh94/gox"
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
func (channel *Channel) RemoteAddr() string {
	return channel.remoteAddr
}

// LocalAddr 获取本地地址
func (channel *Channel) LocalAddr() string {
	return channel.localAddr
}

// Send 发送数据
func (channel *Channel) Send(data []byte) {
	defer channel.writeMutex.Unlock()
	channel.writeMutex.Lock()
	if !channel.IsRun {
		return
	}
	sendMax := gox.AppConf.Network.SendMsgMaxLen
	if sendMax > 0 { //分片发送
		DLen := len(data)
		pos := 0
		var endPos int
		for channel.IsRun && pos < DLen {
			endPos = pos + sendMax
			if DLen < endPos {
				endPos = DLen
			}
			channel.wfn(data[pos:endPos])
			pos = endPos
		}
	} else {
		channel.wfn(data)
	}
}

// OnStop 停止信道
func (channel *Channel) OnStop() {
	if channel.Session != nil {
		channel.Session.release()
		channel.Session = nil
	}
	channel.localAddr = ""
	channel.remoteAddr = ""
	channel.wfn = nil
}

func (channel *Channel) SetSession(session types.ISession) {
	channel.Session = session.(*Session)
}

// Init 初始化
func (channel *Channel) Init(wfn func([]byte), remoteAddr string, localAddr string) {
	channel.wfn = wfn
	channel.remoteAddr = remoteAddr
	channel.localAddr = localAddr
}

// Read
func (channel *Channel) Read(r io.Reader) bool {

	// header, err := ioutil.ReadAll(io.LimitReader(r, 2))
	// if err != nil {
	// 	stopFunc() //超时断开链接
	// 	return
	// }
	// msgLen := util.Bytes2Uint16(header)
	// if int(msgLen) > consts.ReadMsgMaxLen {
	// 	xlog.Errorf("read msg size exceed local:[%s] remote:[%s]", channel.LAddr, channel.RAddr)
	// 	stopFunc() //超过读取最大限制
	// 	return
	// }
	// if msgLen == 0 {
	// 	xlog.Errorf("read msgLen=0 local:[%s] remote:[%s]", channel.LAddr, channel.RAddr)
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
	// channel.session.OnRead(buf)
	if channel.Session != nil {
		return channel.Session.parseReader(r)
	}
	return true
}
