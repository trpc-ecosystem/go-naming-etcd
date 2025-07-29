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
	"errors"
	"sync"
	"time"

	tdiscovery "trpc.group/trpc-go/trpc-go/naming/discovery"
	tregistry "trpc.group/trpc-go/trpc-go/naming/registry"
	etcderror "trpc.group/trpc-go/trpc-naming-etcd/error"
	"trpc.group/trpc-go/trpc-naming-etcd/model"

	clientv3 "go.etcd.io/etcd/client/v3"
)

var (
	errStaleData = errors.New("store data is stale")
)

// cache
type cache struct {
	sync.RWMutex
	// nodeCache 缓存
	nodeCache map[string][]*tregistry.Node
	// expires 服务缓存过期时间
	expires map[string]time.Time
	// watched 是否被调用过，调用过才关注改变，否则不关注
	watched map[string]bool
	// 当前缓存的 etcd 数据版本
	version int64
	// 退出
	exit chan bool
	// watcher 监听 etcd 变更
	watcher *etcdWatcher
}

// setLocked 设置服务节点，必须要获取锁后操作
func (c *cache) setLocked(serviceName string, nodes []*tregistry.Node) {
	c.nodeCache[serviceName] = nodes
	c.expires[serviceName] = time.Now().Add(30 * time.Second)
}

// invalidCache 删除服务缓存
func (c *cache) invalidCache(serviceName string) {
	c.Lock()
	c.deleteLocked(serviceName)
	c.Unlock()
}

// deleteLocked 删除服务缓存，必须要获取锁后操作
func (c *cache) deleteLocked(serviceName string) {
	delete(c.nodeCache, serviceName)
	delete(c.expires, serviceName)
}

// cache 缓存服务节点
func (c *cache) cache(serviceName string, version int64, nodes []*tregistry.Node) error {
	if nodes == nil {
		return nil
	}

	c.Lock()
	defer c.Unlock()
	// 过时数据
	if version < c.version {
		return errStaleData
	}
	c.version = version
	if _, ok := c.watched[serviceName]; !ok {
		return nil
	}
	if len(nodes) == 0 {
		c.setLocked(serviceName, emptyNodes)
		return nil
	}
	c.setLocked(serviceName, nodes)
	return nil
}

// update 根据 etcd 变更更新缓存
func (c *cache) update(result *watchResult) {
	if result == nil || result.Node == nil {
		return
	}

	c.Lock()
	defer c.Unlock()
	serviceName := result.Node.Name
	// 过时数据
	if result.Version < c.version {
		return
	}
	// 不关注的节点直接返回
	if _, ok := c.watched[serviceName]; !ok {
		return
	}
	nodes, ok := c.nodeCache[serviceName]
	if !ok {
		// 只在获取过全量数据后才开始增量更新
		return
	}
	// 更新数据版本
	c.version = result.Version

	var node *tregistry.Node
	var index int
	for i, n := range nodes {
		if n.Address == result.Node.Address {
			node = n
			index = i
		}
	}

	switch result.EventType {
	case Create, Update:
		// 之前没有缓存过该节点则新增
		if node == nil {
			c.setLocked(serviceName, append(nodes, model.ConvertNode(result.Node)))
			return
		}
		// 之前已经缓存过该节点则覆盖
		nodes[index] = model.ConvertNode(result.Node)
	case Delete:
		if node == nil {
			return
		}
		var newCacheNodes []*tregistry.Node
		for _, cacheNode := range nodes {
			if cacheNode.Address != result.Node.Address {
				newCacheNodes = append(newCacheNodes, cacheNode)
			}
		}
		if len(newCacheNodes) == 0 {
			newCacheNodes = emptyNodes
		}
		c.setLocked(serviceName, newCacheNodes)
	default:
		return
	}
}

// List 从缓存获取服务节点
func (c *cache) List(serviceName string, opts ...tdiscovery.Option) ([]*tregistry.Node, error) {
	// 先从缓存拿
	c.RLock()
	nodes := c.nodeCache[serviceName]
	expire := c.expires[serviceName]
	// 缓存是否过期
	if c.isValid(nodes, expire) {
		c.RUnlock()
		if len(nodes) == 0 {
			return nodes, etcderror.ErrServerNotAvailable
		}
		return nodes, nil
	}
	// 设置关注的服务
	_, ok := c.watched[serviceName]
	c.RUnlock()
	if !ok {
		c.Lock()
		c.watched[serviceName] = true
		c.Unlock()
	}
	return nil, nil
}

// isValid 判断缓存是否有效
func (c *cache) isValid(nodes []*tregistry.Node, expire time.Time) bool {
	if nodes == nil {
		return false
	}
	if expire.IsZero() {
		return false
	}
	if time.Since(expire) > 0 {
		return false
	}
	return true
}

// watch 关注 etcd 改变来修改缓存
func (c *cache) watch() {
	resultChan := c.watcher.watch()
	for result := range resultChan {
		select {
		case <-c.exit:
			return
		default:
			c.update(result)
		}
	}
}

// Stop 主要用来停止
func (c *cache) stop() {
	c.Lock()
	defer c.Unlock()

	select {
	case <-c.exit:
		return
	default:
		close(c.exit)
	}
	c.watcher.stop()
}

// newCache 新建缓存
func newCache(etcdClient *clientv3.Client, cfg *Config) (*cache, error) {
	watcher := newEtcdWatcher(etcdClient, cfg)
	c := &cache{
		watched:   make(map[string]bool),
		nodeCache: make(map[string][]*tregistry.Node),
		expires:   make(map[string]time.Time),
		exit:      make(chan bool),
		watcher:   watcher,
	}
	go c.watch()
	return c, nil
}
