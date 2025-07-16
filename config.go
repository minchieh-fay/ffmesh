package main

import (
	"crypto/rand"
	"fmt"
	"io/ioutil"
	"log"
	"time"

	"gopkg.in/yaml.v3"
)

// 代理配置结构
type ProxyConfig struct {
	LocalPort     int    `yaml:"local_port"`
	TargetNodeID  string `yaml:"target_node_id"`
	TargetAddress string `yaml:"target_address"`
	Name          string `yaml:"name"`
}

// 上级节点配置结构
type UpstreamConfig struct {
	NodeID  string `yaml:"node_id"`
	Address string `yaml:"address"`
	Name    string `yaml:"name"`
}

// QUIC配置结构
type QuicConfig struct {
	ListenPort int              `yaml:"listen_port,omitempty"` // omitempty表示如果为0则不输出到YAML
	Upstreams  []UpstreamConfig `yaml:"upstreams,omitempty"`
}

// 主配置结构
type Config struct {
	NodeID  string        `yaml:"node_id"`
	Proxies []ProxyConfig `yaml:"proxies,omitempty"`
	Quic    QuicConfig    `yaml:"quic,omitempty"`
}

// 生成随机节点ID
func generateNodeID() string {
	// 使用当前时间戳（秒）+ 随机数生成10位16进制字符串
	timestamp := time.Now().Unix()

	// 生成4字节随机数
	randomBytes := make([]byte, 4)
	rand.Read(randomBytes)

	// 组合时间戳和随机数，生成10位16进制字符串
	nodeID := fmt.Sprintf("%08x%02x", timestamp&0xFFFFFFFF, randomBytes[0])

	return nodeID
}

// 保存配置到文件
func (c *Config) SaveConfig(filename string) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("序列化配置失败: %v", err)
	}

	if err := ioutil.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("写入配置文件失败: %v", err)
	}

	return nil
}

// 读取配置文件
func LoadConfig(filename string) (*Config, error) {
	// 读取文件内容
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %v", err)
	}

	// 解析YAML
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %v", err)
	}

	// 检查节点ID是否为空，如果为空则生成新的
	if config.NodeID == "" {
		config.NodeID = generateNodeID()
		fmt.Printf("⚠️  检测到节点ID为空，自动生成新ID: %s\n", config.NodeID)

		// 将新的配置写回文件
		if err := config.SaveConfig(filename); err != nil {
			return nil, fmt.Errorf("保存新节点ID失败: %v", err)
		}

		fmt.Printf("✅ 新节点ID已保存到配置文件\n")
	}

	// 验证配置
	if err := validateConfig(&config); err != nil {
		return nil, fmt.Errorf("配置验证失败: %v", err)
	}

	return &config, nil
}

// 验证配置
func validateConfig(config *Config) error {
	// 验证节点ID
	if config.NodeID == "" {
		return fmt.Errorf("节点ID不能为空")
	}
	if len(config.NodeID) < 8 || len(config.NodeID) > 12 {
		return fmt.Errorf("节点ID长度应该在8-12位之间")
	}

	// 验证代理配置
	for i, proxy := range config.Proxies {
		if proxy.LocalPort <= 0 || proxy.LocalPort > 65535 {
			return fmt.Errorf("代理[%d]本地端口号无效: %d", i, proxy.LocalPort)
		}
		if proxy.TargetNodeID == "" {
			return fmt.Errorf("代理[%d]目标节点ID不能为空", i)
		}
		if proxy.TargetAddress == "" {
			return fmt.Errorf("代理[%d]目标地址不能为空", i)
		}
		if proxy.Name == "" {
			return fmt.Errorf("代理[%d]名称不能为空", i)
		}
	}

	// 验证QUIC配置（如果配置了监听端口）
	if config.Quic.ListenPort > 0 {
		if config.Quic.ListenPort > 65535 {
			return fmt.Errorf("QUIC监听端口号无效: %d", config.Quic.ListenPort)
		}
	}

	// 验证上级节点配置（最多2个）
	if len(config.Quic.Upstreams) > 2 {
		return fmt.Errorf("上级节点最多只能配置2个，当前配置了%d个", len(config.Quic.Upstreams))
	}

	for i, upstream := range config.Quic.Upstreams {
		if upstream.NodeID == "" {
			return fmt.Errorf("上级节点[%d]节点ID不能为空", i)
		}
		if upstream.Address == "" {
			return fmt.Errorf("上级节点[%d]地址不能为空", i)
		}
		if upstream.Name == "" {
			return fmt.Errorf("上级节点[%d]名称不能为空", i)
		}
	}

	return nil
}

// 检查是否启用QUIC
func (c *Config) IsQuicEnabled() bool {
	return c.Quic.ListenPort > 0
}

// 打印配置信息
func (c *Config) PrintConfig() {
	fmt.Printf("=== FFMesh 配置信息 ===\n")
	fmt.Printf("节点ID: %s\n\n", c.NodeID)

	fmt.Printf("代理配置:\n")
	if len(c.Proxies) == 0 {
		fmt.Printf("  无代理配置\n")
	} else {
		for i, proxy := range c.Proxies {
			fmt.Printf("  [%d] %s\n", i+1, proxy.Name)
			fmt.Printf("      本地端口: %d (TCP)\n", proxy.LocalPort)
			fmt.Printf("      目标节点: %s\n", proxy.TargetNodeID)
			fmt.Printf("      目标地址: %s\n", proxy.TargetAddress)
		}
	}

	fmt.Printf("\nQUIC配置:\n")
	if c.IsQuicEnabled() {
		fmt.Printf("  监听端口: %d\n", c.Quic.ListenPort)
	} else {
		fmt.Printf("  监听端口: 未配置 (QUIC功能禁用)\n")
	}

	if len(c.Quic.Upstreams) == 0 {
		fmt.Printf("  上级节点: 无\n")
	} else {
		fmt.Printf("  上级节点:\n")
		for i, upstream := range c.Quic.Upstreams {
			fmt.Printf("    [%d] %s (%s) -> %s\n", i+1, upstream.Name, upstream.NodeID, upstream.Address)
		}
	}

	fmt.Printf("========================\n\n")
}

// 获取代理配置
func (c *Config) GetProxyByPort(port int) *ProxyConfig {
	for _, proxy := range c.Proxies {
		if proxy.LocalPort == port {
			return &proxy
		}
	}
	return nil
}

// 获取上级节点配置
func (c *Config) GetUpstreamByNodeID(nodeID string) *UpstreamConfig {
	for _, upstream := range c.Quic.Upstreams {
		if upstream.NodeID == nodeID {
			return &upstream
		}
	}
	return nil
}

// 示例使用
func main_testconfig() {
	// 加载配置文件
	config, err := LoadConfig("config.yaml")
	if err != nil {
		log.Fatal("加载配置失败:", err)
	}

	// 打印配置信息
	config.PrintConfig()

	// 启动FFMesh节点
	fmt.Printf("开始启动 FFMesh 节点: %s\n", config.NodeID)

	// 根据配置决定是否启动QUIC监听器
	if config.IsQuicEnabled() {
		fmt.Printf("启动 QUIC 监听器: 端口 %d\n", config.Quic.ListenPort)

		// 连接上级节点
		for _, upstream := range config.Quic.Upstreams {
			fmt.Printf("连接上级节点: %s (%s) -> %s\n", upstream.Name, upstream.NodeID, upstream.Address)
			// TODO: 实际连接上级节点
		}
	} else {
		fmt.Printf("QUIC功能未启用 (未配置监听端口)\n")
	}

	// 启动代理监听器
	if len(config.Proxies) > 0 {
		for _, proxy := range config.Proxies {
			fmt.Printf("启动代理监听器: %s -> 端口 %d\n", proxy.Name, proxy.LocalPort)
			fmt.Printf("  将代理到节点 %s 的 %s\n", proxy.TargetNodeID, proxy.TargetAddress)
			// TODO: 实际启动代理监听器
		}
	} else {
		fmt.Printf("无代理配置\n")
	}
}
