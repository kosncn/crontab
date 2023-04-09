package common

import (
	"encoding/json"
)

// Response 通讯响应
type Response struct {
	State   string      `json:"state"`   // 状态
	Message string      `json:"message"` // 消息
	Data    interface{} `json:"data"`    // 数据
}

// NewResponse 实例化通讯响应对象
func NewResponse() *Response {
	return &Response{}
}

// Build 构建通讯响应数据
func (r *Response) Build(state string, message string, data interface{}) ([]byte, error) {
	// 通讯响应对象赋值
	r.State = state
	r.Message = message
	r.Data = data

	// 序列化通讯响应对象并返回
	return json.Marshal(r)
}
