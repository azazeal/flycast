package buffer

import "sync"

// Size denotes the size of all buffers.
const Size = 1 << 14

// Dup returns a buffer which contains a copy of msg.
//
// the buffer may be released to the pool.
func Dup(msg []byte) (buf []byte) {
	buf = Get()
	copy(buf[:len(msg):Size], msg)

	return
}

// Get returns an available packet buffer.
func Get() []byte {
	return pool.Get().([]byte)
}

// Put releases the buffer
func Put(b []byte) {
	pool.Put(b[:Size:Size])
}

var pool = sync.Pool{
	New: func() interface{} {
		return make([]byte, Size, Size)
	},
}
