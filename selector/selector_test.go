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

package selector

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	tdiscovery "trpc.group/trpc-go/trpc-go/naming/discovery"
	"trpc.group/trpc-go/trpc-go/naming/loadbalance"
	tregistry "trpc.group/trpc-go/trpc-go/naming/registry"
	tselector "trpc.group/trpc-go/trpc-go/naming/selector"
	"trpc.group/trpc-go/trpc-naming-etcd/discovery"

	"github.com/golang/mock/gomock"
	clientv3 "go.etcd.io/etcd/client/v3"

	. "github.com/agiledragon/gomonkey"
	. "github.com/glycerine/goconvey/convey"
)

func TestSelector_Select(t *testing.T) {
	Convey("Select寻址", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		c, err := clientv3.New(clientv3.Config{
			Endpoints: []string{"127.0.0.1:2379"},
		})
		So(err, ShouldBeNil)
		var d tdiscovery.Discovery
		d, err = discovery.NewDiscovery(c, &discovery.Config{})
		So(err, ShouldBeNil)
		So(d, ShouldNotBeNil)
		s := NewSelector(d, &Config{})
		So(s, ShouldNotBeNil)

		patch := ApplyMethod(reflect.TypeOf(d), "List", func(e *discovery.Discovery, service string,
			opt ...tdiscovery.Option) (nodes []*tregistry.Node, err error) {
			return nil, nil
		}).ApplyFunc(loadbalance.Get, func(name string) loadbalance.LoadBalancer {
			return &loadbalance.Random{}
		})

		node, err := s.Select("test", tselector.WithContext(context.Background()))
		_ = s.Report(node, time.Second, err)

		// 使用不存在的load balance会报错
		_, err = s.Select("test", tselector.WithLoadBalanceType("test"))
		So(err, ShouldNotBeNil)
		patch.Reset()

		// mock List返回错误
		patch = ApplyMethod(reflect.TypeOf(d), "List", func(e *discovery.Discovery, service string,
			opt ...tdiscovery.Option) (nodes []*tregistry.Node, err error) {
			return nil, errors.New("list fail")
		})
		_, err = s.Select("test", tselector.WithContext(context.Background()))
		So(err, ShouldNotBeNil)
		patch.Reset()

		// mock load balance返回空
		patch = ApplyMethod(reflect.TypeOf(d), "List", func(e *discovery.Discovery, service string,
			opt ...tdiscovery.Option) (nodes []*tregistry.Node, err error) {
			return nil, nil
		}).ApplyFunc(loadbalance.Get, func(name string) loadbalance.LoadBalancer {
			return nil
		})
		_, err = s.Select("test", tselector.WithContext(context.Background()))
		So(err, ShouldNotBeNil)
		patch.Reset()
	})
}
