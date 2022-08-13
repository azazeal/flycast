package buffer

import "sync"

// Size denotes the size of all buffers.
const Size = 1 << 14

type Buffer [Size]byte

// Get returns an available packet buffer.
func Get() *Buffer {
	return pool.Get().(*Buffer)
}

// Put releases the buffer
func Put(b *Buffer) {
	pool.Put(b)
}

var pool = sync.Pool{
	New: func() interface{} {
		return new(Buffer)
	},
}
