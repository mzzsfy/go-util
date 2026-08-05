package storage

import _ "unsafe"

//go:linkname fastrand runtime.fastrand
func fastrand() uint32
