package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"github.com/pkg/errors"
)

var DefaultAesSalt = "www.zorktech.com"

type Aes struct {
	Salt string
}

func NewAes(salt string) *Aes {
	return &Aes{Salt: salt}
}

func DefaultAes() *Aes {
	return &Aes{
		Salt: DefaultAesSalt,
	}
}

// ZeroPadding 添加Zero填充
func ZeroPadding(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	if padding == 0 {
		padding = blockSize
	}
	padText := make([]byte, padding)
	return append(data, padText...)
}

// ZeroUnPadding 移除Zero填充
func ZeroUnPadding(data []byte) []byte {
	length := len(data)
	if length == 0 {
		return data
	}
	// 从末尾开始移除连续的0字节
	for i := length - 1; i >= 0; i-- {
		if data[i] != 0 {
			return data[:i+1]
		}
	}
	// 如果全部都是0，返回空数组
	return []byte{}
}

// Encrypt 加密
func (a *Aes) Encrypt(plaintextOrigin string) (string, error) {
	if a.Salt == "" {
		a.Salt = DefaultAesSalt
	}
	// 将字符串转换为字节数组
	keyBytes := []byte(a.Salt)
	dataBytes := []byte(plaintextOrigin)
	// 检查密钥长度，必须是16, 24, 或32字节
	keyLen := len(keyBytes)
	if keyLen != 16 && keyLen != 24 && keyLen != 32 {
		return "", errors.New("密钥长度必须是16、24或32字节")
	}
	// 创建AES cipher
	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		return "", err
	}
	// Zero填充到块大小的倍数
	paddedData := ZeroPadding(dataBytes, aes.BlockSize)
	// 使用相同的key作为IV（为了兼容前端代码）
	iv := keyBytes[:aes.BlockSize] // 取密钥的前16字节作为IV
	// 创建CBC模式加密器
	mode := cipher.NewCBCEncrypter(block, iv)
	// 加密数据
	ciphertext := make([]byte, len(paddedData))
	mode.CryptBlocks(ciphertext, paddedData)
	// 返回base64编码的结果（兼容前端.toString()）
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt 解密
func (a *Aes) Decrypt(encryptedData string) ([]byte, error) {
	if a.Salt == "" {
		a.Salt = DefaultAesSalt
	}
	// 将字符串转换为字节数组
	keyBytes := []byte(a.Salt)
	// 检查密钥长度
	keyLen := len(keyBytes)
	if keyLen != 16 && keyLen != 24 && keyLen != 32 {
		return nil, errors.New("密钥长度必须是16、24或32字节")
	}
	// base64解码
	ciphertext, err := base64.StdEncoding.DecodeString(encryptedData)
	if err != nil {
		return nil, err
	}
	// 检查密文长度
	if len(ciphertext)%aes.BlockSize != 0 {
		return nil, errors.New("密文长度不是块大小的倍数")
	}
	// 创建AES cipher
	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		return nil, err
	}
	// 使用相同的key作为IV
	iv := keyBytes[:aes.BlockSize]
	// 创建CBC模式解密器
	mode := cipher.NewCBCDecrypter(block, iv)
	// 解密数据
	plaintext := make([]byte, len(ciphertext))
	mode.CryptBlocks(plaintext, ciphertext)
	// 移除Zero填充
	unPaddedData := ZeroUnPadding(plaintext)
	return unPaddedData, nil
}

// Decrypt 解密
func Decrypt(ciphertext, salt string) (string, error) {
	if salt == "" {
		salt = DefaultAesSalt
	}
	a := &Aes{
		Salt: salt,
	}
	data, err := a.Decrypt(ciphertext)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
