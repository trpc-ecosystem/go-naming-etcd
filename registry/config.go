//
//
// Tencent is pleased to support the open source community by making tRPC available.
//
// Copyright (C) 2025 Tencent.
// All rights reserved.
//
// If you have downloaded a copy of the tRPC source code from Tencent,
// please note that tRPC source code is licensed under the Apache 2.0 License,
// A copy of the Apache 2.0 License is included in this file.
//
//

package registry

// TLSConfig TLS配置
type TLSConfig struct {
	CertFile string `json:"certfile"`
	KeyFile  string `json:"keyfile"`
	CaFile   string `json:"cafile"`
}

// Service 服务配置
type Service struct {
	ServiceName string            `yaml:"name,omitempty"`
	Weight      int               `yaml:"weight,omitempty"`
	TTL         int               `yaml:"ttl,omitempty"`
	Metadata    map[string]string `yaml:"metadata,omitempty"`
}

// FactoryConfig 组件配置
type FactoryConfig struct {
	Address  string    `yaml:"address,omitempty"`
	Timeout  int       `yaml:"timeout,omitempty"`
	Username string    `yaml:"username,omitempty"`
	Password string    `yaml:"password,omitempty"`
	TLS      TLSConfig `yaml:"tls,omitempty"`
	Prefix   string    `yaml:"Prefix,omitempty"`
	Services []Service `yaml:"service"`
}

// Config 配置
type Config struct {
	// Prefix 注册前缀
	Prefix string
	// Weight 权重
	Weight int `yaml:"weight,omitempty"`
	// TTL 租约过期时间 单位秒，默认5秒
	TTL int `yaml:"ttl,omitempty"`
	// Metadata 元数据
	Metadata map[string]string `yaml:"metadata,omitempty"`
}
