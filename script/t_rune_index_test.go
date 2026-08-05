package script

import "testing"

// Test_RuneIndex_UTF8 UTF-8字符串按rune索引
func Test_RuneIndex_UTF8(t *testing.T) {
	// len按rune, 索引按rune, 切片按rune - 三者一致
	runIntTest(t, `len("中文")`, 2)
	runStringTest(t, `"中文"[0]`, "中")
	runStringTest(t, `"中文"[1]`, "文")
	runStringTest(t, `"中文"[0:1]`, "中")
	runStringTest(t, `"中文"[0:2]`, "中文")
	runStringTest(t, `"中文"[1:]`, "文")
}

func Test_RuneIndex_Mixed(t *testing.T) {
	// 混合ASCII和多字节字符
	runIntTest(t, `len("a中b")`, 3)
	runStringTest(t, `"a中b"[0]`, "a")
	runStringTest(t, `"a中b"[1]`, "中")
	runStringTest(t, `"a中b"[2]`, "b")
	runStringTest(t, `"a中b"[1:3]`, "中b")
}

func Test_RuneIndex_Emoji(t *testing.T) {
	// 4字节emoji字符
	runIntTest(t, `len("hi")`, 2)
	runStringTest(t, `"hello"[0]`, "h")
	runStringTest(t, `"hello"[4]`, "o")
}

func Test_RuneIndex_OutOfBounds(t *testing.T) {
	result, err := Eval(`"中文"[2]`)
	if err != nil {
		t.Fatal(err)
	}
	if !result.IsNil() {
		t.Fatalf("越界索引应返回nil, 得到%v", result)
	}
}
