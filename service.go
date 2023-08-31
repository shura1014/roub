package roub

import (
	"github.com/shura1014/common/type/atom"
	"github.com/shura1014/roub/rpclog"
	"reflect"
)

type Service struct {
	name         string
	serviceValue reflect.Value
	serviceType  reflect.Type
	methods      map[string]*Method
}

type Method struct {
	method    reflect.Value
	callCount *atom.Int64
	name      string
}

func NewMethod(name string, method reflect.Value) *Method {
	return &Method{
		method:    method,
		name:      name,
		callCount: atom.NewInt64(),
	}
}

func (s *Service) RegisterMethods() {
	for i := 0; i < s.serviceValue.NumMethod(); i++ {
		method := s.serviceValue.Method(i)
		name := s.serviceType.Method(i).Name
		s.methods[name] = NewMethod(name, method)
		rpclog.Debug("注册服务 %s_%s", s.name, name)
	}
}
