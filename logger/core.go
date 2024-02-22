package logger

import (
    "github.com/mzzsfy/go-util/helper"
    "io"
    "os"
    "strings"
)

const (
    // TraceLevel 记录细节,参数等信息
    TraceLevel Level = iota
    // DebugLevel 记录重要的参数等信息
    DebugLevel
    // InfoLevel 记录信息用户易读信息
    InfoLevel
    // WarnLevel 记录用户需要关注信息
    WarnLevel
    // ErrorLevel 记录用户需要处理的信息
    ErrorLevel
    // FatalLevel 致命错误,一般影响到系统正常运行
    FatalLevel
)

type Level int8

func (l *Level) String() string {
    switch *l {
    //info和debug通常使用的次数会多一些
    case InfoLevel:
        return "I"
    case DebugLevel:
        return "D"
    case TraceLevel:
        return "T"
    case WarnLevel:
        return "W"
    case ErrorLevel:
        return "E"
    case FatalLevel:
        return "F"
    default:
        if *l > FatalLevel {
            return "F"
        } else {
            return "T"
        }
    }
}

func (l *Level) UnmarshalJSON(b []byte) error {
    *l = FormString(strings.TrimSuffix(helper.BytesToString(b[1:]), `"`))
    return nil
}

func (l *Level) UnmarshalBinary(b []byte) error {
    *l = FormString(helper.BytesToString(b))
    return nil
}

func (l *Level) UnmarshalText(b []byte) error {
    *l = FormString(helper.BytesToString(b))
    return nil
}

func (l *Level) MarshalJSON() ([]byte, error) {
    return helper.StringToBytes(`"` + l.String() + `"`), nil
}
func (l *Level) MarshalYAML() (any, error) {
    return l.String(), nil
}
func (l *Level) MarshalBinary() ([]byte, error) {
    return helper.StringToBytes(l.String()), nil
}
func (l *Level) MarshalText() ([]byte, error) {
    return helper.StringToBytes(l.String()), nil
}

func FormString(s string) Level {
    //info和debug通常使用的次数会多一些
    if s == "" {
        return InfoLevel
    }
    b := s[0]
    if b == 'I' || b == 'i' {
        return InfoLevel
    }
    if b == 'D' || b == 'd' {
        return DebugLevel
    }
    if b == 'T' || b == 't' {
        return TraceLevel
    }
    if b == 'W' || b == 'w' {
        return WarnLevel
    }
    if b == 'E' || b == 'e' {
        return ErrorLevel
    }
    if b == 'F' || b == 'f' {
        return FatalLevel
    }
    return InfoLevel
}

var (
    defaultLevel = func() *Level {
        level := InfoLevel
        return &level
    }()
    writerTarget io.Writer = os.Stdout
)

func DefaultWriterTarget() io.Writer {
    return writerTarget
}

func SetDefaultWriterTarget(w io.Writer) {
    writerTarget = w
}

func SetDefaultLogLevel(l Level) {
    *defaultLevel = l
}
func DefaultLogLevel() Level {
    return *defaultLevel
}
