package helper

import (
    "strings"
    "testing"
)

// TestTruncate 验证 Truncate 截断逻辑
func TestTruncate(t *testing.T) {
    tests := []struct {
        name   string
        input  string
        toLen  int
        right  bool
        expect string
    }{
        {"无需截断", "abc", 5, false, "abc"},
        {"恰好等长", "abc", 3, false, "abc"},
        {"左截断", "abcdef", 3, false, "abc"},
        {"右截断", "abcdef", 3, true, "def"},
        {"中文字符", "你好世界测试", 3, false, "你好世"},
        {"中文字符右截断", "你好世界测试", 3, true, "界测试"},
        {"空字符串", "", 3, false, ""},
        {"零长度", "abc", 0, false, ""},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := Truncate(tt.input, tt.toLen, tt.right)
            if got != tt.expect {
                t.Errorf("Truncate(%q, %d, %v) = %q, want %q", tt.input, tt.toLen, tt.right, got, tt.expect)
            }
        })
    }
}

// TestPadding 验证 Padding 填充逻辑
func TestPadding(t *testing.T) {
    tests := []struct {
        name   string
        input  string
        toLen  int
        right  bool
        expect string
    }{
        {"无需填充", "abc", 3, false, "abc"},
        {"右填充", "abc", 5, false, "abc  "},
        {"左填充", "abc", 5, true, "  abc"},
        {"已经超长", "abcdef", 3, false, "abcdef"},
        {"中文右填充", "你好", 4, false, "你好  "},
        {"中文左填充", "你好", 4, true, "  你好"},
        {"空字符串填充", "", 3, false, "   "},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := Padding(tt.input, tt.toLen, tt.right)
            if got != tt.expect {
                t.Errorf("Padding(%q, %d, %v) = %q, want %q", tt.input, tt.toLen, tt.right, got, tt.expect)
            }
        })
    }
}

// TestPaddingOrTruncate 验证 PaddingOrTruncate 组合逻辑
func TestPaddingOrTruncate(t *testing.T) {
    tests := []struct {
        name   string
        input  string
        toLen  int
        right  bool
        expect string
    }{
        {"截断", "abcdef", 3, false, "abc"},
        {"填充", "ab", 4, false, "ab  "},
        {"恰好", "abc", 3, false, "abc"},
        {"右截断", "abcdef", 3, true, "def"},
        {"右填充", "ab", 4, true, "  ab"},
        {"中文截断", "你好世界", 2, false, "你好"},
        {"中文填充", "你", 3, false, "你  "},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := PaddingOrTruncate(tt.input, tt.toLen, tt.right)
            if got != tt.expect {
                t.Errorf("PaddingOrTruncate(%q, %d, %v) = %q, want %q", tt.input, tt.toLen, tt.right, got, tt.expect)
            }
        })
    }
}

// TestTruncateAndAppendSuffix 验证截断后追加后缀
func TestTruncateAndAppendSuffix(t *testing.T) {
    tests := []struct {
        name   string
        input  string
        toLen  int
        suffix string
        right  bool
        expect string
    }{
        {"截断追加", "abcdef", 3, "...", false, "abc..."},
        {"未截断不追加", "abc", 3, "...", false, "abc"},
        {"短字符串不追加", "ab", 5, "...", false, "ab"},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := TruncateAndAppendSuffix(tt.input, tt.toLen, tt.suffix, tt.right)
            if got != tt.expect {
                t.Errorf("TruncateAndAppendSuffix(%q, %d, %q, %v) = %q, want %q",
                    tt.input, tt.toLen, tt.suffix, tt.right, got, tt.expect)
            }
        })
    }
}

// TestSub 验证 Sub 字符串截取
func TestSub(t *testing.T) {
    tests := []struct {
        name   string
        src    string
        flag   string
        before bool
        last   bool
        expect string
    }{
        {"前部首个", "a/b/c", "/", true, false, "a"},
        {"后部首个", "a/b/c", "/", false, false, "b/c"},
        {"前部最后", "a/b/c", "/", true, true, "a/b"},
        {"后部最后", "a/b/c", "/", false, true, "c"},
        {"无匹配", "abc", "/", true, false, "abc"},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := Sub(tt.src, tt.flag, tt.before, tt.last)
            if got != tt.expect {
                t.Errorf("Sub(%q, %q, %v, %v) = %q, want %q",
                    tt.src, tt.flag, tt.before, tt.last, got, tt.expect)
            }
        })
    }
}

// TestSubBefore 验证 SubBefore
func TestSubBefore(t *testing.T) {
    tests := []struct {
        src    string
        flag   string
        expect string
    }{
        {"a/b/c", "/", "a"},
        {"abc", "/", "abc"},
        {"hello world", " ", "hello"},
    }
    for _, tt := range tests {
        got := SubBefore(tt.src, tt.flag)
        if got != tt.expect {
            t.Errorf("SubBefore(%q, %q) = %q, want %q", tt.src, tt.flag, got, tt.expect)
        }
    }
}

// TestSubAfter 验证 SubAfter
func TestSubAfter(t *testing.T) {
    tests := []struct {
        src    string
        flag   string
        expect string
    }{
        {"a/b/c", "/", "c"},
        {"abc", "/", "abc"},
        {"hello world", " ", "world"},
    }
    for _, tt := range tests {
        got := SubAfter(tt.src, tt.flag)
        if got != tt.expect {
            t.Errorf("SubAfter(%q, %q) = %q, want %q", tt.src, tt.flag, got, tt.expect)
        }
    }
}

// TestSubByte 验证 SubByte 按 byte 截取
func TestSubByte(t *testing.T) {
    tests := []struct {
        name   string
        src    string
        flag   byte
        before bool
        last   bool
        expect string
    }{
        {"前部首个", "a/b/c", '/', true, false, "a"},
        {"后部首个", "a/b/c", '/', false, false, "b/c"},
        {"前部最后", "a/b/c", '/', true, true, "a/b"},
        {"后部最后", "a/b/c", '/', false, true, "c"},
        {"无匹配", "abc", '/', true, false, "abc"},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := SubByte(tt.src, tt.flag, tt.before, tt.last)
            if got != tt.expect {
                t.Errorf("SubByte(%q, %q, %v, %v) = %q, want %q",
                    tt.src, string(tt.flag), tt.before, tt.last, got, tt.expect)
            }
        })
    }
}

// TestSubByteBefore 验证 SubByteBefore
func TestSubByteBefore(t *testing.T) {
    if got := SubByteBefore("a/b/c", '/'); got != "a" {
        t.Errorf("SubByteBefore = %q, want %q", got, "a")
    }
    if got := SubByteBefore("abc", '/'); got != "abc" {
        t.Errorf("SubByteBefore no match = %q, want %q", got, "abc")
    }
}

// TestSubByteAfter 验证 SubByteAfter
func TestSubByteAfter(t *testing.T) {
    if got := SubByteAfter("a/b/c", '/'); got != "c" {
        t.Errorf("SubByteAfter = %q, want %q", got, "c")
    }
    if got := SubByteAfter("abc", '/'); got != "abc" {
        t.Errorf("SubByteAfter no match = %q, want %q", got, "abc")
    }
}

// TestHash 验证 Hash 函数稳定性和不同输入不同输出
func TestHash(t *testing.T) {
    h1 := Hash("hello")
    h2 := Hash("hello")
    h3 := Hash("world")
    if h1 != h2 {
        t.Errorf("相同输入应产生相同哈希: %d != %d", h1, h2)
    }
    if h1 == h3 {
        t.Errorf("不同输入应产生不同哈希: %d == %d", h1, h3)
    }
}

// TestStringBuilder 验证 StringBuilder 链式调用
func TestStringBuilder(t *testing.T) {
    var sb StringBuilder
    got := sb.Append("a").AppendByte('b').AppendBytes([]byte("c")).String()
    if got != "abc" {
        t.Errorf("StringBuilder chain = %q, want %q", got, "abc")
    }
}

// TestPaddingOrTruncate_ChineseBoundary 中文 rune 边界测试
func TestPaddingOrTruncate_ChineseBoundary(t *testing.T) {
    // 中文 rune 长度按 rune 计数, 一个中文字符 = 1 rune
    input := "你好世界"
    got := PaddingOrTruncate(input, 2)
    if got != "你好" {
        t.Errorf("PaddingOrTruncate(%q, 2) = %q, want %q", input, got, "你好")
    }

    // 验证截断后的字符串是合法 UTF-8
    for _, r := range got {
        if r == 0xFFFD {
            t.Errorf("截断结果包含无效 rune: %q", got)
        }
    }
}

// TestTruncate_MultibyteRune 验证截断不会在多字节 rune 中间截断
func TestTruncate_MultibyteRune(t *testing.T) {
    input := "abc你好"
    got := Truncate(input, 4)
    if !strings.HasPrefix(got, "abc") {
        t.Errorf("Truncate 多字节字符串损坏: %q", got)
    }
    // 确保截断结果为 4 个 rune
    count := 0
    for range got {
        count++
    }
    if count != 4 {
        t.Errorf("Truncate 结果 rune 数 = %d, want 4", count)
    }
}
