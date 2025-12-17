package goohttp

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
)

var (
	ErrEncryptFailed     = errors.New("加密失败")
	ErrDecryptFailed     = errors.New("解密失败")
	ErrInvalidKey        = errors.New("无效的密钥")
	ErrInvalidCiphertext = errors.New("无效的密文")
)

type Encryptor interface {
	Encrypt(plaintext []byte) ([]byte, error)
	Decrypt(ciphertext []byte) ([]byte, error)
	GetCipher() cipher.AEAD
}

type AES256GCMEncryptor struct {
	aead cipher.AEAD
	key  []byte
	mu   sync.RWMutex
}

func NewAES256GCMEncryptor(key []byte) (*AES256GCMEncryptor, error) {
	if len(key) != 32 {
		return nil, ErrInvalidKey
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	return &AES256GCMEncryptor{
		aead: aead,
		key:  key,
	}, nil
}

func (e *AES256GCMEncryptor) Encrypt(plaintext []byte) ([]byte, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	nonce := make([]byte, e.aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	ciphertext := e.aead.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

func (e *AES256GCMEncryptor) Decrypt(ciphertext []byte) ([]byte, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	nonceSize := e.aead.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, ErrInvalidCiphertext
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := e.aead.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDecryptFailed, err)
	}

	return plaintext, nil
}

func (e *AES256GCMEncryptor) GetCipher() cipher.AEAD {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.aead
}

func (e *AES256GCMEncryptor) SetKey(key []byte) error {
	if len(key) != 32 {
		return ErrInvalidKey
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return fmt.Errorf("failed to create cipher: %w", err)
	}

	aead, err := cipher.NewGCM(block)
	if err != nil {
		return fmt.Errorf("failed to create GCM: %w", err)
	}

	e.mu.Lock()
	defer e.mu.Unlock()
	e.aead = aead
	e.key = key

	return nil
}

// 加密响应写入器
type encryptResponseWriter struct {
	gin.ResponseWriter
	encryptor     Encryptor
	buffer        *bytes.Buffer
	mu            sync.Mutex
	headerWritten bool
	statusCode    int
}

func (w *encryptResponseWriter) Write(data []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	// 写入到缓冲区
	return w.buffer.Write(data)
}

func (w *encryptResponseWriter) WriteHeader(statusCode int) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.statusCode = statusCode
	w.headerWritten = true
}

func (w *encryptResponseWriter) flush() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.buffer.Len() == 0 {
		return nil
	}

	// 如果还没有写入状态码，使用默认值
	if !w.headerWritten {
		w.statusCode = 200
		w.ResponseWriter.WriteHeader(w.statusCode)
	} else {
		w.ResponseWriter.WriteHeader(w.statusCode)
	}

	// 加密数据
	data := w.buffer.Bytes()
	encrypted, err := w.encryptor.Encrypt(data)
	if err != nil {
		return err
	}

	// 写入加密后的数据
	_, err = w.ResponseWriter.Write(encrypted)
	return err
}

func (w *encryptResponseWriter) release() {
	if w.buffer != nil {
		putBuffer(w.buffer)
		w.buffer = nil
	}
}

func EncryptMiddleware(encryptor Encryptor) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 解密请求体
		if c.Request.Body != nil && c.Request.ContentLength > 0 {
			buf := getBuffer()
			defer putBuffer(buf)

			if _, err := io.Copy(buf, c.Request.Body); err != nil {
				ErrorWithStatus(&Context{Context: c}, http.StatusBadRequest, 4001, "获取请求数据失败")
				return
			}

			decrypted, err := encryptor.Decrypt(buf.Bytes())
			if err != nil {
				ErrorWithStatus(&Context{Context: c}, http.StatusBadRequest, 4002, "获取请求数据失败")
				return
			}

			c.Request.Body = io.NopCloser(bytes.NewReader(decrypted))
			c.Request.ContentLength = int64(len(decrypted))
		}

		// 包装响应写入器以加密响应
		writer := &encryptResponseWriter{
			ResponseWriter: c.Writer,
			encryptor:      encryptor,
			buffer:         getBuffer(),
		}
		c.Writer = writer

		c.Next()

		// 刷新加密数据
		if err := writer.flush(); err != nil {
			// 如果刷新失败，记录错误但不中断请求
			// todo:: 记录日志
		}

		// 释放资源
		writer.release()
	}
}
