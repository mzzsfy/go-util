//go:build go1.18

package unsafe

import (
    "unsafe"
)

// go/src/runtime/type.go
type maptype struct {
    typ    _type
    key    *_type
    elem   *_type
    bucket *_type
    // function for hashing keys (ptr to key, seed) -> hash
    hasher     func(unsafe.Pointer, uintptr) uintptr
    keysize    uint8
    elemsize   uint8
    bucketsize uint16
    flags      uint32
}

// go/src/runtime/type.go
type tflag uint8
type nameOff int32
type typeOff int32

// go/src/runtime/type.go
type _type struct {
    size       uintptr
    ptrdata    uintptr
    hash       uint32
    tflag      tflag
    align      uint8
    fieldAlign uint8
    kind       uint8
    equal      func(unsafe.Pointer, unsafe.Pointer) bool
    gcdata     *byte
    str        nameOff
    ptrToThis  typeOff
}
