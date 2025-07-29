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

// Package selector 从 etcd 中获取服务节点并根据负载均衡策略返回节点
package selector

import (
	"time"

	tdiscovery "trpc.group/trpc-go/trpc-go/naming/discovery"
	"trpc.group/trpc-go/trpc-go/naming/loadbalance"
	"trpc.group/trpc-go/trpc-go/naming/registry"
	"trpc.group/trpc-go/trpc-go/naming/selector"
	tselector "trpc.group/trpc-go/trpc-go/naming/selector"
	etcderror "trpc.group/trpc-go/trpc-naming-etcd/error"
)

const (
	// defaultLoadBalancer 负载均衡策略
	defaultLoadBalancer = "random"
)

// Config 配置
type Config struct {
	//LoadBalancer 负载均衡策略
	LoadBalancer string
}

// Selector 路由
type Selector struct {
	cfg       *Config
	discovery tdiscovery.Discovery
}

// NewSelector 新建路由
func NewSelector(d tdiscovery.Discovery, cfg *Config) tselector.Selector {
	if cfg.LoadBalancer == "" {
		cfg.LoadBalancer = defaultLoadBalancer
	}
	return &Selector{
		cfg:       cfg,
		discovery: d,
	}
}

// Select 选择节点
func (s *Selector) Select(serviceName string, opts ...selector.Option) (*registry.Node, error) {
	o := &selector.Options{}
	for _, opt := range opts {
		opt(o)
	}
	nodes, err := s.discovery.List(serviceName, tdiscovery.WithContext(o.Ctx))
	if err != nil {
		return nil, err
	}
	load := loadbalance.Get(s.cfg.LoadBalancer)
	if load == nil {
		return nil, etcderror.ErrBalancerNotExist
	}
	loadBalanceOpts := []loadbalance.Option{
		loadbalance.WithContext(o.Ctx),
		loadbalance.WithLoadBalanceType(s.cfg.LoadBalancer),
		loadbalance.WithKey(o.Key),
		loadbalance.WithNamespace(o.Namespace),
		loadbalance.WithReplicas(o.Replicas),
	}
	return load.Select(serviceName, nodes, loadBalanceOpts...)
}

// Report 上报调用结果
func (s *Selector) Report(node *registry.Node, cost time.Duration, err error) error {
	return nil
}
