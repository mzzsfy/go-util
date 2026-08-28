//go:build (amd64 || arm64) && go1.23 && !go1.25

package unsafe

const goroutineIDOffset = 160
