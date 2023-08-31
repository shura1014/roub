package roub

import (
	"github.com/shura1014/common/goerr"
	"github.com/shura1014/roub/internal/compress"
)

type CompressType byte

func (t CompressType) String() string {
	switch t {
	case Gzip:
		return "Gzip"
	default:
		return ""
	}
}

// Gzip 压缩类型
const (
	NONE CompressType = iota
	Gzip
)

var compressMap map[CompressType]compress.Compress

func init() {
	compressMap = make(map[CompressType]compress.Compress)
	RegisterCompress(Gzip, compress.NewGzip())
}

func RegisterCompress(compressType CompressType, compress compress.Compress) {
	compressMap[compressType] = compress
}

func Compress(compressType CompressType, body []byte) ([]byte, error) {
	if NONE == compressType {
		return body, nil
	}
	if c, ok := compressMap[compressType]; ok {
		return c.Compress(body)
	}
	return nil, goerr.Text("not find compressType %s", compressType)
}

func UnCompress(compressType CompressType, body []byte) ([]byte, error) {
	if NONE == compressType {
		return body, nil
	}
	if c, ok := compressMap[compressType]; ok {
		return c.UnCompress(body)
	}
	return nil, goerr.Text("not find compressType %s", compressType)
}
