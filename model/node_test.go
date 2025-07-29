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

package model

import "testing"

func Test_NodePath(t *testing.T) {
	type args struct {
		prefix  string
		service string
		id      string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "normal",
			args: args{
				prefix:  "prefix",
				service: "service",
				id:      "id",
			},
			want: "prefix/service/id",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NodePath(tt.args.prefix, tt.args.service, tt.args.id); got != tt.want {
				t.Errorf("nodePath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_ServiceId(t *testing.T) {
	type args struct {
		service string
		host    string
		port    string
		pid     string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "normal",
			args: args{
				host: "host",
				port: "port",
				pid:  "pid",
			},
			want: "host-port-pid",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ServiceID(tt.args.host, tt.args.port, tt.args.pid); got != tt.want {
				t.Errorf("serviceId() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_ServicePath(t *testing.T) {
	type args struct {
		prefix  string
		service string
	}
	tests := []struct {
		name string
		args args
		want string
	}{

		{
			name: "normal",
			args: args{
				prefix:  "prefix",
				service: "service",
			},
			want: "prefix/service",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ServicePath(tt.args.prefix, tt.args.service); got != tt.want {
				t.Errorf("servicePath() = %v, want %v", got, tt.want)
			}
		})
	}
}
