package main

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
)

func tcp_proxy_main(local_port int, target_node_id string, target_address string) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", local_port))
	if err != nil {
		fmt.Printf("监听本地端口失败: %v\n", err)
		return
	}
	defer listener.Close()
	fmt.Printf("监听本地端口: %d\n", local_port)
	errorcount := 0
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("接受连接失败: %v\n", err)
			errorcount++
			if errorcount > 5 {
				fmt.Printf("⚠️  接受连接失败次数过多，退出\n")
				os.Exit(1)
			}
			continue
		}
		errorcount = 0
		go tcp_proxy_handle(conn, target_node_id, target_address)
	}
}

func tcp_proxy_handle(conn net.Conn, target_node_id string, target_address string) {
	if len(fm.quic_client) != 1 {
		fmt.Printf("⚠️  跳跃节点数量不正确: %d\n", len(fm.quic_client))
		conn.Close()
		return
	}
	// fm.quic_client里面应该只有一个跳跃节点，找出那个，但是id应该不是target_node_id
	proxynodeid := ""
	var proxyquic *quic_client
	for node_id, quic_client := range fm.quic_client {
		proxynodeid = node_id
		proxyquic = quic_client
		break
	}

	if proxyquic == nil {
		fmt.Printf("⚠️  没有找到目标节点: %s\n", target_node_id)
		conn.Close()
		return
	}

	// 建立数据通道
	stream, err := proxyquic.conn.OpenStreamSync(context.Background())
	if err != nil {
		fmt.Printf("打开stream失败: %v\n", err)
		conn.Close()
		return
	}

	// 数据通道握手
	ok := send_syn_data(proxynodeid, stream, target_node_id, target_address)
	if !ok {
		conn.Close()
		return
	}
	// 进行io.copy
	ch := make(chan bool)
	go func() {
		io.Copy(conn, stream)
		ch <- true
	}()
	go func() {
		io.Copy(stream, conn)
		ch <- true
	}()
	<-ch
	stream.Close()
	conn.Close()
}
