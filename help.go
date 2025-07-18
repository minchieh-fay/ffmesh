package main

import (
	"fmt"
	"time"

	"github.com/quic-go/quic-go"
)

func delete_quic_client(conn quic.Connection) {
	for node_id, streaminfo := range fm.quic_client {
		if streaminfo.conn == conn {
			if fm.quic_client[node_id].conn != nil {
				fm.quic_client[node_id].conn.CloseWithError(quic.ApplicationErrorCode(quic.NoError), "accept stream error")
			}
			for _, streaminfo := range streaminfo.streaminfo {
				if streaminfo.stream != nil {
					streaminfo.stream.Close()
				}
			}
			delete(fm.quic_client, node_id)
			break
		}
	}
}

// 如果id+conn 发生了改变，删除原有conn
func delete_quic_client_when_conn_change(node_id string, conn quic.Connection) {
	client := fm.quic_client[node_id]
	if client == nil {
		return
	}
	if client.conn != conn {
		client.conn.CloseWithError(quic.ApplicationErrorCode(quic.NoError), "already connected")
		for _, streaminfo := range client.streaminfo {
			streaminfo.stream.Close()
		}
		delete(fm.quic_client, node_id)
	}
}

// 存入新的quic的stream
func save_quic_stream(node_id string, conn quic.Connection, stream quic.Stream, is_up bool, version int) {
	if fm.quic_client[node_id] == nil {
		fm.quic_client[node_id] = &quic_client{
			node_id:    node_id,
			conn:       conn,
			streaminfo: []streaminfo{},
			is_up:      is_up,
			version:    version,
		}
	}
	fm.quic_client[node_id].streaminfo = append(fm.quic_client[node_id].streaminfo, streaminfo{stream: stream, link_type: LINK_TYPE_MSG})
}

func quic_send_syn_msg(stream quic.SendStream, node_id string) {
	isup := false
	if fm.config.IsQuicEnabled() {
		isup = true
	}
	msgsyn := NewQuicMessage(MSG_TYPE_SYN_MSG, node_id, SynMsgMessage{Version: VERSION, NodeID: fm.config.NodeID, IsUp: isup})
	stream.Write(msgsyn.ToBuffer())
}

// 如果id+conn 发生了改变，删除原有conn
func delete_quic_client_by_id(node_id string) {
	client := fm.quic_client[node_id]
	if client == nil {
		return
	}
	if client.conn != nil {
		client.conn.CloseWithError(quic.ApplicationErrorCode(quic.NoError), "already connected")
	}
	for _, streaminfo := range client.streaminfo {
		streaminfo.stream.Close()
	}
	delete(fm.quic_client, node_id)
}

// 验证syn ack
func get_syn_ack(stream quic.Stream, msgtype int) *SynAckMsgMessage {
	if stream == nil {
		fmt.Printf("stream为空\n")
		return nil
	}
	msgack := QuicMessageFromStream(stream)
	if msgack == nil {
		fmt.Printf("msgack为空\n")
		return nil
	}
	if msgack.Type != msgtype {
		fmt.Printf("msgack类型不匹配\n")
		return nil
	}
	synack := msgack.Data.(*SynAckMsgMessage)
	if synack.Result == false {
		fmt.Printf("synack失败: %s\n", synack.Reason)
		return nil
	}
	return synack
}

func send_syn_data(remote_node_id string, stream quic.Stream, target_node_id string, target_address string) bool {
	ok := false

	go func() {
		time.Sleep(time.Second * 3)
		if !ok {
			stream.Close()
		}
	}()

	// 发送syn
	msgsyn := NewQuicMessage(MSG_TYPE_SYN_DATA, remote_node_id, SynDataMessage{NodeID: fm.config.NodeID, TargetID: target_node_id, TargetTcpAddr: target_address})
	stream.Write(msgsyn.ToBuffer())

	// 接收synack
	msgack := QuicMessageFromStream(stream)
	if msgack == nil || msgack.Type != MSG_TYPE_SYN_ACK_DATA {
		fmt.Printf("接收synack失败\n")
		return false
	}
	ok = true
	return true
}
