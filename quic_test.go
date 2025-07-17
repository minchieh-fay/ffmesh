package main

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/quic-go/quic-go"
)

// 测试用的配置
var testConfig = &Config{
	NodeID: "test-server-001",
	Quic: QuicConfig{
		ListenPort: 3333,
		Upstreams: []UpstreamConfig{
			{
				Name:    "test-upstream",
				NodeID:  "test-client-001",
				Address: "127.0.0.1:3333",
			},
		},
	},
	Proxies: []ProxyConfig{
		{
			Name:          "test-proxy",
			LocalPort:     8080,
			TargetNodeID:  "test-client-001",
			TargetAddress: "127.0.0.1:80",
		},
	},
}

func TestQuicLocalConnection(t *testing.T) {
	fmt.Println("=== QUIC本地连接测试 ===")

	// 初始化ffmesh
	fm = new_ffmesh()
	fm.config = testConfig

	// 启动本地QUIC监听器
	go func() {
		fmt.Println("启动本地QUIC监听器...")
		quic_local_main()
	}()

	// 等待监听器启动
	time.Sleep(2 * time.Second)

	// 测试连接本地节点
	fmt.Println("测试连接本地节点...")
	quic_connect_upstream_do("test-server-001", "127.0.0.1:3333")

	// 等待一段时间观察结果
	time.Sleep(5 * time.Second)
}

func TestQuicDirectConnection(t *testing.T) {
	fmt.Println("=== QUIC直接连接测试 ===")

	// 启动服务端
	go func() {
		listener, err := quic.ListenAddr("127.0.0.1:3334", GetServerTLSConfig(), GetQuicServerConfig())
		if err != nil {
			t.Errorf("服务端启动失败: %v", err)
			return
		}
		defer listener.Close()
		fmt.Println("✅ 服务端启动成功，监听 127.0.0.1:3334")

		conn, err := listener.Accept(context.Background())
		if err != nil {
			t.Errorf("服务端接受连接失败: %v", err)
			return
		}
		fmt.Printf("✅ 服务端接受连接: %s\n", conn.RemoteAddr())
		defer conn.CloseWithError(0, "test complete")
	}()

	// 等待服务端启动
	time.Sleep(1 * time.Second)

	// 启动客户端
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	fmt.Println("尝试连接 127.0.0.1:3334...")
	conn, err := quic.DialAddr(ctx, "127.0.0.1:3334", GetClientTLSConfig(), GetQuicClientConfig())
	if err != nil {
		t.Errorf("❌ 客户端连接失败: %v", err)
		return
	}
	defer conn.CloseWithError(0, "test complete")

	fmt.Printf("✅ 客户端连接成功: %s\n", conn.RemoteAddr())
	fmt.Println("✅ QUIC直接连接测试通过！")
}

func TestQuicConfig(t *testing.T) {
	fmt.Println("=== QUIC配置测试 ===")

	// 测试TLS配置
	tlsConfig := GetServerTLSConfig()
	if tlsConfig == nil {
		t.Error("服务端TLS配置为空")
	}
	fmt.Println("✅ 服务端TLS配置正常")

	clientTLSConfig := GetClientTLSConfig()
	if clientTLSConfig == nil {
		t.Error("客户端TLS配置为空")
	}
	fmt.Println("✅ 客户端TLS配置正常")

	// 测试QUIC配置
	quicServerConfig := GetQuicServerConfig()
	fmt.Printf("服务端QUIC配置: %+v\n", quicServerConfig)

	quicClientConfig := GetQuicClientConfig()
	fmt.Printf("客户端QUIC配置: %+v\n", quicClientConfig)

	fmt.Println("✅ QUIC配置测试完成")
}

func TestQuicMessageProtocol(t *testing.T) {
	fmt.Println("=== QUIC消息协议测试 ===")

	// 测试消息创建
	synMsg := NewQuicMessage(MSG_TYPE_SYN_MSG, "test-node", SynMsgMessage{
		NodeID: "test-node",
	})
	if synMsg == nil {
		t.Error("创建SYN消息失败")
	}
	fmt.Printf("✅ 创建SYN消息成功: %+v\n", synMsg)

	// 测试消息序列化
	buffer := synMsg.ToBuffer()
	if len(buffer) == 0 {
		t.Error("消息序列化失败")
	}
	fmt.Printf("✅ 消息序列化成功，长度: %d\n", len(buffer))

	// 测试消息类型
	if synMsg.Type != MSG_TYPE_SYN_MSG {
		t.Errorf("消息类型错误，期望: %d, 实际: %d", MSG_TYPE_SYN_MSG, synMsg.Type)
	}
	fmt.Println("✅ 消息类型正确")

	fmt.Println("✅ QUIC消息协议测试完成")
}
