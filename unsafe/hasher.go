package unsafe

import "unsafe"

type Hasher[K comparable] interface {
    Hash(key K) uint64
    NewSeed() Hasher[K]
    WithSeed(uintptr) Hasher[K]
}

// Hasher hashes values of type K.
// Uses runtime AES-based hashing.
type hasher[K comparable] struct {
    hash hashFn
    seed uintptr
}

// NewHasher creates a new Hasher[K] with a random seed.
func NewHasher[K comparable]() Hasher[K] {
    return &hasher[K]{
        hash: getRuntimeHasher[K](),
        seed: hashSeed,
    }
}

// Hash hashes |key|.
func (h hasher[K]) Hash(key K) uint64 {
    // promise to the compiler that pointer
    // |p| does not escape the stack.
    p := noescape(unsafe.Pointer(&key))
    return uint64(h.hash(p, h.seed))
}

func (h hasher[K]) NewSeed() Hasher[K] {
    h1 := NewHasher[K]()
    h1.(*hasher[K]).seed = newHashSeed()
    return h1
}

func (h hasher[K]) WithSeed(seed uintptr) Hasher[K] {
    return &hasher[K]{
        hash: h.hash,
        seed: seed,
    }
}
