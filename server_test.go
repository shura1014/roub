package roub

import "testing"

type TestService struct {
}

func (test *TestService) Test(data map[string]any) map[string]any {
	data["info"] = "服务端返回"
	return data
}

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

func TestServer2(t *testing.T) {
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
