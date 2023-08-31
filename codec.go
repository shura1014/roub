package roub

import (
	"github.com/shura1014/common/goerr"
	"github.com/shura1014/roub/internal/codec"
)

type CodecType byte

func (t CodecType) String() string {
	switch t {
	case Gob:
		return "Gob"
	default:
		return ""
	}
}

// Gob 序列化类型
const (
	Gob CodecType = iota
)

var codecMap map[CodecType]codec.Codec[any]

func init() {
	codecMap = make(map[CodecType]codec.Codec[any])
	RegisterCodec(Gob, codec.NewGobCodec[any]())
}

func RegisterCodec(codecType CodecType, codec codec.Codec[any]) {
	codecMap[codecType] = codec
}

func Encoder(codeType CodecType, data any) ([]byte, error) {
	if c, ok := codecMap[codeType]; ok {
		return c.Encoder(data)
	}
	return nil, goerr.Text("not find codecType %s", codeType)
}

func Decoder(codeType CodecType, data []byte, msg any) error {
	if c, ok := codecMap[codeType]; ok {
		return c.Decoder(data, msg)
	}
	return goerr.Text("not find codecType %b", codeType)
}
