package roub

import (
	"encoding/gob"
	"github.com/shura1014/common/goerr"
	"github.com/shura1014/common/type/atom"
)

type Map map[string]any

func init() {
	gob.Register(Map{})
	gob.Register(map[string]any{})
}

var reqId = atom.NewInt64()

type RpcRequest struct {
	ReqId       int64
	ServiceName string
	Data        []any
}

type RpcResponse struct {
	ReqId int64
	Err   *goerr.BizError
	Data  []any
}
