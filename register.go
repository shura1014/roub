package roub

import (
	"encoding/json"
	"github.com/shura1014/balance"
	"github.com/shura1014/common/container/concurrent"
	"github.com/shura1014/common/goerr"
	"github.com/shura1014/httpclient"
	"github.com/shura1014/registry"
	"github.com/shura1014/roub/rpclog"
	"net/http"
	"sync"
	"time"
)

type Registry interface {
	Init(option RegisterOption)
	Registry(service string, addr string) error
	Instance(service string) (string, error)
}

type registryMod int
type RegistryType int

const (
	// RegisterApplication 默认就用应用级注册
	RegisterApplication = 1
	RegistryMethod      = 2 // Not implemented
	MemRegistryType     = 1
	ZkType              = 2 // Not implemented
	NacosType           = 3 // Not implemented
)

var once sync.Once

func GetRegister(option RegisterOption) Registry {
	var r Registry
	switch option.RegisterType {
	case MemRegistryType:
		r = NewMemRegistry()
	default:
		r = NewMemRegistry()
	}
	r.Init(option)
	return r
}

type RegisterOption struct {
	Mode         registryMod
	Address      string
	RegisterType RegistryType
}

var unavailable = goerr.Text("Registry Service unavailable")

var (
	registerPath = "/mem/register"
	discoverPath = "/mem/discover"
	probePath    = "/mem/probe"
	monitorPath  = "/mem/monitor"
)

type MemRegistry struct {
	client *httpclient.Client
	addr   string // 暂定一个
	alive  bool

	appInfo *concurrent.Map[string, balance.Balance[*registry.Address]]

	RegisterOption
}

func NewMemRegistry() *MemRegistry {
	client := httpclient.NewClient()
	// 服务端长轮询为30秒，客户端超时时间应该大于此时间
	client.SetTimeout(35 * time.Second)
	client.SetHeader(httpclient.ContentType, httpclient.ApplicationJson)
	memRegister := MemRegistry{
		client:  client,
		appInfo: concurrent.NewMap[string, balance.Balance[*registry.Address]](),
	}

	return &memRegister
}

func (server *MemRegistry) Init(option RegisterOption) {
	server.RegisterOption = option
	server.Health()
}

func (server *MemRegistry) Registry(service string, addr string) error {
	if !server.alive {
		return unavailable
	}
	data := &registry.Data{
		Key:   service,
		Value: addr,
	}
	resp, err := server.client.Post("http://"+server.RegisterOption.Address+registerPath, data)
	if resp.StatusCode() == http.StatusOK {
		return nil
	}

	return goerr.Wrapf(err, "注册失败")
}

func (server *MemRegistry) Instance(service string) (string, error) {
	if !server.alive {
		return "", unavailable
	}
	value := server.appInfo.Get(service)

	if value == nil {
		instance := server.discover(service)
		if instance == "" {
			return "", goerr.Text("No instances of the application %s are available", service)
		}
		return instance, nil
	}
	instance, err := value.Instance()
	if err != nil {
		return "", err
	}

	return instance.Addr(), nil
}

func (server *MemRegistry) discover(appName string) string {
	value := server.appInfo.Get(appName)
	if value != nil {
		instance, err := value.Instance()
		if err != nil {
			return ""
		}
		return instance.Addr()
	}
	//	todo 并发安全
	resp, err := server.client.Get("http://"+server.RegisterOption.Address+discoverPath, registry.Data{Key: appName})
	Println(err)

	if resp.StatusCode() == http.StatusOK {
		m := make([]*registry.Address, 0)
		err := json.Unmarshal(resp.GetBody(), &m)
		if err != nil {
			rpclog.Error(err)

		}
		b := balance.GetBalance[*registry.Address](balance.RoundRobin_)
		b.InitNodes(m...)
		server.appInfo.Put(appName, b)
		once.Do(func() {
			go server.monitor()
		})
		instance, _ := b.Instance()
		return instance.Addr()
	}

	return ""

}

func (server *MemRegistry) monitor() {
	defer func() {
		if err := recover(); err != nil {
			rpclog.Error(err)
		}
	}()

	for {
		var keys []string
		server.appInfo.Iterator(func(key string, value balance.Balance[*registry.Address]) bool {
			keys = append(keys, key)
			return true
		})

		resp, _ := server.client.Post("http://"+server.RegisterOption.Address+monitorPath, keys)
		result := make(map[string]any)
		_ = json.Unmarshal(resp.GetBody(), &result)
		if result["code"] == http.StatusNotModified {
			continue
		}

		if result["code"] == http.StatusOK {
			if result["appName"] != "" {
				appName := result["appName"].(string)
				if result["Data"] == nil {
					server.appInfo.Remove(appName)
					continue
				}
				addresses := result["Data"].([]*registry.Address)
				server.appInfo.Get(appName).InitNodes(addresses...)
			}
		}
	}
}

func (server *MemRegistry) Health() {
	resp, err := http.Get("http://" + server.RegisterOption.Address + probePath)
	if err != nil {
		rpclog.Error(err)
	}
	if resp.StatusCode == http.StatusOK {
		server.alive = true
		return
	}
	server.alive = false
}
