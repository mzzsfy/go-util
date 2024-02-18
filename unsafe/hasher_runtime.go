package unsafe

import (
    "math/rand"
    "unsafe"
)

type hashFn func(unsafe.Pointer, uintptr) uintptr

func getRuntimeHasher[K comparable]() (h hashFn) {
    a := any(make(map[K]struct{}))
    //offsetof := unsafe.Offsetof(maptype{}.bucket)
    //size1 := unsafe.Sizeof(_type{})
    //fmt.Println(size1, offsetof)
    //hash := ***(***_type)(unsafe.Pointer(uintptr(unsafe.Pointer(&a)) + offsetof))
    //fmt.Printf("%v\n", hash)
    h = (**(**maptype)(unsafe.Pointer(&a))).hasher
    return
}

//nolint:gosec
var hashSeed = rand.Int()

func newHashSeed() uintptr {
    return uintptr(hashSeed)
}

// noescape hides a pointer from escape analysis. It is the identity function
// but escape analysis doesn't think the output depends on the input.
// noescape is inlined and currently compiles down to zero instructions.
// USE CAREFULLY!
// This was copied from the runtime (via pkg "strings"); see issues 23382 and 7921.
//
//go:nosplit
//go:nocheckptr
func noescape(p unsafe.Pointer) unsafe.Pointer {
    x := uintptr(p)
    return unsafe.Pointer(x ^ 0)
}

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
