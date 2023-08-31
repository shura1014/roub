package codec

type Codec[V any] interface {
	Encoder(data V) ([]byte, error)
	Decoder(data []byte, msg any) error
}
