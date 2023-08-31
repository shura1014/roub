package roub

type InvokeFunc func(*RpcRequest) *RpcResponse

type Middlewares []Middleware
type Middleware struct {
	Order   int
	Handler func(InvokeFunc) InvokeFunc
}

func (m Middlewares) Len() int {
	return len(m)
}

func (m Middlewares) Swap(i, j int) {
	m[i], m[j] = m[j], m[i]
}

// Less 比较，将order大的排到前面，实际上是最后执行
func (m Middlewares) Less(i, j int) bool {
	return m[i].Order < m[j].Order
}
