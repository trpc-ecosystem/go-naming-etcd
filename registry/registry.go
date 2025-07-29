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

// Package registry 将服务注册到 etcd 中，并且关注 etcd 的变化
package registry

import (
	"context"
	"fmt"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/cenkalti/backoff/v4"

	"trpc.group/trpc-go/trpc-go/log"
	tregistry "trpc.group/trpc-go/trpc-go/naming/registry"
	"trpc.group/trpc-go/trpc-naming-etcd/client"
	"trpc.group/trpc-go/trpc-naming-etcd/model"

	clientv3 "go.etcd.io/etcd/client/v3"
)

// Registry etcd 注册对象
type Registry struct {
	cfg          *Config
	pid          string
	leaseManager client.LeaseManager
	etcdClient   *clientv3.Client
	ctx          context.Context
	cancel       context.CancelFunc
	id           string
}

// NewRegistry 新建 etcd 注册对象
func NewRegistry(etcdClient *clientv3.Client, cfg *Config) (tregistry.Registry, error) {
	if cfg.Prefix == "" {
		cfg.Prefix = client.DefaultEtcdPrefix
	}
	if cfg.Weight == 0 {
		cfg.Weight = client.DefaultWeight
	}
	if cfg.TTL == 0 {
		cfg.TTL = client.DefaultTTL
	}
	ctx, cancel := context.WithCancel(context.Background())
	e := &Registry{
		cfg:          cfg,
		pid:          strconv.Itoa(os.Getpid()),
		leaseManager: client.NewLeaseManager(etcdClient),
		etcdClient:   etcdClient,
		ctx:          ctx,
		cancel:       cancel,
	}
	return e, nil
}

// Register 注册服务
func (r *Registry) Register(serviceName string, opts ...tregistry.Option) error {
	options := &tregistry.Options{}
	for _, opt := range opts {
		opt(options)
	}

	host, port, err := net.SplitHostPort(options.Address)
	if err != nil {
		return err
	}
	id := model.ServiceID(host, port, r.pid)
	r.id = id
	node := &model.Node{
		Name:     serviceName,
		ID:       id,
		Address:  fmt.Sprintf("%s:%s", host, port),
		Metadata: r.cfg.Metadata,
		Weight:   r.cfg.Weight,
	}
	// 开始注册
	go r.etcdRegister(node)
	return nil
}

// etcdRegister 注册到etcd
func (r *Registry) etcdRegister(node *model.Node) {
	key := model.NodePath(r.cfg.Prefix, node.Name, node.ID)
	value, err := model.Marshal(node)
	if err != nil {
		log.Errorf("marshal node fail, err: %s\n", err.Error())
		return
	}
	for {
		select {
		case <-r.ctx.Done():
			return
		default:
		}
		var leaseExpire chan bool
		operation := func() error {
			// 获取租约
			var leaseID clientv3.LeaseID
			leaseID, leaseExpire, err = r.leaseManager.GetLease(r.ctx, time.Duration(r.cfg.TTL)*time.Second)
			if err != nil {
				log.Tracef("get lease fail, serviceName:%s, err:%v", node.Name, err)
				return err
			}
			// 注册
			if _, err = r.etcdClient.Put(r.ctx, key, value, clientv3.WithLease(leaseID)); err != nil {
				log.Tracef("register %s fail, err:%v", node.Name, err)
				return err
			}
			log.Tracef("register %s success", node.Name)
			return nil
		}
		if err = backoff.Retry(operation, backoff.NewExponentialBackOff()); err != nil {
			continue
		}
		select {
		case <-leaseExpire:
			continue
		case <-r.ctx.Done():
			return
		}
	}
}

// Deregister 取消注册
func (r *Registry) Deregister(serviceName string) error {
	r.cancel()
	ctx, cancel := context.WithTimeout(context.Background(), client.DefaultTimeout)
	defer cancel()
	if _, err := r.etcdClient.Delete(ctx, model.NodePath(r.cfg.Prefix, serviceName, r.id)); err != nil {
		return err
	}
	return nil
}
