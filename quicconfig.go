package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"net"
	"sync"
	"time"

	"github.com/quic-go/quic-go"
)

// 全局TLS配置缓存
var (
	tlsConfigOnce sync.Once
	tlsConfig     *tls.Config
)

// 生成自签名TLS证书
func generateTLSConfig() *tls.Config {
	// 生成RSA私钥
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}

	// 创建证书模板
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization:  []string{"FFMesh"},
			Country:       []string{"CN"},
			Province:      []string{""},
			Locality:      []string{"Beijing"},
			StreetAddress: []string{""},
			PostalCode:    []string{""},
		},
		NotBefore:   time.Now().Add(-24 * time.Hour),            // 当前时间减去1天，避免时间差异
		NotAfter:    time.Now().Add(100 * 365 * 24 * time.Hour), // 100年有效期
		KeyUsage:    x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		IPAddresses: []net.IP{
			net.IPv4(127, 0, 0, 1),
			net.IPv4(0, 0, 0, 0),
			net.IPv6loopback,
			net.IPv6unspecified,
		},
		DNSNames: []string{"localhost", "*"},
	}

	// 创建证书
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		panic(err)
	}

	return &tls.Config{
		Certificates: []tls.Certificate{{
			Certificate: [][]byte{certDER},
			PrivateKey:  key,
		}},
		NextProtos: []string{"ffmesh-quic"},
	}
}

// 获取QUIC服务端配置
func GetQuicServerConfig() *quic.Config {
	return &quic.Config{
		// 握手超时时间
		HandshakeIdleTimeout: 10 * time.Second,

		// 连接空闲超时时间
		MaxIdleTimeout: 30 * time.Second,

		// 最大传入流数量
		MaxIncomingStreams: 100,

		// 最大传入单向流数量
		MaxIncomingUniStreams: 100,

		// 禁用数据报
		EnableDatagrams: false,

		// 禁用路径MTU发现
		DisablePathMTUDiscovery: true,

		// 保持连接活跃
		KeepAlivePeriod: 10 * time.Second,

		// 禁用0-RTT
		Allow0RTT: false,
	}
}

// 获取QUIC客户端配置
func GetQuicClientConfig() *quic.Config {
	return &quic.Config{
		// 握手超时时间
		HandshakeIdleTimeout: 10 * time.Second,

		// 连接空闲超时时间
		MaxIdleTimeout: 30 * time.Second,

		// 最大传入流数量
		MaxIncomingStreams: 10,

		// 最大传入单向流数量
		MaxIncomingUniStreams: 10,

		// 禁用路径MTU发现
		DisablePathMTUDiscovery: true,

		// 保持连接活跃
		KeepAlivePeriod: 10 * time.Second,

		// 禁用0-RTT
		Allow0RTT: false,
	}
}

// 获取客户端TLS配置（跳过证书验证）
func GetClientTLSConfig() *tls.Config {
	return &tls.Config{
		// 跳过证书验证
		InsecureSkipVerify: true,

		// 应用层协议协商
		NextProtos: []string{"ffmesh-quic"},
	}
}

// 获取服务端TLS配置
func GetServerTLSConfig() *tls.Config {
	tlsConfigOnce.Do(func() {
		tlsConfig = generateTLSConfig()
	})
	return tlsConfig
}
