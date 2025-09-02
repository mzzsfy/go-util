package storage

import "github.com/mzzsfy/go-util/unsafe"

func NewDefaultHasher[K comparable]() unsafe.Hasher[K] {
    return unsafe.NewHasher[K]()
}
