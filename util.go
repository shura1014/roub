package roub

import (
	"github.com/shura1014/roub/rpclog"
	"strings"
)

func assert(err error) {
	if err != nil {
		panic(err)
	}
}

func Println(err error) {
	if err != nil {
		rpclog.Error(err)
	}
}

func ParseServiceName(serviceName string) []string {
	split := strings.Split(serviceName, "/")
	if len(split) != 3 {
		return nil
	}
	return split
}
