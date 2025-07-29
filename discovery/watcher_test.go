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
	"testing"

	"trpc.group/trpc-go/trpc-naming-etcd/model"

	"go.etcd.io/etcd/api/v3/etcdserverpb"
	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"

	. "github.com/glycerine/goconvey/convey"
)

// watcher 实现etcd的Watcher接口
type watcher struct {
}

// Watch 关注key变化
func (c *watcher) Watch(ctx context.Context, key string, opts ...clientv3.OpOption) clientv3.WatchChan {
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
	// 更新
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
					ModRevision:    101,
				},
			},
		},
		Canceled: false,
	}
	// 删除
	etcdCh <- clientv3.WatchResponse{
		Header: etcdserverpb.ResponseHeader{
			Revision: 100,
		},
		Events: []*clientv3.Event{
			{
				PrevKv: &mvccpb.KeyValue{
					Value: []byte(value),
				},
				Type: clientv3.EventTypeDelete,
				Kv: &mvccpb.KeyValue{
					Key:            []byte(key),
					Value:          []byte(value),
					CreateRevision: 100,
					ModRevision:    101,
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
}

// RequestProgress 请求处理
func (c *watcher) RequestProgress(ctx context.Context) error {
	return nil
}

// Close 关闭clientWatcher
func (c *watcher) Close() error {
	return nil
}

// newEtcdClient 获取etcd客户端
func newEtcdClient() *clientv3.Client {
	client, _ := clientv3.New(clientv3.Config{
		Endpoints: []string{"127.0.0.1:2379"},
	})
	client.Watcher = newWatcher()
	return client
}

// newWatcher 新建一个watcher，实现etcd的watcher接口
func newWatcher() clientv3.Watcher {
	return &watcher{}
}

func Test_etcdWatcher_stop(t *testing.T) {
	Convey("测试watcher的isValid函数", t, func() {
		client := newEtcdClient()
		w := newEtcdWatcher(client, &Config{})
		So(w, ShouldNotBeNil)
		w.stop()
	})
}

func Test_etcdWatcher_watch(t *testing.T) {
	Convey("测试watcher的isValid函数", t, func() {
		client := newEtcdClient()
		w := newEtcdWatcher(client, &Config{})
		So(w, ShouldNotBeNil)
		resultChan := w.watch()
		for range resultChan {

		}
	})
}

func Test_newEtcdWatcher(t *testing.T) {
	Convey("测试新建watcher", t, func() {
		client := newEtcdClient()
		w := newEtcdWatcher(client, &Config{})
		So(w, ShouldNotBeNil)
	})
}
