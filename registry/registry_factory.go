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

import (
	"trpc.group/trpc-go/trpc-go/naming/registry"
	tselector "trpc.group/trpc-go/trpc-go/naming/selector"
	"trpc.group/trpc-go/trpc-go/plugin"
	"trpc.group/trpc-go/trpc-naming-etcd/client"
	"trpc.group/trpc-go/trpc-naming-etcd/selector"
)

func init() {
	plugin.Register(pluginName, &Plugin{})
	s := &selector.Selector{}
	tselector.Register(pluginName, s)
}

const (
	pluginType = "registry"
	pluginName = "etcd"
)

// Plugin 插件结构
type Plugin struct{}

// Type 插件类型
func (p *Plugin) Type() string {
	return pluginType
}

// Setup 注册
func (p *Plugin) Setup(name string, decoder plugin.Decoder) error {
	factoryCfg := &FactoryConfig{}
	if err := decoder.Decode(factoryCfg); err != nil {
		return err
	}
	etcdClient, err := client.GenerateEtcdClient(&client.Config{
		Address:  factoryCfg.Address,
		Timeout:  factoryCfg.Timeout,
		Username: factoryCfg.Username,
		Password: factoryCfg.Password,
		Prefix:   factoryCfg.Prefix,
		CaFile:   factoryCfg.TLS.CaFile,
		CertFile: factoryCfg.TLS.CertFile,
		KeyFile:  factoryCfg.TLS.KeyFile,
	})
	if err != nil {
		return err
	}
	for _, service := range factoryCfg.Services {
		cfg := &Config{
			Prefix:   factoryCfg.Prefix,
			Weight:   service.Weight,
			TTL:      service.TTL,
			Metadata: service.Metadata,
		}
		reg, err := NewRegistry(etcdClient, cfg)
		if err != nil {
			return err
		}
		registry.Register(service.ServiceName, reg)
	}
	return nil
}
