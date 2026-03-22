package tools

import (
	"bytes"
	"crypto/md5"
	"hash"
	"sync"
)

// BufferPool is a pool of bytes.Buffer objects
type BufferPool struct {
	pool sync.Pool
}

// NewBufferPool creates a new BufferPool
func NewBufferPool() *BufferPool {
	return &BufferPool{
		pool: sync.Pool{
			New: func() interface{} {
				return &bytes.Buffer{}
			},
		},
	}
}

// Get gets a buffer from the pool
func (p *BufferPool) Get() *bytes.Buffer {
	return p.pool.Get().(*bytes.Buffer)
}

// Put puts a buffer back into the pool
func (p *BufferPool) Put(buf *bytes.Buffer) {
	buf.Reset()
	p.pool.Put(buf)
}

// MD5Pool is a pool of hash.Hash objects
type MD5Pool struct {
	pool sync.Pool
}

// NewMD5Pool creates a new MD5Pool
func NewMD5Pool() *MD5Pool {
	return &MD5Pool{
		pool: sync.Pool{
			New: func() interface{} {
				return md5.New()
			},
		},
	}
}

// Get gets an hash.Hash from the pool
func (p *MD5Pool) Get() hash.Hash {
	return p.pool.Get().(hash.Hash)
}

// Put puts an hash.Hash back into the pool
func (p *MD5Pool) Put(h hash.Hash) {
	h.Reset()
	p.pool.Put(h)
}

// Global pools
var (
	BufferPoolInstance = NewBufferPool()
	MD5PoolInstance    = NewMD5Pool()
)
