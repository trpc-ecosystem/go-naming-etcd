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

package client

import (
	"context"
	"sync"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

var (
	defaultForceKeepAliveTime = time.Second
)

// LeaseManager 租约管理
type LeaseManager interface {
	GetLease(ctx context.Context, ttl time.Duration) (clientv3.LeaseID, chan bool, error)
}

// leaseHolder 租约包装
type leaseHolder struct {
	leaseID            clientv3.LeaseID
	forceKeepAliveTime time.Time
	exit               chan bool
	ttl                time.Duration
}

// leaseManagerImpl 实现租约管理接口，由于租约是个代价较大的行为，因此同一个 ttl 只保留一个租约
type leaseManagerImpl struct {
	client   *clientv3.Client
	leaseMu  sync.Mutex
	leaseMap map[time.Duration]*leaseHolder
}

// NewLeaseManager 新建租约管理
func NewLeaseManager(client *clientv3.Client) LeaseManager {
	return &leaseManagerImpl{
		client:   client,
		leaseMap: make(map[time.Duration]*leaseHolder),
	}
}

// GetLease 获取租约
func (l *leaseManagerImpl) GetLease(ctx context.Context, ttl time.Duration) (clientv3.LeaseID, chan bool, error) {
	now := time.Now()
	l.leaseMu.Lock()
	defer l.leaseMu.Unlock()
	if lease, ok := l.leaseMap[ttl]; ok {
		if now.Before(lease.forceKeepAliveTime) {
			return lease.leaseID, lease.exit, nil
		}
		_, err := l.client.KeepAliveOnce(ctx, lease.leaseID)
		if err == nil {
			lease.forceKeepAliveTime = now.Add(defaultForceKeepAliveTime)
			return lease.leaseID, lease.exit, nil
		}
		l.removeLeaseLocked(lease)
	}
	leaseRsp, err := l.client.Grant(ctx, int64(ttl.Seconds()))
	if err != nil {
		return clientv3.LeaseID(0), nil, err
	}
	lease := &leaseHolder{
		leaseID:            leaseRsp.ID,
		forceKeepAliveTime: now.Add(defaultForceKeepAliveTime),
		exit:               make(chan bool),
		ttl:                ttl,
	}
	l.leaseMap[ttl] = lease
	go l.leaseKeepAlive(lease)
	return lease.leaseID, lease.exit, nil
}

// leaseKeepAlive 自动续约
func (l *leaseManagerImpl) leaseKeepAlive(lease *leaseHolder) {
	// 自动续租
	alive, err := l.client.KeepAlive(context.Background(), lease.leaseID)
	if err != nil {
		l.leaseMu.Lock()
		defer l.leaseMu.Unlock()
		l.removeLeaseLocked(lease)
		return
	}

	for range alive {
	}
	l.leaseMu.Lock()
	defer l.leaseMu.Unlock()
	l.removeLeaseLocked(lease)
}

// removeLeaseLocked 移除租约
func (l *leaseManagerImpl) removeLeaseLocked(lease *leaseHolder) {
	if existLease, ok := l.leaseMap[lease.ttl]; ok {
		if existLease.leaseID == lease.leaseID {
			delete(l.leaseMap, lease.ttl)
			close(existLease.exit)
		}
	}
}
