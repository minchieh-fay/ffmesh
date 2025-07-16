package main

import "github.com/quic-go/quic-go"

type streaminfo struct {
	stream    quic.Stream
	link_type int
}

type quic_client struct {
	node_id    string
	conn       quic.Connection
	streaminfo []streaminfo
}

type ffmesh struct {
	config      *Config
	quic_client map[string]*quic_client
}

func new_ffmesh() *ffmesh {
	return &ffmesh{}
}

func (f *ffmesh) start() {

}
