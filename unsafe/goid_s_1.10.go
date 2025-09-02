//go:build (arm64 || arm || amd64 || amd64p32 || 386) && go1.10 && !go1.23

package unsafe

const goroutineIDOffset = 152
