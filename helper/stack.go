package helper

import (
    "bytes"
    "fmt"
    "os"
    "runtime"
)

var (
    dunno     = []byte("???")
    centerDot = []byte("·")
    dot       = []byte(".")
    slash     = []byte("/")
)

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
func CallerStack(skip int, limit ...int) Stacks {
    var stacks []Stack
    l := 0
    if len(limit) > 0 {
        l = limit[0]
    }
    if skip < 0 {
        skip = 0
    }
    skip += 1
    for i := skip; ; i++ { // Skip the expected number of frames
        pc, file, line, ok := runtime.Caller(i)
        if !ok {
            break
        }
        if l > 0 && i-skip >= l {
            break
        }
        stacks = append(stacks, Stack{
            PC:   pc,
            File: file,
            Line: line,
        })
    }
    return stacks
}

func FormatStack(stack []Stack) string {
    buf := new(bytes.Buffer)
    var lines [][]byte
    var lastFile string
    for _, s := range stack {
        fmt.Fprintf(buf, "%s:%d\n", s.File, s.Line)
        if s.File != lastFile {
            data, err := os.ReadFile(s.File)
            if err != nil {
                continue
            }
            lines = bytes.Split(data, []byte{'\n'})
            lastFile = s.File
        }
        fmt.Fprintf(buf, "\t%s: %s\n", SimpleFunctionName(s.PC), source(lines, s.Line))
    }
    return buf.String()
}

func SimpleFunctionName(pc uintptr) string {
    name := []byte(FunctionName(pc))
    if lastSlash := bytes.LastIndex(name, slash); lastSlash >= 0 {
        name = name[lastSlash+1:]
    }
    if period := bytes.Index(name, dot); period >= 0 {
        name = name[period+1:]
    }
    name = bytes.Replace(name, centerDot, dot, -1)
    return string(name)
}

func FunctionName(pc uintptr) string {
    fn := runtime.FuncForPC(pc)
    if fn == nil {
        return ""
    }
    return fn.Name()
}

func source(lines [][]byte, n int) []byte {
    n--
    if n < 0 || n >= len(lines) {
        return dunno
    }
    return bytes.TrimSpace(lines[n])
}
