package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/quic-go/quic-go"
)

func quic_connect_upstream(node_id string, address string) {
	for {
		quic_connect_upstream_do(node_id, address)
		fmt.Printf("连接上级节点失败，1秒后重试: %s\n", address)
		time.Sleep(time.Second * 3)
	}
}

func quic_connect_upstream_do(remote_node_id string, address string) {
	conn, err := quic.DialAddr(context.Background(), address, GetClientTLSConfig(), GetQuicClientConfig())
	if err != nil {
		fmt.Printf("连接上级节点失败: %v\n", err)
		return
	}
	defer delete_quic_client_by_id(remote_node_id)

	// 建立消息通道
	stream, err := conn.OpenStreamSync(context.Background())
	if err != nil {
		fmt.Printf("打开stream失败: %v\n", err)
		return
	}
	quic_send_syn_msg(stream, remote_node_id)

	if verify_syn_ack(stream, MSG_TYPE_SYN_ACK_MSG) == false {
		fmt.Printf("⚠️  消息通道: 没有收到synack\n")
		os.Exit(1)
		return
	}

	save_quic_stream(remote_node_id, conn, stream, LINK_TYPE_MSG)

	// 处理数据通道
	go handleQuicConnection_remote(conn)

	//处理msg
	for {
		msg := QuicMessageFromStream(stream)
		if msg == nil {
			fmt.Printf("⚠️  消息通道收到空消息\n")
			return
		}
		switch msg.Type {
		case MSG_TYPE_PING:
			// 创建一个pong消息
			msgpong := NewQuicMessage(MSG_TYPE_PONG, remote_node_id, PongMessage{})
			stream.Write(msgpong.ToBuffer())
		default:
			fmt.Printf("⚠️  消息通道收到未知消息: %v\n", msg)
		}
	}
}

func handleQuicConnection_remote(conn quic.Connection) {
	for {
		stream, err := conn.AcceptStream(context.Background())
		errorcount := 0
		if err != nil {
			fmt.Printf("⚠️  接受流失败: %v\n", err)
			errorcount++
			if errorcount > 5 {
				fmt.Printf("⚠️  接受流失败次数过多，退出\n")
				// 删除这个连接
				delete_quic_client(conn)
				return
			}
			continue
		}
		errorcount = 0
		go handleQuicStream(conn, stream)
	}
}
