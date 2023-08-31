package roub

import (
	"encoding/binary"
	"github.com/shura1014/common/goerr"
	"github.com/shura1014/common/utils/reflectutil"
	"github.com/shura1014/roub/rpclog"
	"io"
	"net"
	"reflect"
)

// 收到请求  filter处理

const (
	networkTcp     = "tcp"
	networkHttp    = "http"
	defaultAppName = "default"
	defaultAddress = ":0"
)

type RpcServer struct {
	listen         net.Listener
	network        string
	address        string
	appName        string
	enableRegistry bool
	registry       Registry
	registryOption RegisterOption
	services       map[string]*Service
}

func NewTcpServer(address ...string) *RpcServer {
	addr := defaultAddress
	if len(address) > 0 {
		addr = address[0]
	}
	return NewServer(networkTcp, addr)
}

func NewHttpServer(address ...string) *RpcServer {
	addr := defaultAddress
	if len(address) > 0 {
		addr = address[0]
	}
	return NewServer(networkHttp, addr)
}

func NewServer(network, address string) *RpcServer {
	return &RpcServer{
		address:  address,
		network:  network,
		appName:  defaultAppName,
		services: map[string]*Service{},
	}
}

func (server *RpcServer) SetAppName(appName string) {
	server.appName = appName
}

func (server *RpcServer) InitRegister(option RegisterOption) {
	server.registryOption = option
	server.registry = GetRegister(option)
	server.enableRegistry = true
}

func (server *RpcServer) Run(tasks ...func(server *RpcServer)) {
	server.listen, _ = net.Listen(server.network, server.address)
	rpclog.Info(server.listen.Addr())
	if server.enableRegistry && server.registry != nil {
		// 应用级注册
		if server.registryOption.Mode == RegisterApplication {
			err := server.registry.Registry(server.appName, server.listen.Addr().String())
			if err != nil {
				rpclog.Error("服务注册失败：%+v", err)

			} else {
				rpclog.Info("服务注册成功 %s", server.appName)
			}
		}
	}
	for _, task := range tasks {
		task(server)
	}
	for {
		conn, err := server.listen.Accept()
		if err != nil {
			rpclog.Error(err)
			continue
		}
		// 收到连接
		go server.handlerAccept(conn)
	}
}

func (server *RpcServer) handlerAccept(conn io.ReadWriteCloser) {
	defer func() {
		if err := recover(); err != nil {
			if err != io.EOF {
				rpclog.Error(err)
			}
		}
		err := conn.Close()
		if err != nil {
			rpclog.Error(err)
		}
	}()
	header, err := readHeader(conn)
	assert(err)
	body := make([]byte, header.body)
	_, err = io.ReadFull(conn, body)
	assert(err)
	body, err = UnCompress(header.compressType, body)
	assert(err)
	request := &RpcRequest{}
	err = Decoder(header.codecType, body, request)
	assert(err)
	server.doExecute(request, header, conn)
}

func (server *RpcServer) doExecute(request *RpcRequest, header Header, write io.ReadWriteCloser) {
	rpclog.Info(request)

	serviceName := request.ServiceName
	names := ParseServiceName(serviceName)
	service := server.services[names[1]]
	method := service.methods[names[2]]
	args := make([]reflect.Value, len(request.Data))

	for i := range request.Data {
		args[i] = reflect.ValueOf(request.Data[i])
	}
	results, err := server.Execute(method, args)
	resp := &RpcResponse{
		ReqId: header.reqId,
	}
	if err != nil {
		// 异常
		rpclog.Error("服务 %s 调用失败  %+v", serviceName, err)
		resp.Err = goerr.Wrap(err)
	}

	if len(results) == 0 {
		// 成功
		rpclog.Info("服务 %s 执行成功，无返回值", serviceName)
	}
	resp.Data = results

	server.response(header, resp, write)
}

func (server *RpcServer) response(header Header, rpcResponse *RpcResponse, write io.ReadWriteCloser) {
	writeHeader := server.writeHeader(header)

	body, err := Encoder(header.codecType, rpcResponse)
	assert(err)
	body, err = Compress(header.compressType, body)
	assert(err)
	binary.BigEndian.PutUint32(writeHeader[13:], uint32(len(body)))
	_, err = write.Write(writeHeader[:])
	assert(err)
	_, err = write.Write(body)
	assert(err)
}

func (server *RpcServer) Execute(method *Method, args []reflect.Value) ([]any, error) {

	results := method.method.Call(args)

	resultValue := make([]interface{}, len(results))
	for i := 0; i < len(results); i++ {
		resultValue[i] = results[i].Interface()
	}

	if len(results) > 0 {
		err, ok := resultValue[len(resultValue)-1].(error)
		if ok {
			return nil, err
		}
	}
	return resultValue, nil
}

func (server *RpcServer) RegisterService(s any) {
	serviceStructValue := reflect.ValueOf(s)
	serviceStructType := reflect.TypeOf(s)
	serviceStructName := reflectutil.GetName(s)

	service := &Service{
		name:         serviceStructName,
		serviceValue: serviceStructValue,
		serviceType:  serviceStructType,
		methods:      map[string]*Method{},
	}
	server.services[serviceStructName] = service
	service.RegisterMethods()
}
