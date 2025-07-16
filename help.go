package main

import "github.com/quic-go/quic-go"

func delete_quic_client(conn quic.Connection) {
	for node_id, streaminfo := range fm.quic_client {
		if streaminfo.conn == conn {
			for _, streaminfo := range streaminfo.streaminfo {
				if streaminfo.stream != nil {
					streaminfo.stream.Close()
				}
			}
			if fm.quic_client[node_id].conn != nil {
				fm.quic_client[node_id].conn.CloseWithError(quic.ApplicationErrorCode(quic.NoError), "accept stream error")
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
func save_quic_stream(node_id string, conn quic.Connection, stream quic.Stream, link_type int) {
	if fm.quic_client[node_id] == nil {
		fm.quic_client[node_id] = &quic_client{
			node_id:    node_id,
			conn:       conn,
			streaminfo: []streaminfo{},
		}
	}
	fm.quic_client[node_id].streaminfo = append(fm.quic_client[node_id].streaminfo, streaminfo{stream: stream, link_type: LINK_TYPE_MSG})
}

func quic_send_syn_msg(stream quic.SendStream, node_id string) {
	msgsyn := NewQuicMessage(MSG_TYPE_SYN_MSG, SynMsgMessage{NodeID: fm.config.NodeID})
	msgsyn.NodeId = node_id
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
