package helper

import (
    "testing"
)

func TestMd5(t *testing.T) {
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

func TestMd5Base64(t *testing.T) {
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

func TestBase64(t *testing.T) {
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

func TestDeBase64(t *testing.T) {
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

func TestBase64RoundTrip(t *testing.T) {
    tests := []string{"hello world", "测试base64", "", "!@#$%^&*()"}
    for _, input := range tests {
        encoded := Base64(input)
        decoded := DeBase64(encoded)
        if decoded != input {
            t.Errorf("Base64 round trip failed: input=%q, encoded=%q, decoded=%q", input, encoded, decoded)
        }
    }
}

func TestDeBase64Byte(t *testing.T) {
    got := DeBase64Byte("aGVsbG8=")
    if string(got) != "hello" {
        t.Errorf("DeBase64Byte = %q, want %q", string(got), "hello")
    }
}
