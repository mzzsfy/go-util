package util

import (
    "encoding/json"
    "fmt"
    "strings"
    "time"
)

const (
    Duration10m    = time.Minute * 10
    Duration1m     = time.Minute
    Duration10s    = time.Second * 10
    Duration1s     = time.Second
    Duration100ms  = time.Millisecond * 100
    Duration01s    = Duration100ms
    Duration10ms   = time.Millisecond * 10
    Duration1ms    = time.Millisecond
    Duration100us  = time.Microsecond * 100
    Duration01ms   = Duration100us
    Duration10us   = time.Microsecond * 10
    Duration001ms  = Duration10us
    Duration1us    = time.Microsecond
    Duration01us   = time.Nanosecond * 100
    Duration001us  = time.Nanosecond * 10
    DateTimeLayout = "2006-01-02 15:04:05"
)

type LocalTime time.Time

func (t *LocalTime) UnmarshalJSON(bytes []byte) error {
    return t.Parse(strings.Trim(string(bytes), `"`))
}

func (t *LocalTime) UnmarshalBinary(b []byte) error {
    return t.Parse(string(b))
}

func (t *LocalTime) UnmarshalText(b []byte) error {
    return t.Parse(string(b))
}

func (t *LocalTime) MarshalJSON() ([]byte, error) {
    return json.Marshal(t.String())
}

func (t *LocalTime) MarshalYAML() (any, error) {
    return t.String(), nil
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
func (t *LocalTime) Parse(str string) error {
    var err error
    *t, err = ParseLocalTime(str)
    return err
}

// ParseLocalTime 只能使用 DateTimeLayout 格式
func ParseLocalTime(str string) (LocalTime, error) {
    return ParseLocalTimeWithLayout(DateTimeLayout, str)
}
func ParseLocalTimeWithLayout(layout, str string) (LocalTime, error) {
    parse, err := time.ParseInLocation(layout, str, time.Local)
    return LocalTime(parse), err
}

// ParseLocalTimeAuto 自动匹配常见格式,只支持数字格式
func ParseLocalTimeAuto(str string) (LocalTime, error) {
    str = strings.TrimSpace(str)
    if IsInteger(str) {
        switch len(str) {
        case len("yyyyMMddHHmmss"):
            return ParseLocalTimeWithLayout(`20060102150405`, str)
        case len("yyyyMMddHHmmssSSS"):
            return ParseLocalTimeWithLayout(`20060102150405000`, str)
        case len("yyyyMMdd"):
            return ParseLocalTimeWithLayout(`20060102`, str)
        case len("HHmmss"):
            return ParseLocalTimeWithLayout(`150405`, str)
        }
    } else {
        switch len(str) {
        case len("yyyy-MM-ddTHH:mm:ssZ07:00"):
            parse, err := time.Parse(time.RFC3339, str)
            return LocalTime(parse), err
        case len("yyyy-MM-dd'T'HH:mm:ss'Z'"):
            return ParseLocalTimeWithLayout(`2006-01-02'T'15:04:05'Z'`, str)
        case len("yyyy-MM-ddTHH:mm:ssZ"):
            return ParseLocalTimeWithLayout(`2006-01-02T15:04:05Z`, str)
        case len("yyyy-MM-dd HH:mm:ss.SSS"):
            return ParseLocalTimeWithLayout(`2006-01-02 15:04:05.000`, str)
        case len("yyyy-MM-dd HH:mm:ss"):
            return ParseLocalTimeWithLayout(`2006`+str[4:5]+`01`+str[7:8]+`02`+str[10:11]+`15`+str[13:14]+`04`+str[16:17]+`05`, str)
        case len("yyyy-MM-dd HH:mm"):
            return ParseLocalTimeWithLayout(`2006-01-02 15:04`, str)
        case len("yyyy-MM-dd"):
            return ParseLocalTimeWithLayout(`2006-01-02`, str)
        case len("HH:mm:ss"):
            return ParseLocalTimeWithLayout(`15:04:05`, str)
        case len("HH:mm:ss.SSS"):
            return ParseLocalTimeWithLayout(`15:04:05.000`, str)
        }
    }
    return LocalTime(time.Time{}), fmt.Errorf("无法解析时间: %s", str)
}

// FormatDuration 格式化time.Duration 使其长度尽量为7位
func FormatDuration(duration time.Duration) time.Duration {
    d := duration
    if d == 0 {
        return d
    }
    //11m11s
    if d > Duration10m {
        return d.Round(Duration1s)
    } else
    //1m11.1s
    if d > Duration1m {
        return d.Round(Duration01s)
    } else
    //1.111s
    if d > Duration10s {
        return d.Round(Duration1ms)
    } else
    //1.1111s
    if d > Duration100ms {
        return d.Round(Duration01ms)
    } else
    //11.111ms
    if d > Duration10ms {
        return d.Round(Duration001ms)
    } else
    //11.111ms
    if d > Duration1ms {
        return d.Round(Duration1us)
    } else
    //111.1µs
    {
        return d.Round(Duration01us)
    }
    //精度再高意义不大了,而且需要这种场景的一般不会使用这个工具
}
