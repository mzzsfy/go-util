package helper

import (
    "errors"
    "fmt"
    "strings"
    "time"
)

type LocalTime time.Time

func (t *LocalTime) UnmarshalJSON(bytes []byte) error {
    if len(bytes) == 0 {
        return fmt.Errorf("empty input")
    }
    if bytes[0] == '"' {
        bytes = bytes[1 : len(bytes)-1]
    }
    return t.Parse(BytesToString(bytes))
}

func (t *LocalTime) UnmarshalBinary(b []byte) error {
    return t.Parse(string(b))
}

func (t *LocalTime) UnmarshalText(b []byte) error {
    return t.Parse(string(b))
}

func (t *LocalTime) MarshalBinary() ([]byte, error) {
    return []byte(t.String()), nil
}

func (t *LocalTime) MarshalText() ([]byte, error) {
    return []byte(t.String()), nil
}
func (t *LocalTime) StringWithLocal(location *time.Location) string {
    return t.Time().In(location).Format(DateTimeLayout)
}
func (t *LocalTime) String() string {
    return t.Time().Local().Format(DateTimeLayout)
}
func (t *LocalTime) Time() time.Time {
    return time.Time(*t).Local()
}
func (t *LocalTime) Scan(value any) error {
    switch v := value.(type) {
    case string:
        auto, err := ParseLocalTimeAuto(v)
        if err != nil {
            return err
        }
        *t = auto
    case []byte:
        auto, err := ParseLocalTimeAuto(string(v))
        if err != nil {
            return err
        }
        *t = auto
    case time.Time:
        *t = LocalTime(v)
    default:
        return fmt.Errorf("unsupported type %T", value)
    }
    return nil
}

func (t *LocalTime) Parse(str string) error {
    var err error
    *t, err = ParseLocalTime(str)
    return err
}

// ParseLocalTime 只能使用 DateTimeLayout 格式
func ParseLocalTime(str string) (LocalTime, error) {
    if len(str) != len(DateTimeLayout) {
        return ParseLocalTimeAuto(str)
    }
    return ParseLocalTimeWithLayout(DateTimeLayout, str)
}

// ParseLocalTimeWithLayout 使用 layout 格式解析
func ParseLocalTimeWithLayout(layout, str string) (LocalTime, error) {
    parse, err := time.ParseInLocation(layout, str, time.Local)
    return LocalTime(parse), err
}

// ParseLocalTimeAuto 自动匹配常见格式,只支持数字格式,支持常见中文
func ParseLocalTimeAuto(str string) (LocalTime, error) {
    //可优化 go对 time.RFC3339 格式有性能提升
    //纯数字格式
    if len(str) <= len("yyyyMMddHHmmssSSSSSSSSS") && StringAllIsNumber(str) {
        switch len(str) {
        case len("yyyyMMddHHmmss"):
            return ParseLocalTimeWithLayout(`20060102150405`, str)
        case len("yyyyMMddHHmm"):
            return ParseLocalTimeWithLayout(`200601021504`, str)
        case len("yyyyMMddHH"):
            return ParseLocalTimeWithLayout(`2006010215`, str)
        case len("yyyyMMdd"):
            return ParseLocalTimeWithLayout(`20060102`, str)
        case len("HHmmss"):
            return ParseLocalTimeWithLayout(`150405`, str)
        }
        if len(str) >= len("yyyyMMddHHmmssSS") {
            return ParseLocalTimeWithLayout("20060102150405.999999999", str[:len("20060102150405")]+"."+str[len("20060102150405"):])
        }
    } else {
        localTime, err := parse2(str)
        if err == nil {
            return localTime, err
        }
    }
    if len(str) > len("yyyy-MM-dd HH:mm:ss.SS -0700 MST") {
        i := strings.Index(str, " m=")
        if i != -1 {
            return ParseLocalTimeWithLayout("2006-01-02 15:04:05.999999999 -0700 MST", str[:i])
        }
        return ParseLocalTimeWithLayout("2006-01-02 15:04:05.999999999 -0700 MST", str)
    }
    if strings.Contains(str, "年") {
        return ParseLocalTimeAuto(removeChinese(str))
    }
    if len(str) >= len("yyyy-MM-dd HH:mm:ss.SS") {
        if str[4] == '-' && str[10] == ' ' && str[19] == '.' {
            return ParseLocalTimeWithLayout(`2006-01-02 15:04:05.999999999`, str)
        }
    }
    return LocalTime(time.Time{}), fmt.Errorf("无法解析时间: %s", str)
}

func parse2(str string) (LocalTime, error) {
    switch len(str) {
    case len("yyyy-MM-ddTHH:mm:ssZ07:00"):
        parse, err := time.Parse(time.RFC3339, str)
        return LocalTime(parse), err
    case len("yyyy-MM-dd'T'HH:mm:ss'Z'"):
        return ParseLocalTimeWithLayout(`2006-01-02'T'15:04:05'Z'`, str)
    case len("yyyy-MM-dd'T'HH:mm:ss"):
        return ParseLocalTimeWithLayout(`2006-01-02'T'15:04:05`, str)
    case len("yyyy-MM-ddTHH:mm:ssZ"):
        return ParseLocalTimeWithLayout(`2006-01-02T15:04:05Z`, str)
    case len("yyyy-MM-dd HH:mm:ss.SSS"):
        return ParseLocalTimeWithLayout(`2006-01-02 15:04:05.000`, str)
    case len("yyyy-MM-dd HH:mm:ss"):
        //return ParseLocalTimeWithLayout(`2006`+str[4:5]+`01`+str[7:8]+`02`+str[10:11]+`15`+str[13:14]+`04`+str[16:17]+`05`, str)
        if str[11] == 'T' {
            return ParseLocalTimeWithLayout(`2006-01-02T15:04:05`, str)
        }
        if str[11] == ' ' {
            if str[4] == '-' {
                return ParseLocalTimeWithLayout(`2006-01-02 15:04:05`, str)
            } else if str[4] == '/' {
                return ParseLocalTimeWithLayout(`2006/01/02 15:04:05`, str)
            }
        }
        return ParseLocalTimeWithLayout(`20060102150405`, str[0:4]+str[5:7]+str[8:10]+str[11:13]+str[14:16]+str[17:19])
    case len("yyyy-MM-dd HH:mm"):
        return ParseLocalTimeWithLayout(`2006-01-02 15:04`, str)
    case len("yyyy-MM-dd"):
        if str[4] == '/' {
            return ParseLocalTimeWithLayout(`2006/01/02`, str)
        }
        return ParseLocalTimeWithLayout(`2006-01-02`, str)
    case len("HH:mm:ss"):
        return ParseLocalTimeWithLayout(`15:04:05`, str)
    case len("HH:mm:ss.SSS"):
        return ParseLocalTimeWithLayout(`15:04:05.000`, str)
    }
    return LocalTime{}, errors.New("无法解析")
}

// removeChinese 单次遍历替换中文时间字符为标准格式
func removeChinese(str string) string {
    var b strings.Builder
    b.Grow(len(str))
    for _, r := range str {
        switch r {
        case '年', '月':
            b.WriteByte('-')
        case '日':
            b.WriteByte(' ')
        case '点', '时', '分':
            b.WriteByte(':')
        case '秒':
            // 移除
        default:
            b.WriteRune(r)
        }
    }
    return b.String()
}
