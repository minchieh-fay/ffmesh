package main

import (
	"fmt"
	"log"
	"os"
	"runtime"
)

var fm *ffmesh = new_ffmesh()

func main() {
	// 设置单线程
	runtime.GOMAXPROCS(1)
	// 启动定时器
	timer_main()
	// 检查命令行参数
	configFile := "config.yaml"
	if len(os.Args) > 1 {
		configFile = os.Args[1]
	}

	fmt.Printf("=== FFMesh 启动 ===\n")
	fmt.Printf("使用配置文件: %s\n\n", configFile)

	// 加载配置文件
	config, err := LoadConfig(configFile)
	if err != nil {
		log.Fatal("加载配置失败:", err)
	}
	fm.config = config

	// 打印配置信息
	config.PrintConfig()

	// 启动FFMesh节点
	fmt.Printf("开始启动 FFMesh 节点: %s\n", config.NodeID)

	// 1. 启动本地QUIC监听器（如果配置了）
	go quic_local_main()

	// 2. 连接上级节点（独立于本地监听）
	if len(config.Quic.Upstreams) > 0 {
		fmt.Printf("📡 连接上级节点:\n")
		for _, upstream := range config.Quic.Upstreams {
			fmt.Printf("   - %s (%s) -> %s\n", upstream.Name, upstream.NodeID, upstream.Address)
			// TODO: 实际连接上级节点
			go quic_connect_upstream(upstream.NodeID, upstream.Address)
		}
	} else {
		fmt.Printf("⚠️  无上级节点配置\n")
	}

	// 3. 启动代理服务（独立功能）
	if len(config.Proxies) > 0 {
		fmt.Printf("\n🔧 启动代理服务:\n")
		for _, proxy := range config.Proxies {
			fmt.Printf("   - %s: 本地端口 %d -> 节点 %s (%s)\n",
				proxy.Name, proxy.LocalPort, proxy.TargetNodeID, proxy.TargetAddress)
			// TODO: 实际启动代理监听器
			go tcp_proxy_main(proxy.LocalPort, proxy.TargetNodeID, proxy.TargetAddress)
		}
	} else {
		fmt.Printf("\n⚠️  无代理配置\n")
	}

	fmt.Printf("\n🚀 FFMesh 节点启动完成\n")
	select {}
}
