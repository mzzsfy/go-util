package helper

import (
    "bytes"
    "fmt"
    "os"
    "runtime"
    "strings"
    "sync"
)

var (
    dunno     = []byte("???")
    centerDot = "·"
    slash     = "/"
)

// stackBufPool 复用 FormatStack 的 bytes.Buffer
var stackBufPool = sync.Pool{
    New: func() any {
        return new(bytes.Buffer)
    },
}

type Stack struct {
    PC   uintptr
    File string
    Line int
}

type Stacks []Stack

func (s Stacks) String() string {
    return FormatStack(s)
}

func CallerStackString(skip int, limit ...int) string {
    return FormatStack(CallerStack(skip, limit...))
}

// CallerStack 获取调用栈信息
// skip: 跳过的栈帧数(不包含本函数), limit: 最大返回帧数(0表示无限制)
func CallerStack(skip int, limit ...int) Stacks {
    l := 0
    if len(limit) > 0 {
        l = limit[0]
    }
    if skip < 0 {
        skip = 0
    }
    skip++ // 跳过 CallerStack 自身

    // 预分配缓冲区,避免多次调用 runtime.Caller
    const maxFrames = 64
    var pcs [maxFrames]uintptr
    n := runtime.Callers(skip, pcs[:])

    frames := runtime.CallersFrames(pcs[:n])
    stacks := make(Stacks, 0, n)
    for {
        frame, more := frames.Next()
        // 跳过 runtime 内部帧 (gopanic, goexit 等)
        if !strings.HasPrefix(frame.Function, "runtime.") {
            if l <= 0 || len(stacks) < l {
                stacks = append(stacks, Stack{
                    PC:   frame.PC,
                    File: frame.File,
                    Line: frame.Line,
                })
            }
        }
        if !more {
            break
        }
    }
    return stacks
}

// FormatStack 格式化调用栈为字符串
// 优化: 使用 sync.Pool 复用 buffer, 手动行解析避免 bytes.Split 分配
func FormatStack(stack []Stack) string {
    buf := stackBufPool.Get().(*bytes.Buffer)
    buf.Reset()
    defer stackBufPool.Put(buf)

    var data []byte
    var lastFile string
    for _, s := range stack {
        fmt.Fprintf(buf, "%s:%d\n", s.File, s.Line)
        if s.File != lastFile {
            var err error
            data, err = os.ReadFile(s.File)
            if err != nil {
                continue
            }
            lastFile = s.File
        }
        fmt.Fprintf(buf, "\t%s: %s\n", SimpleFunctionName(s.PC), sourceLine(data, s.Line))
    }
    return buf.String()
}

// sourceLine 从文件内容中找指定行的内容, 零分配版本
// 手动遍历找换行符, 避免 bytes.Split 创建大量切片
func sourceLine(data []byte, n int) []byte {
    if n <= 0 || len(data) == 0 {
        return dunno
    }

    // 找到第 n-1 个换行符后的位置
    lineStart := 0
    lineNum := 1
    for i := 0; i < len(data); i++ {
        if data[i] == '\n' {
            if lineNum == n {
                // 找到了第 n 行的结束位置 (i 是换行符位置)
                return bytes.TrimSpace(data[lineStart:i])
            }
            lineStart = i + 1
            lineNum++
        }
    }
    // 处理最后一行没有换行符的情况
    if lineNum == n && lineStart < len(data) {
        return bytes.TrimSpace(data[lineStart:])
    }
    return dunno
}

// SimpleFunctionName 返回简化的函数名(去除完整包路径和包名前缀)
// 例如: "github.com/mzzsfy/go-util/helper.Test.func1" -> "Test.func1"
func SimpleFunctionName(pc uintptr) string {
    name := FunctionName(pc)
    // 找最后一个 /, 去掉完整包路径前缀
    if lastSlash := strings.LastIndex(name, slash); lastSlash >= 0 {
        name = name[lastSlash+1:]
    }
    // 找第一个 ., 去掉包名 (如 "helper.")
    if period := strings.IndexByte(name, '.'); period >= 0 {
        name = name[period+1:]
    }
    // 替换 Unicode 中点 (Go 泛型方法使用)
    name = strings.ReplaceAll(name, centerDot, ".")
    return name
}

func FunctionName(pc uintptr) string {
    fn := runtime.FuncForPC(pc)
    if fn == nil {
        return ""
    }
    return fn.Name()
}
