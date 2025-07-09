package bytebuffer

import "github.com/valyala/bytebufferpool"

// ByteBuffer is the alias of bytebufferpool.ByteBuffer.
type ByteBuffer = bytebufferpool.ByteBuffer

var (
	// Get returns an empty byte buffer from the pool, exported from znet/bytebuffer.
	Get = bytebufferpool.Get
	// Put returns byte buffer to the pool, exported from znet/bytebuffer.
	Put = func(b *ByteBuffer) {
		if b != nil {
			bytebufferpool.Put(b)
		}
	}
)
