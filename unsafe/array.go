package unsafe

import (
    "sync/atomic"
    "unsafe"
)

type Array[T any] struct {
    keyShifts uintptr        // Pointer size - log2 of array size, to be used as arr in the data array
    count     atomic.Uintptr // count of filled elements in the slice
    array     unsafe.Pointer // pointer to slice data array
    arr       []T            // storage for the slice for the garbage collector to not clean it up
}

// item returns the item for the given hashed key.
func (s *Array[T]) get(index int) T {
}

func (s *Array[T]) set(t T) {

}
func (s *Array[T]) push(t T) {
    
}
