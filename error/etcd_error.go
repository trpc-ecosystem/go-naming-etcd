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

// Package error 插件返回的错误
package error

import "errors"

var (
	// ErrServerNotAvailable 服务不可用 没有可用节点
	ErrServerNotAvailable = errors.New("server can not available")
	// ErrBalancerNotExist 没有对应的负载均衡策略
	ErrBalancerNotExist = errors.New("load balancer is not exist")
)
