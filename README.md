# roub

简易的rpc框架，适用于学习适用

# 快速使用

## 单实例服务

> 定义一个服务

```go
type TestService struct {
}

func (test *TestService) Test(data map[string]any) map[string]any {
	data["info"] = "服务端返回"
	return data
}
```

> 启动注册中心


> 启动服务

```go
func TestServer(t *testing.T) {
	server := NewTcpServer()
	server.InitRegister(RegisterOption{
		Mode:         RegisterApplication,
		Address:      "127.0.0.1:9999",
		RegisterType: MemRegistryType,
	})
	server.SetAppName("upc")
	server.Run(func(server *RpcServer) {
		server.RegisterService(&TestService{})
	})
}
```

```text
# 服务日志
[roub] 2023-08-31 11:33:14 shura/roub/server.go:70 INFO [::]:56983 
[roub] 2023-08-31 11:33:14 shura/roub/server.go:79 INFO 服务注册成功 upc 
[roub] 2023-08-31 11:33:14 shura/roub/service.go:35 DEBUG 注册服务 TestService_Test 

注册中心日志
[registry] 2023-08-31 11:33:14 shura/registry/mem.go:52 INFO register /upc &[{[::]:56983}] 
```

> 客户端访问

```go
func TestClient(t *testing.T) {
	resp := testClient.Call(context.TODO(), "upc/TestService/Test", map[string]any{"name": "shura"})
	t.Log(resp)
	resp = testClient.Call(context.TODO(), "upc/TestService/Test", map[string]any{"name": "shura"})
	t.Log(resp)
	resp = testClient.Call(context.TODO(), "upc/TestService/Test", map[string]any{"name": "shura"})
	t.Log(resp)
	resp = testClient.Call(context.TODO(), "upc/TestService/Test", map[string]any{"name": "shura"})
	t.Log(resp)
}
```

```text
=== RUN   TestClient
    client_test.go:21: &{ReqId:1 Err:<nil> Data:[map[info:服务端返回 name:shura]]}
    client_test.go:23: &{ReqId:2 Err:<nil> Data:[map[info:服务端返回 name:shura]]}
    client_test.go:25: &{ReqId:3 Err:<nil> Data:[map[info:服务端返回 name:shura]]}
    client_test.go:27: &{ReqId:4 Err:<nil> Data:[map[info:服务端返回 name:shura]]}
--- PASS: TestClient (0.01s)
PASS
```

> 压测

一个注册中心、一个服务实例、一个客户端

```go
func BenchmarkTest(b *testing.B) {
	for i := 0; i < b.N; i++ {
		testClient.Call(context.TODO(), "upc/TestService/Test", map[string]any{"name": "shura"})
	}
}


goos: darwin
goarch: arm64
pkg: github.com/shura1014/roub
BenchmarkTest
BenchmarkTest-8   	    2280	    590852 ns/op
PASS
```

## 多实例服务

> 再注册一个服务
```text
=== RUN   TestServer2
[roub] 2023-08-31 11:36:10 shura/roub/server.go:70 INFO [::]:57367 
[roub] 2023-08-31 11:36:10 shura/roub/server.go:79 INFO 服务注册成功 upc 
[roub] 2023-08-31 11:36:10 shura/roub/service.go:35 DEBUG 注册服务 TestService_Test 

[registry] 2023-08-31 11:36:10 shura/registry/mem.go:52 INFO register /upc &[{[::]:56983} {[::]:57367}] 
```
> 客户端访问

服务1日志

```text
[roub] 2023-08-31 11:38:06 shura/roub/server.go:123 INFO &{ReqId:1 ServiceName:upc/TestService/Test Data:[map[name:shura]]} 
[roub] 2023-08-31 11:38:06 shura/roub/server.go:123 INFO &{ReqId:3 ServiceName:upc/TestService/Test Data:[map[name:shura]]}

```

服务2日志

```text
[roub] 2023-08-31 11:38:06 shura/roub/server.go:123 INFO &{ReqId:2 ServiceName:upc/TestService/Test Data:[map[name:shura]]} 
[roub] 2023-08-31 11:38:06 shura/roub/server.go:123 INFO &{ReqId:4 ServiceName:upc/TestService/Test Data:[map[name:shura]]}
```