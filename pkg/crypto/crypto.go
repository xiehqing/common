package crypto

type Crypto interface {
	// Encrypt 加密
	Encrypt(data string) (string, error)
	// Decrypt 解密
	Decrypt(data string) ([]byte, error)
}

var (
	Base64Crypto = NewBase64()
	AesCrypto    = DefaultAes()
)
