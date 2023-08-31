package roub

import (
	"encoding/binary"
	"github.com/shura1014/roub/rpclog"
	"io"
)

// 魔数 1 版本 1 序列化类型 1 压缩类型 1 请求id(int64) 8 报文类型 1 数据长度 4 Data
var (
	magic   byte = 0x00 // 魔数字
	version byte = 0x1
)

type MsgType byte

const (
	// Request Response 消息类型
	Request  MsgType = iota
	Response MsgType = 2
)

type Header struct {
	magic        byte
	version      byte
	codecType    CodecType
	compressType CompressType
	reqId        int64
	msgType      MsgType
	body         uint32
}

func (client *RpcClient) writeHeader(reqId int64, s int) error {
	header := make([]byte, 17)
	header[0] = magic
	header[1] = version
	header[2] = byte(client.CodecType)
	header[3] = byte(client.CompressType)
	binary.BigEndian.PutUint64(header[4:12], uint64(reqId))
	header[12] = byte(Request)
	binary.BigEndian.PutUint32(header[13:], uint32(s))
	_, err := client.buf.Write(header[:])
	if err != nil {
		rpclog.Error(err)
		return err
	}
	return nil
}

func (server *RpcServer) writeHeader(reqHeader Header) []byte {
	header := make([]byte, 17)
	header[0] = magic
	header[1] = version
	header[2] = byte(reqHeader.codecType)
	header[3] = byte(reqHeader.compressType)
	binary.BigEndian.PutUint64(header[4:12], uint64(reqHeader.reqId))
	header[12] = byte(Response)
	return header
}

func readHeader(read io.ReadWriteCloser) (Header, error) {
	// 先读取头
	var h = Header{}
	headers := make([]byte, 17)
	_, err := io.ReadFull(read, headers)
	if err != nil {
		return h, err
	}
	h.magic = headers[0]
	h.version = headers[1]
	h.codecType = CodecType(headers[2])
	h.compressType = CompressType(headers[3])
	h.reqId = int64(binary.BigEndian.Uint64(headers[4:12]))
	h.msgType = MsgType(headers[12])
	h.body = binary.BigEndian.Uint32(headers[13:])
	return h, nil
}
