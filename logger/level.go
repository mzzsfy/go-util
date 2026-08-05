package logger

// Level 日志级别
type Level int8

const (
	TraceLevel Level = iota
	DebugLevel
	InfoLevel
	WarnLevel
	ErrorLevel
	FatalLevel
)

// LevelUnset 哨兵值: 清除本地级别设置, 继承父级
const LevelUnset Level = -1

// levelShort 单字符标识, 索引对应 Level
var levelShort = [6]byte{'T', 'D', 'I', 'W', 'E', 'F'}

// levelName 全名, 索引对应 Level
var levelName = [6]string{"Trace", "Debug", "Info", "Warn", "Error", "Fatal"}

// tag 返回级别的单字符标识
func (l Level) tag() byte {
	if l >= 0 && int(l) < len(levelShort) {
		return levelShort[l]
	}
	return 'I'
}

// String 返回级别全名
func (l Level) String() string {
	if l >= 0 && int(l) < len(levelName) {
		return levelName[l]
	}
	return "Info"
}

// FullName 返回级别全名, LevelUnset 返回 "Unset"
func (l Level) FullName() string {
	if l == LevelUnset {
		return "Unset"
	}
	return l.String()
}

// FromString 从字符串解析级别, 大小写不敏感, 空串返回 InfoLevel
func FromString(s string) Level {
	if s == "" {
		return InfoLevel
	}
	switch s[0] | 0x20 {
	case 't':
		return TraceLevel
	case 'd':
		return DebugLevel
	case 'i':
		return InfoLevel
	case 'w':
		return WarnLevel
	case 'e':
		return ErrorLevel
	case 'f':
		return FatalLevel
	default:
		return InfoLevel
	}
}
