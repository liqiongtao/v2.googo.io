package goohttp

import (
	"bytes"
	"sync"
)

var bufferPool = sync.Pool{
	New: func() any {
		return &bytes.Buffer{}
	},
}

func getBuffer() *bytes.Buffer {
	return bufferPool.Get().(*bytes.Buffer)
}

func putBuffer(buf *bytes.Buffer) {
	// 重置Buffer, 避免数据残留
	buf.Reset()

	// 如果Buffer太大，不放入翅中（避免内存泄露）
	if buf.Cap() > 1*1024*1024 {
		return
	}

	bufferPool.Put(buf)
}
