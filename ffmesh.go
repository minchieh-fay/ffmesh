package main

import (
	"github.com/quic-go/quic-go"
)

// 流信息结构
type streaminfo struct {
	stream    quic.Stream
	link_type int
}

// QUIC客户端连接信息
type quic_client struct {
	node_id    string
	conn       quic.Connection
	streaminfo []streaminfo
	is_up      bool
	version    int
}

// FFMesh主结构体
type ffmesh struct {
	config      *Config
	quic_client map[string]*quic_client
}

// 创建新的FFMesh实例
func new_ffmesh() *ffmesh {
	return &ffmesh{
		quic_client: make(map[string]*quic_client),
	}
}

func (f *ffmesh) start() {

}
