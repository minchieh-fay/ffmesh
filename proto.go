package main

import (
	"encoding/binary"
	"encoding/json"

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
		return nil
	}
	var msg QuicMessage
	err = json.Unmarshal(buf, &msg)
	if err != nil {
		return nil
	}
	if msg.NodeId != fm.config.NodeID {
		return nil
	}
	return &msg
}

const (
	LINK_TYPE_MSG  = 0 // 消息通道
	LINK_TYPE_DATA = 1 // 数据通道
)

// 握手-并告知这是一条控制信令通道
type SynMsgMessage struct {
	NodeID string `json:"node_id"` // 发送方节点ID
}

type SynAckMsgMessage struct {
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
