package logger

import (
	"math/rand"
	"testing"
	"time"
)

// Given appendDuration 复刻 time.Duration.String
// When 遍历边界值与随机值域
// Then 输出与标准库逐字节一致
func Test_AppendDuration_MatchesStdlib(t *testing.T) {
	cases := []time.Duration{
		0, 1, 9, 10, 99, 100, 999, 1000, 1001, 1500, 9999,
		time.Microsecond - 1, time.Microsecond, time.Microsecond + 1,
		time.Millisecond - 1, time.Millisecond, time.Millisecond + 1,
		time.Second - 1, time.Second, time.Second + 1,
		59 * time.Second, time.Minute - 1, time.Minute, time.Minute + 1,
		time.Hour - time.Second, time.Hour, time.Hour + time.Minute,
		25 * time.Hour, 2540400 * time.Hour, // 最大可表示时长
		-1, -time.Second, -time.Minute - time.Millisecond, -2540400 * time.Hour,
	}
	rnd := rand.New(rand.NewSource(1))
	for i := 0; i < 10000; i++ {
		cases = append(cases, time.Duration(rnd.Int63n(int64(time.Hour)))*time.Duration(rnd.Intn(1000)))
	}
	// 全量程随机, 含负数
	for i := 0; i < 10000; i++ {
		cases = append(cases, time.Duration(rnd.Int63()))
		cases = append(cases, -time.Duration(rnd.Int63()))
	}
	for _, d := range cases {
		got := string(appendDuration(nil, d))
		want := d.String()
		if got != want {
			t.Errorf("appendDuration(%d) = %q, want %q", int64(d), got, want)
		}
	}
}
