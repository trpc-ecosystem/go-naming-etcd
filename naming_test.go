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

package naming

import (
	"testing"

	"trpc.group/trpc-go/trpc-go"
	_ "trpc.group/trpc-go/trpc-go/http"

	. "github.com/glycerine/goconvey/convey"
)

func TestPlugin_Setup(t *testing.T) {
	Convey("测试配置文件加载", t, func() {
		s := trpc.NewServer()
		So(s, ShouldNotBeNil)
	})
}
