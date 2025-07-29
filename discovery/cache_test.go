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

package discovery

import (
	"context"
	"reflect"
	"testing"
	"time"

	tregistry "trpc.group/trpc-go/trpc-go/naming/registry"
	"trpc.group/trpc-go/trpc-naming-etcd/model"

	"go.etcd.io/etcd/api/v3/etcdserverpb"
	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"

	. "github.com/agiledragon/gomonkey"
	. "github.com/glycerine/goconvey/convey"
)

// cacheWatcher 实现etcd的Watcher接口
type cacheWatcher struct {
}

// Watch 关注key变化
func (c *cacheWatcher) Watch(ctx context.Context, key string, opts ...clientv3.OpOption) clientv3.WatchChan {
	return make(chan clientv3.WatchResponse, 1)
}

// RequestProgress 请求处理
func (c *cacheWatcher) RequestProgress(ctx context.Context) error {
	return nil
}

// Close 关闭clientWatcher
func (c *cacheWatcher) Close() error {
	return nil
}

// newCacheEtcdClient 获取etcd客户端
func newCacheEtcdClient() *clientv3.Client {
	client, _ := clientv3.New(clientv3.Config{
		Endpoints: []string{"127.0.0.1:2379"},
	})
	client.Watcher = newClientWatcher()
	return client
}

// newClientWatcher 新建一个watcher，实现etcd的watcher接口
func newClientWatcher() clientv3.Watcher {
	return &cacheWatcher{}
}

func Test_cache_List(t *testing.T) {
	Convey("通过缓存获取服务列表", t, func() {
		client := newCacheEtcdClient()
		c, err := newCache(client, &Config{})
		So(err, ShouldBeNil)
		So(c, ShouldNotBeNil)
		// 兼容nil
		c.update(nil)
		firstNode := &model.Node{
			Name:     "test",
			ID:       model.ServiceID("127.0.0.1", "8080", "123"),
			Address:  "127.0.0.1:8080",
			Metadata: map[string]string{"key": "value"},
			Weight:   100,
		}
		// 由于没有关注过test服务，此时更新不会生效
		c.update(&watchResult{
			EventType: Create,
			Version:   1,
			Node:      firstNode,
		})
		// 由于节点为0，此时应该报错，同时List也代表着关注了test服务，后续更新可以生效
		nodes, err := c.List("test")
		So(err, ShouldBeNil)
		So(len(nodes), ShouldEqual, 0)
		// 先缓存一次，模拟获取过数据
		_ = c.cache("test", 2, emptyNodes)
		c.update(&watchResult{
			EventType: Create,
			Version:   2,
			Node:      firstNode,
		})
		// 此时能够获取到缓存
		nodes, err = c.List("test")
		So(err, ShouldBeNil)
		So(len(nodes), ShouldNotEqual, 0)
		firstNode.Weight = 50
		// 更新老版本数据，应该跳过不更新
		c.update(&watchResult{
			EventType: Update,
			Version:   3,
			Node:      firstNode,
		})
		// 此时能够获取到缓存
		nodes, err = c.List("test")
		So(err, ShouldBeNil)
		So(len(nodes), ShouldNotEqual, 0)
		So(nodes[0].Weight, ShouldEqual, 50)
		// 再添加一个节点
		secondNode := &model.Node{
			Name:     "test",
			ID:       model.ServiceID("127.0.0.1", "8081", "123"),
			Address:  "127.0.0.1:8081",
			Metadata: map[string]string{"key": "value"},
			Weight:   100,
		}
		c.update(&watchResult{
			EventType: Create,
			Version:   4,
			Node:      secondNode,
		})
		// 删除一个节点
		c.update(&watchResult{
			EventType: Delete,
			Version:   5,
			Node:      firstNode,
		})

		// 此时只能获取到第二个节点
		nodes, err = c.List("test")
		So(err, ShouldBeNil)
		So(len(nodes), ShouldEqual, 1)
		So(nodes[0].Address, ShouldEqual, "127.0.0.1:8081")

		// 删除第二个 节点
		c.update(&watchResult{
			EventType: Delete,
			Version:   6,
			Node:      secondNode,
		})
		// 此时什么也获取不到
		nodes, err = c.List("test")
		So(err, ShouldNotBeNil)
	})
}

func Test_cache_cache(t *testing.T) {
	Convey("cache", t, func() {
		client := newCacheEtcdClient()
		c, err := newCache(client, &Config{})
		So(err, ShouldBeNil)
		So(c, ShouldNotBeNil)
		// 缓存空
		err = c.cache("test", 2, nil)
		So(err, ShouldBeNil)

		newNode := &model.Node{
			Name:     "test",
			ID:       model.ServiceID("127.0.0.1", "8080", "123"),
			Address:  "127.0.0.1:8080",
			Metadata: map[string]string{"key": "value"},
			Weight:   100,
		}
		// 没有获取过缓存不会watch，就不会缓存
		err = c.cache("test", 2, []*tregistry.Node{model.ConvertNode(newNode)})
		nodes, err := c.List("test")
		So(len(nodes), ShouldEqual, 0)
		So(err, ShouldBeNil)

		// 先获取一次，代表关注此服务，后续此服务更新才能生效
		_, err = c.List("test")
		// 缓存
		err = c.cache("test", 2, []*tregistry.Node{model.ConvertNode(newNode)})
		So(err, ShouldBeNil)
		// 缓存老版本报错
		err = c.cache("test", 1, []*tregistry.Node{model.ConvertNode(newNode)})
		So(err, ShouldNotBeNil)
		nodes, err = c.List("test")
		So(len(nodes), ShouldEqual, 1)
		c.invalidCache("test")
		nodes, err = c.List("test")
		So(len(nodes), ShouldEqual, 0)
	})
}

func Test_cache_isValid(t *testing.T) {
	Convey("测试cache的isValid函数", t, func() {
		client := newCacheEtcdClient()
		c, err := newCache(client, &Config{})
		So(err, ShouldBeNil)
		So(c, ShouldNotBeNil)

		now := time.Now()
		So(c.isValid(nil, now), ShouldBeFalse)
		So(c.isValid(emptyNodes, time.Time{}), ShouldBeFalse)
		m, _ := time.ParseDuration("-1s")
		So(c.isValid(emptyNodes, now.Add(m)), ShouldBeFalse)
		m, _ = time.ParseDuration("1s")
		So(c.isValid([]*tregistry.Node{
			{
				ServiceName: "test",
			},
		}, now.Add(m)), ShouldBeTrue)
	})
}

func Test_cache_stop(t *testing.T) {
	Convey("测试cache的stop函数", t, func() {
		client := newCacheEtcdClient()
		c, err := newCache(client, &Config{})
		So(err, ShouldBeNil)
		So(c, ShouldNotBeNil)
		c.stop()
		// 可以多次停止
		c.stop()
	})
}

func Test_cache_watch(t *testing.T) {
	Convey("测试cache的watch函数", t, func() {
		client := newCacheEtcdClient()
		c, err := newCache(client, &Config{})
		So(err, ShouldBeNil)
		So(c, ShouldNotBeNil)
		nodes, err := c.List("test")
		So(err, ShouldBeNil)
		So(len(nodes), ShouldEqual, 0)
		// 先缓存一次，模拟获取过数据
		_ = c.cache("test", 1, emptyNodes)
		// mock Watch 正常返回
		watchPatch := ApplyMethod(reflect.TypeOf(client.Watcher), "Watch", func(lease *cacheWatcher,
			ctx context.Context, key string, opts ...clientv3.OpOption) clientv3.WatchChan {
			node := &model.Node{
				Name:     "test",
				ID:       model.ServiceID("127.0.0.1", "8080", "123"),
				Address:  "127.0.0.1:8080",
				Metadata: map[string]string{"key": "value"},
				Weight:   100,
			}
			value, _ := model.Marshal(node)
			etcdCh := make(chan clientv3.WatchResponse, 10)
			// 新建
			etcdCh <- clientv3.WatchResponse{
				Header: etcdserverpb.ResponseHeader{
					Revision: 100,
				},
				Events: []*clientv3.Event{
					{
						Type: clientv3.EventTypePut,
						Kv: &mvccpb.KeyValue{
							Key:            []byte(key),
							Value:          []byte(value),
							CreateRevision: 100,
							ModRevision:    100,
						},
					},
				},
				Canceled: false,
			}
			etcdCh <- clientv3.WatchResponse{
				Header: etcdserverpb.ResponseHeader{
					Revision: 100,
				},
				Canceled: true,
			}
			return etcdCh
		})
		defer watchPatch.Reset()
		c.watch()
	})
}

func Test_newCache(t *testing.T) {
	Convey("新建缓存", t, func() {
		client := newCacheEtcdClient()
		c, err := newCache(client, &Config{})
		So(err, ShouldBeNil)
		So(c, ShouldNotBeNil)
	})
}
