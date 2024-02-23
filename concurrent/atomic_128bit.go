//go:build concurrent_fast

package concurrent

type c struct {
    int64
    _ [120]byte
}
