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

// Package model 注册在etcd的节点信息
package model

import (
	"encoding/json"
	"fmt"
	"path"
	"strings"

	tregistry "trpc.group/trpc-go/trpc-go/naming/registry"
)

// Node 服务节点信息
type Node struct {
	Name     string            `json:"name"`     // 服务名
	ID       string            `json:"id"`       // id
	Address  string            `json:"address"`  // ip:port
	Metadata map[string]string `json:"metadata"` // 元数据
	Weight   int               `json:"weight"`   // 权重
}

// Marshal 序列化节点
func Marshal(node *Node) (string, error) {
	b, err := json.Marshal(node)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// Unmarshal 反序列化节点
func Unmarshal(b []byte) (*Node, error) {
	var node *Node
	if err := json.Unmarshal(b, &node); err != nil {
		return nil, err
	}
	return node, nil
}

// ConvertNode 将缓存在etcd中的节点转为trpc的节点
func ConvertNode(node *Node) *tregistry.Node {
	meta := make(map[string]interface{})
	for k, v := range node.Metadata {
		meta[k] = v
	}
	return &tregistry.Node{
		ServiceName: node.Name,
		Address:     node.Address,
		Metadata:    meta,
		Weight:      node.Weight,
	}
}

// NodePath 节点路径
func NodePath(prefix, service, id string) string {
	service = strings.Replace(service, "/", "-", -1)
	id = strings.Replace(id, "/", "-", -1)
	return path.Join(prefix, service, id)
}

// ServicePath 服务路径
func ServicePath(prefix, service string) string {
	if service == "" {
		return prefix
	}
	return path.Join(prefix, strings.Replace(service, "/", "-", -1), "/")
}

// ServiceID 构造生成service实例名 防止重名
func ServiceID(host, port, pid string) string {
	return fmt.Sprintf("%s-%s-%s", host, port, pid)
}
