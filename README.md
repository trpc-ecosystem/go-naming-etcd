# etcd名字服务注册中心

## 关于 etcd

见 [Etcd](https://etcd.io/)

## 示例

配置：

```yaml
plugins:
  registry:
    etcd:
      address: 127.0.0.1:2379,127.0.0.2:2379
      timeout: 5
      service:
        - name: trpc.test.helloworld.Greeter
          ttl: 10
          metadata:
            tags: helloworld
  selector:
    etcd:
      address: 127.0.0.1:2379,127.0.0.2:2379
      timeout: 5
      load_balance:
        name: round_robin

client: #客户端调用的后端配置
  service: #针对单个后端的配置
    - callee: trpc.test.helloworld.Greeter         #后端服务协议文件的service name, 如何callee和下面的name一样，那只需要配置一个即可
      target: etcd://trpc.test.helloworld.Greeter              #后端服务地址 etcd
      network: tcp                                 #后端服务的网络类型 tcp udp
      protocol: http                              #应用层协议 trpc http
      timeout: 10000                               #请求最长处理时间
      serialization: 2                             #序列化方式 0-pb 1-jce 2-json 3-flatbuffer，默认不要配置

```

使用证书访问etcd

```yaml
plugins:
  registry:
    etcd:
      address: 127.0.0.1:2379,127.0.0.2:2379
      timeout: 5
      tls:
        cafile: ./cert/etcd/ca.crt
        certfile: ./cert/etcd/tls.crt
        keyfile: ./cert/etcd/tls.key
      service:
        - name: trpc.test.helloworld.Greeter
          ttl: 10
          metadata:
            tags: helloworld
  selector:
    etcd:
      address: 127.0.0.1:2379,127.0.0.2:2379
      timeout: 5
      tls:
        cafile: ./cert/etcd/ca.crt
        certfile: ./cert/etcd/tls.crt
        keyfile: ./cert/etcd/tls.key
      load_balance:
        name: round_robin
```

## 服务寻址
```go
package main

import (
	_ "trpc.group/trpc-go/trpc-naming-etcd"
	_ "trpc.group/trpc-go/trpc-naming-etcd/registry"
)

func main() {
	clientProxy := pb.NewGreeterClientProxy()
	req := &pb.HelloRequest{
		Msg: "hello",
	}

	rsp, err := clientProxy.SayHello(ctx, req)
	if err != nil {
		log.Error(err.Error())
		return
	}

	log.Info("req:%v, rsp:%v, err:%v", req, rsp, err)
}

```