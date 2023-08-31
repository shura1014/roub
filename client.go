package roub

import (
	"bufio"
	"context"
	"github.com/shura1014/roub/rpclog"
	"io"
	"net"
	"time"
)

type MessageOption struct {
	CodecType
	CompressType
}

var DefaultOption = &MessageOption{
	CodecType:    Gob,
	CompressType: Gzip,
}

type RpcClient struct {
	*MessageOption
	conn           io.ReadWriteCloser // net.Conn
	buf            *bufio.Writer
	middlewares    Middlewares
	registry       Registry
	registryOption RegisterOption
	appName        string
}

func NewClient() *RpcClient {
	client := &RpcClient{
		MessageOption: DefaultOption,
	}

	return client
}

func (client *RpcClient) InitRegister(option RegisterOption) {
	client.registryOption = option
	client.registry = GetRegister(option)
}

func (client *RpcClient) Call(ctx context.Context, serviceName string, data ...any) *RpcResponse {
	// 服务发现
	service := ParseServiceName(serviceName)
	if service == nil {
		rpclog.Error("服务名格式错误 %s", serviceName)
	}

	var (
		instance string
		err      error
	)

	if client.appName == serviceName {
		// 本地服务 todo
		instance = "127.0.0.1:8888"
	} else {
		if client.registryOption.Mode == RegisterApplication {
			// 应用级注册
			instance, err = client.registry.Instance(service[0])
		} else {
			// 服务级别注册
			instance, err = client.registry.Instance(serviceName)
		}
		if err != nil {
			rpclog.Error(err)
			return nil
		}
	}

	conn, err := net.DialTimeout("tcp", instance, 5*time.Second)
	if err != nil {
		rpclog.Error(err)
		return nil
	}
	client.conn = conn
	client.buf = bufio.NewWriter(conn)
	// 注册中心获取连接
	req := &RpcRequest{}
	req.ReqId = reqId.Add(1)
	req.ServiceName = serviceName
	req.Data = data
	return client.handler(req)
}

func (client *RpcClient) handler(req *RpcRequest) *RpcResponse {
	invokeFunc := client.invoke
	for _, middleware := range client.middlewares {
		invokeFunc = middleware.Handler(invokeFunc)
	}

	return invokeFunc(req)
}

func (client *RpcClient) invoke(request *RpcRequest) *RpcResponse {
	defer func() {
		if err := recover(); err != nil {
			rpclog.Error(err)
		}
	}()
	body, err := Encoder(client.CodecType, request)
	assert(err)
	body, err = Compress(client.CompressType, body)
	assert(err)
	err = client.writeHeader(request.ReqId, len(body))
	assert(err)
	_, err = client.buf.Write(body)
	assert(err)

	err = client.buf.Flush()
	assert(err)
	respChan := make(chan *RpcResponse)
	go client.response(respChan)
	return <-respChan
}

func (client *RpcClient) response(resp chan *RpcResponse) {
	defer func() {
		if err := recover(); err != nil {
			if err != io.EOF {
				rpclog.Error(err)
				resp <- nil
			}
		}
	}()
	for {
		header, err := readHeader(client.conn)
		assert(err)
		body := make([]byte, header.body)
		_, err = io.ReadFull(client.conn, body)
		assert(err)
		body, err = UnCompress(header.compressType, body)
		assert(err)
		response := &RpcResponse{}
		err = Decoder(header.codecType, body, response)
		assert(err)
		resp <- response
	}
}

func (client *RpcClient) Close() error {
	return client.conn.Close()
}
