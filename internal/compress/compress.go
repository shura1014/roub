package compress

type Compress interface {
	Compress(body []byte) ([]byte, error)
	UnCompress(body []byte) ([]byte, error)
}
