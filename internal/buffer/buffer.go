package buffer

import "sync"

// Len denotes the len of all buffers.
const Len = 1 << 15

// Dup returns a buffer which contains a copy of msg.
//
// the buffer may be released to the pool.
func Dup(msg []byte) (buf []byte) {
	buf = Get()
	copy(buf[:len(msg):Len], msg)

	return
}

// Get returns an available packet buffer.
func Get() []byte {
	return pool.Get().([]byte)
}

// Put releases the buffer
func Put(b []byte) {
	pool.Put(b[:Len:Len])
}

var pool = sync.Pool{
	New: func() interface{} {
		return make([]byte, Len, Len)
	},
}
