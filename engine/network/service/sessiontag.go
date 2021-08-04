package service

import "github.com/xhaoh94/gox/consts"

type (
	//SessionTag 会话标签
	SessionTag struct {
		tag consts.SessionTag
	}
)

//GetTag 获取标签
func (st *SessionTag) GetTag() consts.SessionTag {
	return st.tag
}

//IsConnector 是否是连接者
func (st *SessionTag) IsConnector() bool {
	return st.tag == consts.Connector
}

//GetTagName 获取名字
func (st *SessionTag) GetTagName() string {
	switch st.tag {
	case consts.Connector:
		return "connector"
	case consts.Accept:
		return "accept"
	default:
		return ""
	}
}
