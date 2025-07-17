package main

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"time"

	"github.com/quic-go/quic-go"
)

func quic_local_main() {
	if !fm.config.IsQuicEnabled() {
		fmt.Printf("⚠️  本地QUIC监听器未启用 (未配置监听端口)\n")
		return
	}
	addr := fmt.Sprintf("0.0.0.0:%d", fm.config.Quic.ListenPort)
	listener, err := quic.ListenAddr(addr, GetServerTLSConfig(), GetQuicServerConfig())
	if err != nil {
		fmt.Printf("⚠️  启动QUIC监听器失败: %v\n", err)
		os.Exit(1)
		return
	}
	for {
		conn, err := listener.Accept(context.Background())
		if err != nil {
			fmt.Printf("⚠️  接受连接失败: %v\n", err)
			time.Sleep(1 * time.Second)
			continue
		}
		go handleQuicConnection(conn)
	}
}

func handleQuicConnection(conn quic.Connection) {
	if conn == nil {
		return
	}
	for {
		stream, err := conn.AcceptStream(context.Background())
		errorcount := 0
		if err != nil {
			fmt.Printf("⚠️  接受流失败: %v\n", err)
			errorcount++
			if errorcount > 5 {
				fmt.Printf("⚠️  接受流失败次数过多，退出\n")
				// 删除这个连接
				conn.CloseWithError(quic.ApplicationErrorCode(quic.NoError), "accept stream error")
				delete_quic_client(conn)
				return
			}
			continue
		}
		errorcount = 0
		go handleQuicStream(conn, stream)
	}
}

func handleQuicStream(conn quic.Connection, stream quic.Stream) {
	msgsyn := QuicMessageFromStream(stream)
	fmt.Printf("收到消息: %v\n", msgsyn)
	if msgsyn == nil {
		fmt.Printf("⚠️  收到空握手消息\n")
		stream.Close()
		conn.CloseWithError(quic.ApplicationErrorCode(quic.NoError), "empty syn message")
		delete_quic_client(conn)
		return
	}
	if msgsyn.Type != MSG_TYPE_SYN_MSG && msgsyn.Type != MSG_TYPE_SYN_DATA {
		fmt.Printf("⚠️  收到非握手消息: %v\n", msgsyn)
		stream.Close()
		conn.CloseWithError(quic.ApplicationErrorCode(quic.NoError), "empty syn message")
		delete_quic_client(conn)
		return
	}

	if msgsyn.Type == MSG_TYPE_SYN_MSG {
		go handleQuicStream_msg(msgsyn, conn, stream)
	} else if msgsyn.Type == MSG_TYPE_SYN_DATA {
		go handleQuicStream_data(msgsyn, conn, stream)
	}
}

func handleQuicStream_msg(msgsyn *QuicMessage, conn quic.Connection, stream quic.Stream) {
	synmsg := msgsyn.Data.(*SynMsgMessage)
	remote_node_id := synmsg.NodeID

	// 如果id+conn 发生了改变，删除原有conn
	delete_quic_client_when_conn_change(remote_node_id, conn)

	// 回复ack
	msgsynack := NewQuicMessage(MSG_TYPE_SYN_ACK_MSG, remote_node_id, SynAckMsgMessage{})
	stream.Write(msgsynack.ToBuffer())

	// 保存新stream信息
	save_quic_stream(remote_node_id, conn, stream, LINK_TYPE_MSG)

	defer func() {
		// msg通道关闭  等价于 conn关闭
		stream.Close()
		conn.CloseWithError(quic.ApplicationErrorCode(quic.NoError), "already connected")
		delete_quic_client(conn)
	}()

	// 处理消息
	go func() {
		for {
			msg := QuicMessageFromStream(stream)
			if msg == nil {
				fmt.Printf("⚠️  消息通道收到空消息\n")
				return
			}
			if msg.Type == MSG_TYPE_PING {
				// 创建一个pong消息
				msgpong := NewQuicMessage(MSG_TYPE_PONG, remote_node_id, PongMessage{})
				stream.Write(msgpong.ToBuffer())
			} else {
				fmt.Printf("⚠️  消息通道收到未知消息: %v\n", msg)
			}
		}
	}()
}

func handleQuicStream_data(msgsyn *QuicMessage, conn quic.Connection, stream quic.Stream) {
	synmsg := msgsyn.Data.(*SynDataMessage)
	target_id := synmsg.TargetID
	remote_node_id := synmsg.NodeID
	target_tcp_addr := synmsg.TargetTcpAddr

	// 应该先去看看targetid是不是自己，或者能不能在client列表中找到
	if target_id == fm.config.NodeID {
		fmt.Printf("⚠️  数据通道：目标节点是自己: %s\n", target_id)
		handleQuicStream_data_target_self(remote_node_id, stream, target_tcp_addr)
		return
	}
	// 转发
	handleQuicStream_data_target_other(remote_node_id, stream, target_id, target_tcp_addr)
}

func handleQuicStream_data_target_self(remote_node_id string, stream quic.Stream, tcptarget string) {
	defer stream.Close()
	tcpconn, err := net.Dial("tcp", tcptarget)
	if err != nil {
		fmt.Printf("⚠️  连接目标地址失败: %v\n", err)
		return
	}
	defer tcpconn.Close()

	// 回复ack
	msgsynack := NewQuicMessage(MSG_TYPE_SYN_ACK_DATA, remote_node_id, SynAckMsgMessage{})
	stream.Write(msgsynack.ToBuffer())

	ch := make(chan struct{}, 1)
	go func() {
		io.Copy(tcpconn, stream)
		ch <- struct{}{}
	}()
	go func() {
		io.Copy(stream, tcpconn)
		ch <- struct{}{}
	}()
	<-ch
}

func handleQuicStream_data_target_other(src_node_id string, srcstream quic.Stream, target_id string, target_tcp_addr string) {
	defer srcstream.Close()
	// 查找我有木有目标节点信息，决定我是否可以帮源请求转发
	dstclient, ok := fm.quic_client[target_id]
	if !ok {
		fmt.Printf("⚠️  数据通道：目标节点不存在: %s\n", target_id)
		return
	}

	// 先和client握手 数据通道，如果通了，再答复stream ack

	// 创建steram
	if dstclient == nil || dstclient.conn == nil {
		fmt.Printf("⚠️  数据通道：目标节点连接不存在: %s\n", target_id)
		return
	}
	// 创建dststream
	dststream, err := dstclient.conn.OpenStreamSync(context.Background())
	if err != nil {
		fmt.Printf("⚠️  创建数据通道失败: %v\n", err)
		return
	}
	defer dststream.Close()

	// 发送syn消息
	msgsyn := NewQuicMessage(MSG_TYPE_SYN_DATA, target_id, SynDataMessage{
		NodeID:        fm.config.NodeID,
		TargetID:      target_id,
		TargetTcpAddr: target_tcp_addr,
	})
	dststream.Write(msgsyn.ToBuffer())

	// 等待dst的ack
	msgsynack := QuicMessageFromStream(dststream)
	if msgsynack == nil {
		fmt.Printf("⚠️  数据通道：收到空ack消息\n")
		return
	}
	if msgsynack.Type != MSG_TYPE_SYN_ACK_DATA {
		fmt.Printf("⚠️  数据通道：收到非ack消息: %v\n", msgsynack)
		return
	}

	// dst数据通道通了，向src回复ack
	msgsynack_src := NewQuicMessage(MSG_TYPE_SYN_ACK_DATA, src_node_id, SynAckMsgMessage{})
	srcstream.Write(msgsynack_src.ToBuffer())

	// 进行数据拷贝
	ch := make(chan struct{}, 1)
	go func() {
		io.Copy(dststream, srcstream)
		ch <- struct{}{}
	}()
	go func() {
		io.Copy(srcstream, dststream)
		ch <- struct{}{}
	}()
	<-ch
}
