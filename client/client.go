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

// Package client etcd 客户端功能封装
package client

import (
	"strings"
	"time"

	"github.com/pkg/errors"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/pkg/transport"
	"trpc.group/trpc-go/trpc-go/log"
)

const (
	// DefaultTimeout 默认 etcd 超时时间
	DefaultTimeout = 5 * time.Second
	// DefaultEtcdPrefix 默认 etcd 注册前缀
	DefaultEtcdPrefix = " trpc/registry/services/"
	// DefaultTTL 默认服务 ttl 时间
	DefaultTTL = 5
	// DefaultWeight 默认服务权重
	DefaultWeight = 1
)

// Config etcd 配置
type Config struct {
	Address  string `yaml:"address,omitempty"`
	Timeout  int    `yaml:"timeout,omitempty"`
	Username string `yaml:"username,omitempty"`
	Password string `yaml:"password,omitempty"`
	Prefix   string `yaml:"Prefix,omitempty"`
	CertFile string `yaml:"certfile,omitempty"`
	KeyFile  string `yaml:"keyfile,omitempty"`
	CaFile   string `yaml:"cafile,omitempty"`
}

// GenerateEtcdClient 生成 etcd 客户端
func GenerateEtcdClient(cfg *Config) (*clientv3.Client, error) {
	address := strings.Split(cfg.Address, ",")
	config := clientv3.Config{
		Endpoints: address,
	}

	if cfg.Timeout == 0 {
		config.DialTimeout = DefaultTimeout
	} else {
		config.DialTimeout = time.Duration(cfg.Timeout) * time.Second
	}
	config.Username = cfg.Username
	config.Password = cfg.Password

	if cfg.CertFile != "" && cfg.KeyFile != "" && cfg.CaFile != "" {
		tlsInfo := transport.TLSInfo{
			CertFile:      cfg.CertFile,
			KeyFile:       cfg.KeyFile,
			TrustedCAFile: cfg.CaFile,
		}
		tlsConfig, err := tlsInfo.ClientConfig()
		if err != nil {
			log.Errorf("init tlsconfig failed, err: %s", err)
			return nil, errors.Wrap(err, "init tlsconfig failed")
		}
		config.TLS = tlsConfig
	}

	return clientv3.New(config)
}
