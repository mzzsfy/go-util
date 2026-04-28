package helper

import (
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
    DateTimeLayout = "2006-01-02 15:04:05" //DateTimeLayout
)

// FormatDuration 格式化time.Duration 使其长度尽量为7位
func FormatDuration(duration time.Duration) time.Duration {
    if duration == 0 {
        return duration
    }
    // 记录符号, 取绝对值处理
    negative := duration < 0
    if negative {
        duration = -duration
    }
    d := duration
    var r time.Duration
    //11m11s
    if d > Duration10m {
        r = d.Round(Duration1s)
    } else
    //1m11.1s
    if d > Duration1m {
        r = d.Round(Duration01s)
    } else
    //1.111s
    if d > Duration10s {
        r = d.Round(Duration1ms)
    } else
    //1.1111s
    if d > Duration100ms {
        r = d.Round(Duration01ms)
    } else
    //11.111ms
    if d > Duration10ms {
        r = d.Round(Duration001ms)
    } else
    //11.111ms
    if d > Duration1ms {
        r = d.Round(Duration1us)
    } else
    //111.1us
    {
        r = d.Round(Duration01us)
    }
    //精度再高意义不大了,而且需要这种场景的一般不会使用这个工具
    if negative {
        r = -r
    }
    return r
}
