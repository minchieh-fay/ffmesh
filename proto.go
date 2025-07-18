package main

import (
	"encoding/binary"
	"encoding/json"
	"fmt"

	"github.com/quic-go/quic-go"
)

// 这里会定义一些结构体。用于传输数据，json格式

// 消息类型常量
const (
	MSG_TYPE_SYN_MSG      = 0  // 握手-并告知这是一条控制信令通道
	MSG_TYPE_SYN_ACK_MSG  = 1  // 握手回复-并告知这是一条控制信令通道
	MSG_TYPE_SYN_DATA     = 2  // 握手-并告知这是一条数据通道
	MSG_TYPE_SYN_ACK_DATA = 3  // 握手回复-并告知这是一条数据通道
	MSG_TYPE_PING         = 4  // ping消息
	MSG_TYPE_PONG         = 5  // pong回复
	MSG_TYPE_PROXY_DATA   = 6  // 代理数据
	MSG_TYPE_NODE_INFO    = 5  // 节点信息
	MSG_TYPE_ROUTE_UPDATE = 5  // 路由更新
	MSG_TYPE_ERROR        = 99 // 错误消息
)

// 基础消息结构
type QuicMessage struct {
	Type   int         `json:"type"`    // 消息类型
	NodeId string      `json:"node_id"` // 接收端节点ID
	Data   interface{} `json:"data"`    // 消息数据
}

func NewQuicMessage(typ int, remote_node_id string, data interface{}) *QuicMessage {
	return &QuicMessage{
		Type:   typ,
		NodeId: remote_node_id,
		Data:   data,
	}
}

func (m *QuicMessage) ToBuffer() []byte {
	buf, err := json.Marshal(m)
	if err != nil {
		return nil
	}
	// 前面2个字节是len
	lenBuf := make([]byte, 2)
	binary.BigEndian.PutUint16(lenBuf, uint16(len(buf)))
	buf = append(lenBuf, buf...)
	return buf
}

func QuicMessageFromStream(stream quic.Stream) *QuicMessage {
	buf := make([]byte, 2)
	_, err := stream.Read(buf)
	if err != nil {
		return nil
	}
	len := binary.BigEndian.Uint16(buf)
	buf = make([]byte, len)
	_, err = stream.Read(buf)
	if err != nil {
		fmt.Printf("读取消息长度失败: %v\n", err)
		return nil
	}
	var msg QuicMessage
	err = json.Unmarshal(buf, &msg)
	if err != nil {
		fmt.Printf("解析消息失败: %v\n", err)
		return nil
	}
	if msg.NodeId != fm.config.NodeID {
		fmt.Printf("节点ID不匹配，丢弃消息: %s\n", msg.NodeId)
		return nil
	}

	// 根据消息类型反序列化Data字段
	switch msg.Type {
	case MSG_TYPE_SYN_MSG:
		var synMsg SynMsgMessage
		dataBytes, _ := json.Marshal(msg.Data)
		json.Unmarshal(dataBytes, &synMsg)
		msg.Data = &synMsg
	case MSG_TYPE_SYN_DATA:
		var synData SynDataMessage
		dataBytes, _ := json.Marshal(msg.Data)
		json.Unmarshal(dataBytes, &synData)
		msg.Data = &synData
	case MSG_TYPE_SYN_ACK_MSG:
		var synAckMsg SynAckMsgMessage
		dataBytes, _ := json.Marshal(msg.Data)
		json.Unmarshal(dataBytes, &synAckMsg)
		msg.Data = &synAckMsg
	case MSG_TYPE_SYN_ACK_DATA:
		var synAckData SynAckDataMessage
		dataBytes, _ := json.Marshal(msg.Data)
		json.Unmarshal(dataBytes, &synAckData)
		msg.Data = &synAckData
	case MSG_TYPE_PING:
		var pingMsg PingMessage
		dataBytes, _ := json.Marshal(msg.Data)
		json.Unmarshal(dataBytes, &pingMsg)
		msg.Data = &pingMsg
	case MSG_TYPE_PONG:
		var pongMsg PongMessage
		dataBytes, _ := json.Marshal(msg.Data)
		json.Unmarshal(dataBytes, &pongMsg)
		msg.Data = &pongMsg
	}

	return &msg
}

const (
	LINK_TYPE_MSG  = 0 // 消息通道
	LINK_TYPE_DATA = 1 // 数据通道
)

const (
	VERSION = 1
)

// 握手-并告知这是一条控制信令通道
type SynMsgMessage struct {
	Version int    `json:"version"` // 版本号
	NodeID  string `json:"node_id"` // 发送方节点ID
	IsUp    bool   `json:"is_up"`   // 是否是上级节点
}

type SynAckMsgMessage struct {
	IsUp    bool   `json:"is_up"`   // 是否是上级节点
	Result  bool   `json:"result"`  // 是否成功
	Reason  string `json:"reason"`  // 原因
	Version int    `json:"version"` // 版本号
}

// 握手-并告知这是一条数据通道
type SynDataMessage struct {
	NodeID        string `json:"node_id"`         // 发送方节点ID
	TargetID      string `json:"target_id"`       // 能帮我传输数据的目标节点ID
	TargetTcpAddr string `json:"target_tcp_addr"` // 目标tcp地址
}

type SynAckDataMessage struct {
}

// Ping消息结构
type PingMessage struct {
}

// Pong回复结构
type PongMessage struct {
}

// 代理数据结构
type ProxyMessage struct {
	ProxyID    string `json:"proxy_id"`    // 代理连接ID
	TargetNode string `json:"target_node"` // 目标节点ID
	TargetAddr string `json:"target_addr"` // 目标地址
}
