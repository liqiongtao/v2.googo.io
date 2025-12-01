package goohttp

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"io"
)

// Encryptor 加密器接口
type Encryptor interface {
	Encrypt(plaintext []byte) ([]byte, error)
}

// Decryptor 解密器接口
type Decryptor interface {
	Decrypt(ciphertext []byte) ([]byte, error)
}

// AESGCMEncryptor AES-256-GCM 加密器
type AESGCMEncryptor struct {
	block cipher.Block
}

// Encrypt 加密数据
func (e *AESGCMEncryptor) Encrypt(plaintext []byte) ([]byte, error) {
	if e.block == nil {
		return nil, ErrInvalidEncryptionKey
	}

	// 创建GCM
	aesGCM, err := cipher.NewGCM(e.block)
	if err != nil {
		return nil, err
	}

	// 创建nonce
	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	// 加密
	ciphertext := aesGCM.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

// AESGCMDecryptor AES-256-GCM 解密器
type AESGCMDecryptor struct {
	block cipher.Block
}

// Decrypt 解密数据
func (d *AESGCMDecryptor) Decrypt(ciphertext []byte) ([]byte, error) {
	if d.block == nil {
		return nil, ErrInvalidEncryptionKey
	}

	// 创建GCM
	aesGCM, err := cipher.NewGCM(d.block)
	if err != nil {
		return nil, err
	}

	// 提取nonce
	nonceSize := aesGCM.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, ErrInvalidEncryptionKey
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	// 解密
	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

// EncryptBase64 加密并Base64编码
func EncryptBase64(encryptor Encryptor, plaintext []byte) (string, error) {
	ciphertext, err := encryptor.Encrypt(plaintext)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// DecryptBase64 Base64解码并解密
func DecryptBase64(decryptor Decryptor, encoded string) ([]byte, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, err
	}
	return decryptor.Decrypt(ciphertext)
}

