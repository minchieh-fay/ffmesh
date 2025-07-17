package main

import (
	"fmt"
	"time"

	"github.com/quic-go/quic-go"
)

func timer_main() {
	go timer_ping_quic()
}

func timer_ping_quic() {
	// 遍历client列表，发送ping消息
	for {
		time.Sleep(time.Second * 5)
		for remote_node_id, client := range fm.quic_client {
			if client == nil {
				continue
			}
			for _, streaminfo := range client.streaminfo {
				if streaminfo.link_type == LINK_TYPE_MSG {
					quic_send_ping_msg(streaminfo.stream, remote_node_id)
				}
			}
		}
	}
}

func quic_send_ping_msg(stream quic.SendStream, remote_node_id string) {
	if stream == nil {
		return
	}
	msgping := NewQuicMessage(MSG_TYPE_PING, remote_node_id, PingMessage{})
	_, err := stream.Write(msgping.ToBuffer())
	if err != nil {
		fmt.Printf("⚠️  发送ping消息失败: %v\n", err)
		// msg通道不通，等价于节点已下线
		delete_quic_client_by_id(remote_node_id)
	}
}
