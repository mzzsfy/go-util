package helper

import "testing"

func Test_StringToBytes(t *testing.T) {
	bs := StringToBytes("test")
	if string(bs) != "test" {
		t.Errorf("expected %v, got %v", "test", string(bs))
	}
}

func Test_BytesToString(t *testing.T) {
	bytes := []byte("test")
	s := BytesToString(bytes)
	if s != "test" {
		t.Errorf("expected %v, got %v", "test", s)
	}
}

func Test_StringToBytes_Empty(t *testing.T) {
	bs := StringToBytes("")
	if len(bs) != 0 {
		t.Errorf("空串转换长度应为0, got %d", len(bs))
	}
}

func Test_BytesToString_Empty(t *testing.T) {
	s := BytesToString([]byte{})
	if s != "" {
		t.Errorf("空切片转换应为空串, got %q", s)
	}
}

func Test_StringBytes_RoundTrip(t *testing.T) {
	// 中英文混合往返内容一致, 验证无内存拷贝语义下内容无损
	for _, s := range []string{"abc", "中文测试", "mixed中英123"} {
		if got := BytesToString(StringToBytes(s)); got != s {
			t.Errorf("往返不一致: got %q, want %q", got, s)
		}
	}
}
