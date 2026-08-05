package unsafe

import (
	"reflect"
	"sync"
	"unsafe"
)

type Hasher[K comparable] interface {
	Hash(key K) uint64
	NewSeed() Hasher[K]
	WithSeed(uintptr) Hasher[K]
}

// hasher[K] 使用 runtime AES 哈希, 相同 K 类型的实例完全一致(函数指针+全局种子)
type hasher[K comparable] struct {
	hash hashFn
	seed uintptr
}

// hasherCache 按 K 类型缓存默认 hasher, 避免每次 NewHasher 都堆分配
var hasherCache sync.Map // reflect.Type(*K) -> Hasher[K]

func NewHasher[K comparable]() Hasher[K] {
	rt := reflect.TypeOf((*K)(nil))
	if v, ok := hasherCache.Load(rt); ok {
		return v.(Hasher[K])
	}
	h := &hasher[K]{
		hash: getRuntimeHasher[K](),
		seed: hashSeed,
	}
	actual, _ := hasherCache.LoadOrStore(rt, h)
	return actual.(Hasher[K])
}

// Hash hashes |key|.
func (h hasher[K]) Hash(key K) uint64 {
	p := noescape(unsafe.Pointer(&key))
	return uint64(h.hash(p, h.seed))
}

func (h hasher[K]) NewSeed() Hasher[K] {
	return &hasher[K]{
		hash: h.hash,
		seed: newHashSeed(),
	}
}

func (h hasher[K]) WithSeed(seed uintptr) Hasher[K] {
	return &hasher[K]{
		hash: h.hash,
		seed: seed,
	}
}
