//
//
// Tencent is pleased to support the open source community by making tRPC available.
//
// Copyright (C) 2023 Tencent.
// All rights reserved.
//
// If you have downloaded a copy of the tRPC source code from Tencent,
// please note that tRPC source code is licensed under the Apache 2.0 License,
// A copy of the Apache 2.0 License is included in this file.
//
//

// Package registry 提供etcd注册服务
package discovery

import (
	"context"
	"math/rand"
	"testing"

	tdiscovery "trpc.group/trpc-go/trpc-go/naming/discovery"
	"trpc.group/trpc-go/trpc-naming-etcd/model"

	"go.etcd.io/etcd/api/v3/etcdserverpb"
	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"

	. "github.com/glycerine/goconvey/convey"
)

// etcdLease 实现etcd Lease接口
type etcdLease struct {
}

// Grant 分配租约
func (e *etcdLease) Grant(ctx context.Context, ttl int64) (*clientv3.LeaseGrantResponse, error) {
	return &clientv3.LeaseGrantResponse{
		ID:  clientv3.LeaseID(rand.Int()),
		TTL: ttl,
	}, nil
}

// Revoke 移除租约
func (e *etcdLease) Revoke(ctx context.Context, id clientv3.LeaseID) (*clientv3.LeaseRevokeResponse, error) {
	return &clientv3.LeaseRevokeResponse{}, nil
}

// TimeToLive 获取租约
func (e *etcdLease) TimeToLive(ctx context.Context, id clientv3.LeaseID,
	opts ...clientv3.LeaseOption) (*clientv3.LeaseTimeToLiveResponse, error) {
	return &clientv3.LeaseTimeToLiveResponse{}, nil
}

// Leases 获取所有租约
func (e *etcdLease) Leases(ctx context.Context) (*clientv3.LeaseLeasesResponse, error) {
	return &clientv3.LeaseLeasesResponse{}, nil
}

// KeepAlive 监听keepAlive
func (e *etcdLease) KeepAlive(ctx context.Context, id clientv3.LeaseID) (<-chan *clientv3.LeaseKeepAliveResponse,
	error) {
	ch := make(chan *clientv3.LeaseKeepAliveResponse, 10)
	ch <- &clientv3.LeaseKeepAliveResponse{
		ID: id,
	}
	return ch, nil
}

// KeepAliveOnce keepAlive一次
func (e *etcdLease) KeepAliveOnce(ctx context.Context, id clientv3.LeaseID) (*clientv3.LeaseKeepAliveResponse, error) {
	return &clientv3.LeaseKeepAliveResponse{
		ID: id,
	}, nil
}

// Close 关闭
func (e *etcdLease) Close() error {
	return nil
}

// newLease 新建一个Lease，实现etcd的Lease接口
func newLease() clientv3.Lease {
	return &etcdLease{}
}

// registryWatcher 实现etcd的Watcher接口
type registryWatcher struct {
}

// Watch 关注key变化
func (c *registryWatcher) Watch(ctx context.Context, key string, opts ...clientv3.OpOption) clientv3.WatchChan {
	return make(chan clientv3.WatchResponse, 1)
}

// RequestProgress 请求处理
func (c *registryWatcher) RequestProgress(ctx context.Context) error {
	return nil
}

// Close 关闭clientWatcher
func (c *registryWatcher) Close() error {
	return nil
}

// registryKv 实现etcd的KV接口
type registryKv struct {
}

// Put 存储kv
func (r *registryKv) Put(ctx context.Context, key, val string,
	opts ...clientv3.OpOption) (*clientv3.PutResponse, error) {
	return &clientv3.PutResponse{}, nil
}

// Get 获取kv
func (r *registryKv) Get(ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.GetResponse, error) {
	node := &model.Node{
		Name:     "test",
		ID:       model.ServiceID("127.0.0.1", "8080", "123"),
		Address:  "127.0.0.1:8080",
		Metadata: map[string]string{"key": "value"},
		Weight:   100,
	}
	value, _ := model.Marshal(node)
	return &clientv3.GetResponse{
		Kvs: []*mvccpb.KeyValue{
			{
				Key:            []byte(key),
				CreateRevision: 100,
				ModRevision:    100,
				Value:          []byte(value),
			},
		},
		Header: &etcdserverpb.ResponseHeader{
			Revision: rand.Int63(),
		},
	}, nil
}

// Delete 删除kv
func (r *registryKv) Delete(ctx context.Context, key string,
	opts ...clientv3.OpOption) (*clientv3.DeleteResponse, error) {
	return &clientv3.DeleteResponse{}, nil
}

// Compact 压缩kv历史
func (r *registryKv) Compact(ctx context.Context, rev int64,
	opts ...clientv3.CompactOption) (*clientv3.CompactResponse, error) {
	return &clientv3.CompactResponse{}, nil
}

// Do 不使用事务执行op
func (r *registryKv) Do(ctx context.Context, op clientv3.Op) (clientv3.OpResponse, error) {
	return clientv3.OpResponse{}, nil
}

// Txn 事务操作
func (r *registryKv) Txn(ctx context.Context) clientv3.Txn {
	return nil
}

// newDiscoveryEtcdClient 获取etcd客户端
func newDiscoveryEtcdClient() *clientv3.Client {
	client, _ := clientv3.New(clientv3.Config{
		Endpoints: []string{"127.0.0.1:2379"},
	})
	client.Watcher = newDiscoveryWatcher()
	client.Lease = newLease()
	client.KV = newRegistryKv()
	return client
}

// newRegistryKv 新建一个watcher，实现etcd的KV接口
func newRegistryKv() clientv3.KV {
	return &registryKv{}
}

// newDiscoveryWatcher 新建一个watcher，实现etcd的watcher接口
func newDiscoveryWatcher() clientv3.Watcher {
	return &cacheWatcher{}
}

// newEtcdRegistry 新建etcd注册客户端
func newEtcdRegistry() *Discovery {
	c := newDiscoveryEtcdClient()
	r, _ := NewDiscovery(c, &Config{})
	return r.(*Discovery)
}

func TestEtcdDiscovery_List(t *testing.T) {
	Convey("测试registry的List函数", t, func() {
		r := newEtcdRegistry()
		nodes, err := r.List("test", tdiscovery.WithContext(context.Background()))
		So(err, ShouldBeNil)
		So(len(nodes), ShouldNotEqual, 0)
	})

}
