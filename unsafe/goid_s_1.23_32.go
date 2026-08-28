//go:build (386 || arm || amd64p32) && go1.23 && !go1.25

package unsafe

// 32 位指针宽, g.goid 偏移按 go1.23~1.24 runtime2.go 计算
const goroutineIDOffset = 84
