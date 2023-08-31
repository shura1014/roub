package roub

import (
	"context"
	"testing"
)

var testClient *RpcClient

func init() {
	testClient = NewClient()
	testClient.InitRegister(RegisterOption{
		Mode:         RegisterApplication,
		Address:      "127.0.0.1:9999",
		RegisterType: MemRegistryType,
	})
}

func TestClient(t *testing.T) {
	resp := testClient.Call(context.TODO(), "upc/TestService/Test", map[string]any{"name": "shura"})
	t.Logf("%+v", resp)
	resp = testClient.Call(context.TODO(), "upc/TestService/Test", map[string]any{"name": "shura"})
	t.Logf("%+v", resp)
	resp = testClient.Call(context.TODO(), "upc/TestService/Test", map[string]any{"name": "shura"})
	t.Logf("%+v", resp)
	resp = testClient.Call(context.TODO(), "upc/TestService/Test", map[string]any{"name": "shura"})
	t.Logf("%+v", resp)
}

func BenchmarkTest(b *testing.B) {
	for i := 0; i < b.N; i++ {
		testClient.Call(context.TODO(), "upc/TestService/Test", map[string]any{"name": "shura"})
	}
}
