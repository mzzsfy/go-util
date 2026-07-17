package helper

import (
    "strings"
    "testing"
)

func Test_Md5(t *testing.T) {
    tests := []struct {
        name   string
        input  string
        expect string
    }{
        {"空字符串", "", "d41d8cd98f00b204e9800998ecf8427e"},
        {"hello", "hello", "5d41402abc4b2a76b9719d911017c592"},
        {"中文", "中文", "a7bac2239fcdcb3a067903d8077c4a07"},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := Md5(tt.input)
            if got != tt.expect {
                t.Errorf("Md5(%q) = %q, want %q", tt.input, got, tt.expect)
            }
        })
    }
}

func Test_Md5Base64(t *testing.T) {
    got := Md5Base64("hello")
    if got == "" {
        t.Error("Md5Base64 返回空字符串")
    }
    // 验证 base64 解码后与 Md5 一致
    gotHex := Md5("hello")
    gotBase64 := Md5Base64("hello")
    if gotBase64 == gotHex {
        t.Error("Md5Base64 不应返回 hex 格式")
    }
}

func Test_Base64(t *testing.T) {
    tests := []struct {
        name   string
        input  string
        expect string
    }{
        {"空字符串", "", ""},
        {"hello", "hello", "aGVsbG8="},
        {"二进制", "\x00\x01\x02", "AAEC"},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := Base64(tt.input)
            if got != tt.expect {
                t.Errorf("Base64(%q) = %q, want %q", tt.input, got, tt.expect)
            }
        })
    }
}

func Test_DeBase64(t *testing.T) {
    tests := []struct {
        name   string
        input  string
        expect string
    }{
        {"空字符串", "", ""},
        {"hello", "aGVsbG8=", "hello"},
        {"二进制", "AAEC", "\x00\x01\x02"},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := DeBase64(tt.input)
            if got != tt.expect {
                t.Errorf("DeBase64(%q) = %q, want %q", tt.input, got, tt.expect)
            }
        })
    }
}

func Test_Base64RoundTrip(t *testing.T) {
    tests := []string{"hello world", "测试base64", "", "!@#$%^&*()"}
    for _, input := range tests {
        encoded := Base64(input)
        decoded := DeBase64(encoded)
        if decoded != input {
            t.Errorf("Base64 round trip failed: input=%q, encoded=%q, decoded=%q", input, encoded, decoded)
        }
    }
}

func Test_DeBase64Byte(t *testing.T) {
    got := DeBase64Byte("aGVsbG8=")
    if string(got) != "hello" {
        t.Errorf("DeBase64Byte = %q, want %q", string(got), "hello")
    }
}

// BenchmarkMd5 测试 MD5 计算性能
func BenchmarkMd5(b *testing.B) {
    b.ReportAllocs()
    data := "benchmark test string for md5 hash"
    for i := 0; i < b.N; i++ {
        Md5(data)
    }
}

// BenchmarkMd5_Short 测试短字符串 MD5 计算性能
func BenchmarkMd5_Short(b *testing.B) {
    b.ReportAllocs()
    data := "hello"
    for i := 0; i < b.N; i++ {
        Md5(data)
    }
}

// BenchmarkMd5_Long 测试长字符串 MD5 计算性能
func BenchmarkMd5_Long(b *testing.B) {
    b.ReportAllocs()
    // 1KB 字符串
    data := strings.Repeat("a", 1024)
    for i := 0; i < b.N; i++ {
        Md5(data)
    }
}

// BenchmarkBase64 测试 Base64 编码性能
func BenchmarkBase64(b *testing.B) {
    b.ReportAllocs()
    data := "benchmark test string for base64 encode"
    for i := 0; i < b.N; i++ {
        Base64(data)
    }
}

// BenchmarkDeBase64 测试 Base64 解码性能
func BenchmarkDeBase64(b *testing.B) {
    b.ReportAllocs()
    data := "YmVuY2htYXJrIHRlc3Qgc3RyaW5nIGZvciBiYXNlNjQgZW5jb2Rl"
    for i := 0; i < b.N; i++ {
        DeBase64(data)
    }
}
