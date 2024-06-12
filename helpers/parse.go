package helpers

import (
	"encoding/hex"
	"encoding/json"
	"github.com/neurlang/goruut/helpers/log"
)

// ParseJson parses JSON data into the specified type.
func ParseJson[T any](data []byte) (result T, err error) {
	err = json.Unmarshal(data, &result)
	return
}

// ParsePtrJson parses JSON data into the specified pointer type.
func ParsePtrJson[T any](data []byte) (result *T, err error) {
	result = new(T)
	err = json.Unmarshal(data, result)
	if err == nil {
		return
	}
	return nil, err
}

// SerializeJson serializes data into a JSON byte array.
func SerializeJson[T any](data T) (result []byte, err error) {
	return json.Marshal(data)
}

// ParseHash parses a hash string into a byte array.
func ParseHash(str string) (out [32]byte) {
	data := log.Error1(hex.DecodeString(str))
	copy(out[:], data)
	return
}
