package codec

import (
	"bytes"
	"encoding/gob"
)

type GobCodec[V any] struct {
}

func NewGobCodec[V any]() *GobCodec[V] {
	return &GobCodec[V]{}
}

func (codec *GobCodec[V]) Encoder(data V) ([]byte, error) {
	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)
	if err := encoder.Encode(data); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func (codec *GobCodec[V]) Decoder(data []byte, msg V) error {
	buffer := bytes.NewBuffer(data)
	decoder := gob.NewDecoder(buffer)
	return decoder.Decode(msg)
}
