package crypto

import (
	"encoding/base64"
	"fmt"
)

type Base64 struct {
}

func NewBase64() *Base64 {
	return &Base64{}
}

func (b *Base64) Encrypt(data string) (string, error) {
	return base64.StdEncoding.EncodeToString([]byte(data)), nil
}

func (b *Base64) Decrypt(data string) ([]byte, error) {
	decodedBytes, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64 string: %v", err)
	}
	return decodedBytes, nil
}
