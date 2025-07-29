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

package discovery

import (
	"context"

	"trpc.group/trpc-go/trpc-go/log"
	"trpc.group/trpc-go/trpc-naming-etcd/model"

	clientv3 "go.etcd.io/etcd/client/v3"
)

// watchResult 包装 etcd 改变通知
type watchResult struct {
	Version   int64
	EventType EventType
	Node      *model.Node
}

// etcdWatcher etcd watcher
type etcdWatcher struct {
	exit       chan bool
	watchPath  string
	etcdClient *clientv3.Client
	cfg        *Config
}

// newEtcdWatcher 新建 etcd watcher
func newEtcdWatcher(etcdClient *clientv3.Client, cfg *Config) *etcdWatcher {
	return &etcdWatcher{
		etcdClient: etcdClient,
		cfg:        cfg,
		exit:       make(chan bool),
		watchPath:  model.ServicePath(cfg.Prefix, ""),
	}
}

// Stop 停止watch
func (ew *etcdWatcher) stop() {
	close(ew.exit)
}

// Watch 返回etcd变更
func (ew *etcdWatcher) watch() <-chan *watchResult {
	resultChan := make(chan *watchResult)
	go func() {
		defer func() {
			close(resultChan)
		}()
		ctx, cancel := context.WithCancel(context.Background())
		go func() {
			<-ew.exit
			cancel()
		}()
		for wresp := range ew.etcdClient.Watch(ctx, ew.watchPath, clientv3.WithPrefix(), clientv3.WithPrevKV()) {
			if wresp.Err() != nil {
				return
			}
			var result *watchResult
			for _, ev := range wresp.Events {
				value := ev.Kv.Value
				var eventType EventType
				switch ev.Type {
				case clientv3.EventTypePut:
					if ev.IsCreate() {
						eventType = Create
					} else if ev.IsModify() {
						eventType = Update
					}
				case clientv3.EventTypeDelete:
					eventType = Delete
					value = ev.PrevKv.Value
				}
				node, err := model.Unmarshal(value)
				if err != nil {
					log.Errorf("unmarshal node fail, err: %s\n", err.Error())
					continue
				}
				if node == nil {
					continue
				}
				result = &watchResult{
					Version:   wresp.Header.Revision,
					EventType: eventType,
					Node:      node,
				}
			}
			if result == nil {
				continue
			}
			select {
			case resultChan <- result:
			case <-ew.exit:
				return
			}
		}
	}()
	return resultChan
}
