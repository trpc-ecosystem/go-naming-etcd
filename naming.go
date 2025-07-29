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

// Package naming 服务发现插件
package naming

import (
	"trpc.group/trpc-go/trpc-naming-etcd/client"
	"trpc.group/trpc-go/trpc-naming-etcd/discovery"

	tselector "trpc.group/trpc-go/trpc-go/naming/selector"
	"trpc.group/trpc-go/trpc-go/plugin"
	"trpc.group/trpc-go/trpc-naming-etcd/selector"
)

func init() {
	plugin.Register(pluginName, &Plugin{})
	s := &selector.Selector{}
	tselector.Register(pluginName, s)
}

const (
	pluginType = "selector"
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
		CertFile: factoryCfg.TLS.CertFile,
		KeyFile:  factoryCfg.TLS.KeyFile,
		CaFile:   factoryCfg.TLS.CaFile,
	})
	if err != nil {
		return err
	}

	d, err := discovery.NewDiscovery(etcdClient, &discovery.Config{
		Prefix: factoryCfg.Prefix,
	})
	if err != nil {
		return err
	}
	tselector.Register(pluginName, selector.NewSelector(d, &selector.Config{
		LoadBalancer: factoryCfg.LoadBalance.Name,
	}))
	return nil
}
