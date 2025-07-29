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

import "testing"

func TestGenerateEtcdClient(t *testing.T) {
	testCases := []struct {
		inputConfig *Config
		hasErr      bool
	}{
		{
			inputConfig: &Config{},
			hasErr:      false,
		},
		{
			inputConfig: &Config{
				CertFile: "/noexisted.crt",
				KeyFile:  "/noexisted.key",
				CaFile:   "/noexisted.ca",
			},
			hasErr: true,
		},
		{
			inputConfig: &Config{
				CertFile: "./test-certs/tls.crt",
				KeyFile:  "./test-certs/tls.key",
				CaFile:   "./test-certs/ca.crt",
			},
			hasErr: false,
		},
	}

	for _, testCase := range testCases {
		_, err := GenerateEtcdClient(testCase.inputConfig)
		if err != nil && !testCase.hasErr {
			t.Errorf("GenerateEtcdClient() error = %v, wantErr %v", err, testCase.hasErr)
			return
		}
		if err == nil && testCase.hasErr {
			t.Errorf("GenerateEtcdClient() error = %v, wantErr %v", err, testCase.hasErr)
			return
		}
	}

}
