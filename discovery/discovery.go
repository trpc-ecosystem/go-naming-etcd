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

// Package discovery 服务发现
package discovery

import (
	"sync"

	"trpc.group/trpc-go/trpc-go/log"
	tdiscovery "trpc.group/trpc-go/trpc-go/naming/discovery"
	tregistry "trpc.group/trpc-go/trpc-go/naming/registry"
	"trpc.group/trpc-go/trpc-naming-etcd/client"
	"trpc.group/trpc-go/trpc-naming-etcd/model"

	clientv3 "go.etcd.io/etcd/client/v3"
	"golang.org/x/sync/singleflight"
)

var (
	emptyNodes = make([]*tregistry.Node, 0)
)

// Config 配置
type Config struct {
	// Prefix 注册前缀
	Prefix string
}

// Discovery 服务发现
type Discovery struct {
	sync.RWMutex
	cache      *cache
	sg         singleflight.Group
	etcdClient *clientv3.Client
	cfg        *Config
}

// NewDiscovery 新建etcd服务发现
func NewDiscovery(etcdClient *clientv3.Client, cfg *Config) (tdiscovery.Discovery, error) {
	if cfg.Prefix == "" {
		cfg.Prefix = client.DefaultEtcdPrefix
	}
	c, err := newCache(etcdClient, cfg)
	if err != nil {
		return nil, err
	}
	e := &Discovery{
		cache:      c,
		etcdClient: etcdClient,
		cfg:        cfg,
	}

	return e, nil
}

// List 获取serviceName的节点
func (d *Discovery) List(serviceName string, opts ...tdiscovery.Option) ([]*tregistry.Node, error) {
	nodes, err := d.cache.List(serviceName, opts...)
	if err != nil {
		return nil, err
	}
	if len(nodes) > 0 {
		return nodes, nil
	}
	// 缓存没找到，去etcd获取
	val, err, _ := d.sg.Do(serviceName, func() (interface{}, error) {
		version, nodes, e := d.listFromEtcd(serviceName, opts...)
		if e != nil {
			return nil, e
		}
		cacheErr := d.cache.cache(serviceName, version, nodes)
		// 如果缓存返回数据过期，代表获取节点期间服务有更新，删除缓存下一次重新获取
		if cacheErr == errStaleData {
			d.cache.invalidCache(serviceName)
		}
		return nodes, nil
	})
	if err != nil {
		log.Errorf("get %s node from etcd fail, err = %v", serviceName, err)
		return nil, err
	}
	return val.([]*tregistry.Node), nil
}

// listFromEtcd 获取serviceName在注册中心注册的节点
func (d *Discovery) listFromEtcd(serviceName string, opts ...tdiscovery.Option) (int64, []*tregistry.Node, error) {
	o := &tdiscovery.Options{}
	for _, opt := range opts {
		opt(o)
	}

	// 从etcd获取
	rsp, err := d.etcdClient.Get(o.Ctx, model.ServicePath(d.cfg.Prefix, serviceName), clientv3.WithPrefix())
	if err != nil {
		return 0, nil, err
	}
	var services []*tregistry.Node
	for _, n := range rsp.Kvs {
		node, err := model.Unmarshal(n.Value)
		if err != nil {
			log.Errorf("unmarshal node fail, err: %s\n", err.Error())
			return 0, nil, err
		}
		services = append(services, model.ConvertNode(node))
	}
	// 没有节点注册
	if services == nil {
		services = emptyNodes
	}

	return rsp.Header.Revision, services, nil
}
