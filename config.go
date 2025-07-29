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

package naming

// FactoryConfig 组件配置
type FactoryConfig struct {
	Address     string            `yaml:"address,omitempty"`
	Timeout     int               `yaml:"timeout,omitempty"`
	Username    string            `yaml:"username,omitempty"`
	Password    string            `yaml:"password,omitempty"`
	Prefix      string            `yaml:"Prefix,omitempty"`
	LoadBalance LoadBalanceConfig `yaml:"load_balance,omitempty"`
	TLS         TLSConfig         `yaml:"tls,omitempty"`
}

// LoadBalanceConfig 负载均衡配置
type LoadBalanceConfig struct {
	// Name 负载均衡策略
	Name string `yaml:"name,omitempty"`
}

// TLSConfig TLS配置
type TLSConfig struct {
	CertFile string `json:"certfile"`
	KeyFile  string `json:"keyfile"`
	CaFile   string `json:"cafile"`
}
