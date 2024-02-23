//go:build concurrent_memory

package concurrent

type c struct {
    int64
    _ [24]byte
}
