package service

type (
	Tag int
	//SessionTag 会话标签
	SessionTag struct {
		tag Tag
	}
)

const (
	//TagConnector 连接者
	TagConnector Tag = 1
	//TagAccept 接收者
	TagAccept Tag = 2
)

//GetTag 获取标签
func (st *SessionTag) GetTag() Tag {
	return st.tag
}

//IsConnector 是否是连接者
func (st *SessionTag) IsConnector() bool {
	return st.tag == TagConnector
}

//GetTagName 获取名字
func (st *SessionTag) GetTagName() string {
	switch st.tag {
	case TagConnector:
		return "connector"
	case TagAccept:
		return "accept"
	default:
		return ""
	}
}
