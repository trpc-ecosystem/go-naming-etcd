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

package client

import (
	"context"
	"errors"
	"math/rand"
	"reflect"
	"testing"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"

	. "github.com/agiledragon/gomonkey"
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

// newLeaseClient 获取etcd客户端
func newLeaseClient() *clientv3.Client {
	client, _ := clientv3.New(clientv3.Config{
		Endpoints: []string{"127.0.0.1:2379"},
	})
	client.Lease = newLease()
	return client
}

// newLease 新建一个Lease，实现etcd的Lease接口
func newLease() clientv3.Lease {
	return &etcdLease{}
}

func Test_newLeaseManager(t *testing.T) {
	Convey("新建租约管理", t, func() {
		c := newLeaseClient()
		m := NewLeaseManager(c)
		So(m, ShouldNotBeNil)
	})
}

func Test_leaseManager_GetLease(t *testing.T) {
	Convey("获取租约", t, func() {
		defaultForceKeepAliveTime = time.Millisecond * 100
		ttl := time.Second * 10
		c := newLeaseClient()
		lm := NewLeaseManager(c)
		firstLeaseId, _, err := lm.GetLease(context.Background(), ttl)
		So(err, ShouldBeNil)
		secondLeaseId, _, err := lm.GetLease(context.Background(), ttl)
		So(err, ShouldBeNil)
		// 同一个ttl获取到的租约一致
		So(firstLeaseId, ShouldEqual, secondLeaseId)

		// 睡眠使得下次获取租约需要强制keepALive
		time.Sleep(defaultForceKeepAliveTime)
		secondLeaseId, _, err = lm.GetLease(context.Background(), ttl)
		So(err, ShouldBeNil)
		// 同一个ttl获取到的租约一致
		So(firstLeaseId, ShouldEqual, secondLeaseId)

		// mock KeepAliveOnce返回错误使得管理器删除
		keepAliveOncePatch := ApplyMethod(reflect.TypeOf(c.Lease), "KeepAliveOnce", func(lease *etcdLease,
			ctx context.Context, id clientv3.LeaseID) (*clientv3.LeaseKeepAliveResponse, error) {
			return nil, errors.New("keepAliveOnce fail")
		})
		defer keepAliveOncePatch.Reset()
		// 睡眠使得下次获取租约需要强制keepALive
		time.Sleep(defaultForceKeepAliveTime)
		newLeaseId, _, err := lm.GetLease(context.Background(), ttl)
		So(err, ShouldBeNil)
		// 此时就变成了不同租约
		So(firstLeaseId, ShouldNotEqual, newLeaseId)
		// mock Grant返回错误
		grantPatch := ApplyMethod(reflect.TypeOf(c.Lease), "Grant", func(lease *etcdLease,
			ctx context.Context, ttl int64) (*clientv3.LeaseGrantResponse, error) {
			return nil, errors.New("grant fail")
		})
		defer grantPatch.Reset()
		// 睡眠使得下次获取租约需要强制keepALive
		time.Sleep(defaultForceKeepAliveTime)
		// 由于grant失败因此返回失败
		_, _, err = lm.GetLease(context.Background(), ttl)
		So(err, ShouldNotBeNil)
	})

}

func Test_leaseManager_leaseKeepAlive(t *testing.T) {
	Convey("续期租约", t, func() {
		c := newLeaseClient()
		lm := NewLeaseManager(c)

		lease := &leaseHolder{
			leaseID:            clientv3.LeaseID(rand.Int()),
			forceKeepAliveTime: time.Now().Add(defaultForceKeepAliveTime),
			exit:               make(chan bool),
			ttl:                time.Second * 10,
		}
		// mock KeepAlive 正常返回
		keepAlivePatch := ApplyMethod(reflect.TypeOf(c.Lease), "KeepAlive", func(lease *etcdLease,
			ctx context.Context, id clientv3.LeaseID) (<-chan *clientv3.LeaseKeepAliveResponse, error) {
			ch := make(chan *clientv3.LeaseKeepAliveResponse, 10)
			ch <- &clientv3.LeaseKeepAliveResponse{
				ID: id,
			}
			go func() {
				time.Sleep(time.Millisecond * 100)
				close(ch)
			}()
			return ch, nil
		})

		lmImpl, ok := lm.(*leaseManagerImpl)
		if !ok {
			return
		}

		lmImpl.leaseKeepAlive(lease)

		keepAlivePatch.Reset()

		// mock KeepAlive 返回错误
		keepAlivePatch = ApplyMethod(reflect.TypeOf(c.Lease), "KeepAlive", func(lease *etcdLease,
			ctx context.Context, id clientv3.LeaseID) (<-chan *clientv3.LeaseKeepAliveResponse, error) {
			return nil, errors.New("keepAlive fail")
		})
		lease.exit = make(chan bool, 1)
		lmImpl.leaseKeepAlive(lease)

	})
}
