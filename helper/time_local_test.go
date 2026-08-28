package helper

import (
	"testing"
	"time"
)

// Test_LocalTime_Scan 数据库扫描语义: 字符串/字节/原生时间可扫描, 其余类型报错
func Test_LocalTime_Scan(t *testing.T) {
	t.Parallel()
	t.Run("字符串", func(t *testing.T) {
		var lt LocalTime
		if err := lt.Scan("2022-12-31 23:59:59"); err != nil {
			t.Fatalf("Scan 字符串不应报错: %v", err)
		}
		if got := lt.Time(); got.Year() != 2022 || got.Month() != time.December || got.Day() != 31 {
			t.Errorf("Scan 字符串结果错误: %v", got)
		}
	})
	t.Run("字节切片", func(t *testing.T) {
		var lt LocalTime
		if err := lt.Scan([]byte("2022-12-31 23:59:59")); err != nil {
			t.Fatalf("Scan 字节切片不应报错: %v", err)
		}
		if got := lt.Time(); got.Year() != 2022 || got.Month() != time.December {
			t.Errorf("Scan 字节切片结果错误: %v", got)
		}
	})
	t.Run("原生时间", func(t *testing.T) {
		var lt LocalTime
		want := time.Date(2022, 12, 31, 23, 59, 59, 0, time.Local)
		if err := lt.Scan(want); err != nil {
			t.Fatalf("Scan 原生时间不应报错: %v", err)
		}
		if !lt.Time().Equal(want) {
			t.Errorf("Scan 原生时间结果错误: %v, 期望 %v", lt.Time(), want)
		}
	})
	t.Run("非法类型报错", func(t *testing.T) {
		var lt LocalTime
		if err := lt.Scan(123); err == nil {
			t.Error("Scan 非法类型应返回错误")
		}
	})
	t.Run("非法字符串报错", func(t *testing.T) {
		var lt LocalTime
		if err := lt.Scan("invalid"); err == nil {
			t.Error("Scan 非法字符串应返回错误")
		}
	})
}

// Test_LocalTime_TextAndBinary_Unmarshal 文本/二进制反序列化走 Parse 语义
func Test_LocalTime_TextAndBinary_Unmarshal(t *testing.T) {
	t.Parallel()
	t.Run("UnmarshalText合法", func(t *testing.T) {
		var lt LocalTime
		if err := lt.UnmarshalText([]byte("2022-12-31 23:59:59")); err != nil {
			t.Fatalf("UnmarshalText 不应报错: %v", err)
		}
		if lt.Time().Year() != 2022 {
			t.Errorf("UnmarshalText 结果错误: %v", lt.Time())
		}
	})
	t.Run("UnmarshalText非法", func(t *testing.T) {
		var lt LocalTime
		if err := lt.UnmarshalText([]byte("invalid")); err == nil {
			t.Error("UnmarshalText 非法输入应返回错误")
		}
	})
	t.Run("UnmarshalBinary合法", func(t *testing.T) {
		var lt LocalTime
		if err := lt.UnmarshalBinary([]byte("2022-12-31 23:59:59")); err != nil {
			t.Fatalf("UnmarshalBinary 不应报错: %v", err)
		}
		if lt.Time().Year() != 2022 {
			t.Errorf("UnmarshalBinary 结果错误: %v", lt.Time())
		}
	})
	t.Run("UnmarshalBinary非法", func(t *testing.T) {
		var lt LocalTime
		if err := lt.UnmarshalBinary([]byte("invalid")); err == nil {
			t.Error("UnmarshalBinary 非法输入应返回错误")
		}
	})
	t.Run("空JSON字符串报错", func(t *testing.T) {
		var lt LocalTime
		if err := lt.UnmarshalJSON([]byte(`""`)); err == nil {
			t.Error("空 JSON 字符串应返回错误")
		}
	})
}

// Test_LocalTime_Marshal_TextAndBinary 序列化输出与 String 一致
func Test_LocalTime_Marshal_TextAndBinary(t *testing.T) {
	t.Parallel()
	lt := LocalTime(time.Date(2022, 12, 31, 23, 59, 59, 0, time.Local))
	text, err := lt.MarshalText()
	if err != nil {
		t.Fatalf("MarshalText 不应报错: %v", err)
	}
	bin, err := lt.MarshalBinary()
	if err != nil {
		t.Fatalf("MarshalBinary 不应报错: %v", err)
	}
	want := "2022-12-31 23:59:59"
	if string(text) != want {
		t.Errorf("MarshalText = %q, 期望 %q", text, want)
	}
	if string(bin) != want {
		t.Errorf("MarshalBinary = %q, 期望 %q", bin, want)
	}
	if lt.String() != want {
		t.Errorf("String = %q, 期望 %q", lt.String(), want)
	}
}

// Test_LocalTime_MarshalUnmarshalRoundTrip 序列化后反序列化还原同一时刻
func Test_LocalTime_MarshalUnmarshalRoundTrip(t *testing.T) {
	t.Parallel()
	want := LocalTime(time.Date(2022, 12, 31, 23, 59, 59, 123456789, time.Local))
	data, err := want.MarshalBinary()
	if err != nil {
		t.Fatalf("MarshalBinary 不应报错: %v", err)
	}
	var got LocalTime
	// DateTimeLayout 不含纳秒, 往返后秒以下截断为相同表示
	if err := got.UnmarshalBinary(data); err != nil {
		t.Fatalf("UnmarshalBinary 不应报错: %v", err)
	}
	if got.String() != want.String() {
		t.Errorf("往返不一致: %v, 期望 %v", got.String(), want.String())
	}
}

// Test_LocalTime_StringWithLocal 指定时区格式化输出
func Test_LocalTime_StringWithLocal(t *testing.T) {
	t.Parallel()
	// 用 UTC 构造避免本地时区影响断言
	lt := LocalTime(time.Date(2022, 12, 31, 23, 59, 59, 0, time.UTC))
	if got := lt.StringWithLocal(time.UTC); got != "2022-12-31 23:59:59" {
		t.Errorf("StringWithLocal(UTC) = %q, 期望原样输出", got)
	}
	// 固定偏移时区: UTC 23:59:59 对应东八区次日 07:59:59
	east8 := time.FixedZone("UTC+8", 8*60*60)
	if got := lt.StringWithLocal(east8); got != "2023-01-01 07:59:59" {
		t.Errorf("StringWithLocal(东八区) = %q, 期望跨日换算结果", got)
	}
}

// Test_LocalTime_Time 转换为本地时区的同一时刻
func Test_LocalTime_Time(t *testing.T) {
	t.Parallel()
	utcTime := time.Date(2022, 12, 31, 15, 59, 59, 0, time.UTC)
	lt := LocalTime(utcTime)
	if !lt.Time().Equal(utcTime) {
		t.Errorf("Time() 应与原时刻相同: %v, 期望 %v", lt.Time(), utcTime)
	}
	if lt.Time().Location() != time.Local {
		t.Errorf("Time() 应返回本地时区, 实际 %v", lt.Time().Location())
	}
}

// Test_ParseLocalTimeWithLayout 自定义 layout 解析
func Test_ParseLocalTimeWithLayout(t *testing.T) {
	t.Parallel()
	t.Run("合法", func(t *testing.T) {
		got, err := ParseLocalTimeWithLayout("2006/01/02", "2022/12/31")
		if err != nil {
			t.Fatalf("合法输入不应报错: %v", err)
		}
		if got.Time().Year() != 2022 || got.Time().Month() != time.December {
			t.Errorf("解析结果错误: %v", got.Time())
		}
	})
	t.Run("非法", func(t *testing.T) {
		if _, err := ParseLocalTimeWithLayout("2006/01/02", "2022-12-31"); err == nil {
			t.Error("格式不匹配应返回错误")
		}
	})
}

// Test_LocalTime_Parse 短格式走 Auto 兜底
func Test_LocalTime_Parse(t *testing.T) {
	t.Parallel()
	var lt LocalTime
	if err := lt.Parse("2022-12-31"); err != nil {
		t.Fatalf("日期格式解析不应报错: %v", err)
	}
	if lt.Time().Day() != 31 {
		t.Errorf("Parse 结果错误: %v", lt.Time())
	}
	if err := lt.Parse("not-a-time"); err == nil {
		t.Error("非法输入应返回错误")
	}
}
