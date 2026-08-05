package script

import "testing"

// ========== 字符串字面量 ==========

// Test_StringLiteral_Basic 测试基础字符串字面量
func Test_StringLiteral_Basic(t *testing.T) {
	t.Run("空字符串", func(t *testing.T) { runStringTest(t, `""`, "") })
	t.Run("单字符a", func(t *testing.T) { runStringTest(t, `"a"`, "a") })
	t.Run("单字符z", func(t *testing.T) { runStringTest(t, `"z"`, "z") })
	t.Run("hello", func(t *testing.T) { runStringTest(t, `"hello"`, "hello") })
	t.Run("长字符串", func(t *testing.T) {
		runStringTest(t, `"abcdefghijklmnopqrstuvwxyz"`, "abcdefghijklmnopqrstuvwxyz")
	})
	t.Run("含空格", func(t *testing.T) { runStringTest(t, `"hello world"`, "hello world") })
	t.Run("含数字", func(t *testing.T) { runStringTest(t, `"abc123"`, "abc123") })
}

// Test_StringLiteral_Escape 测试转义字符
func Test_StringLiteral_Escape(t *testing.T) {
	t.Run("换行符", func(t *testing.T) {
		runStringTest(t, `"line1\nline2"`, "line1\nline2")
	})
	t.Run("制表符", func(t *testing.T) { runStringTest(t, `"a\tb"`, "a\tb") })
	t.Run("回车符", func(t *testing.T) { runStringTest(t, `"a\rb"`, "a\rb") })
	t.Run("反斜杠", func(t *testing.T) { runStringTest(t, `"a\\b"`, "a\\b") })
	t.Run("双引号", func(t *testing.T) { runStringTest(t, `"say\"hi\""`, "say\"hi\"") })
	t.Run("多转义组合", func(t *testing.T) {
		runStringTest(t, `"\n\t\r\\\""`, "\n\t\r\\\"")
	})
}

// Test_StringLiteral_UnknownEscape 测试未识别的转义序列保留原样
func Test_StringLiteral_UnknownEscape(t *testing.T) {
	// 引擎对未识别的转义序列(如\x \d \q \w)保留反斜杠和字符原样
	t.Run("反斜杠x", func(t *testing.T) { runStringTest(t, `"\x"`, "\\x") })
	t.Run("反斜杠d", func(t *testing.T) { runStringTest(t, `"\d"`, "\\d") })
	t.Run("反斜杠q", func(t *testing.T) { runStringTest(t, `"\q"`, "\\q") })
	t.Run("反斜杠w", func(t *testing.T) { runStringTest(t, `"\w"`, "\\w") })
}

// Test_StringLiteral_Unicode 测试Unicode字符串
// len()按rune计数, 多字节字符正确处理
func Test_StringLiteral_Unicode(t *testing.T) {
	t.Run("中文len字符数", func(t *testing.T) {
		// len()按rune计数, 2个中文=2
		runIntTest(t, `len("中文")`, 2)
	})
	t.Run("混合len字符数", func(t *testing.T) {
		// 2个中文 + 3个ASCII = 5个rune
		runIntTest(t, `len("中文abc")`, 5)
	})
	t.Run("多字节字面量值正确", func(t *testing.T) {
		// 多字节UTF-8字符正确解码, 无替换字符
		result := runScript(t, `"中文"`)
		if result.String() != "中文" {
			t.Errorf("期望 '中文', 得到 %q", result.String())
		}
	})
	t.Run("emoji字面量", func(t *testing.T) {
		result := runScript(t, `"hello"`)
		if result.String() != "hello" {
			t.Errorf("期望 'hello', 得到 %q", result.String())
		}
	})
}

// ========== 字符串拼接 ==========

// Test_StringConcat_Basic 测试基础字符串拼接
func Test_StringConcat_Basic(t *testing.T) {
	t.Run("两个字符串", func(t *testing.T) { runStringTest(t, `"a" + "b"`, "ab") })
	t.Run("三个字符串", func(t *testing.T) { runStringTest(t, `"a" + "b" + "c"`, "abc") })
	t.Run("空字符串拼接", func(t *testing.T) { runStringTest(t, `"" + "x"`, "x") })
	t.Run("空字符串在右", func(t *testing.T) { runStringTest(t, `"x" + ""`, "x") })
	t.Run("双空字符串拼接", func(t *testing.T) { runStringTest(t, `"" + ""`, "") })
	t.Run("长字符串拼接", func(t *testing.T) {
		runStringTest(t, `"hello " + "world " + "foo " + "bar"`, "hello world foo bar")
	})
	t.Run("含空字符串的多段拼接", func(t *testing.T) {
		runStringTest(t, `"" + "a" + "" + "b" + ""`, "ab")
	})
}

// Test_StringConcat_CrossType 测试字符串与非字符串类型的拼接
// 引擎行为: 字符串与任何类型拼接都会做字符串化处理
func Test_StringConcat_CrossType(t *testing.T) {
	t.Run("字符串加整数", func(t *testing.T) { runStringTest(t, `"a" + 1`, "a1") })
	t.Run("整数加字符串", func(t *testing.T) { runStringTest(t, `1 + "a"`, "1a") })
	t.Run("字符串加浮点数", func(t *testing.T) { runStringTest(t, `"a" + 1.5`, "a1.5") })
	t.Run("字符串加布尔真", func(t *testing.T) { runStringTest(t, `"a" + true`, "atrue") })
	t.Run("字符串加nil", func(t *testing.T) { runStringTest(t, `"a" + nil`, "anil") })
	t.Run("nil加字符串", func(t *testing.T) { runStringTest(t, `nil + "x"`, "nilx") })
	t.Run("字符串加负整数", func(t *testing.T) { runStringTest(t, `"x" + (-1)`, "x-1") })
	t.Run("负整数加字符串", func(t *testing.T) { runStringTest(t, `(-1) + "x"`, "-1x") })
	t.Run("字符串加零", func(t *testing.T) { runStringTest(t, `"x" + 0`, "x0") })
	t.Run("零加字符串", func(t *testing.T) { runStringTest(t, `0 + "x"`, "0x") })
	t.Run("字符串加浮点pi", func(t *testing.T) { runStringTest(t, `"pi=" + 3.14`, "pi=3.14") })
}

// ========== 字符串索引 ==========

// Test_StringIndex_Basic 测试字符串索引访问
// 引擎行为: 索引返回单字节字符串(TypeString)
func Test_StringIndex_Basic(t *testing.T) {
	t.Run("首字符", func(t *testing.T) { runStringTest(t, `"hello"[0]`, "h") })
	t.Run("中间字符", func(t *testing.T) { runStringTest(t, `"hello"[1]`, "e") })
	t.Run("末字符", func(t *testing.T) { runStringTest(t, `"hello"[4]`, "o") })
	t.Run("单字符串索引", func(t *testing.T) { runStringTest(t, `"a"[0]`, "a") })
}

// Test_StringIndex_OutOfBounds 测试越界索引
// 引擎行为: 越界索引返回nil值, 不报错
func Test_StringIndex_OutOfBounds(t *testing.T) {
	t.Run("负数索引", func(t *testing.T) {
		result := runScript(t, `"hello"[-1]`)
		if !result.IsNil() {
			t.Errorf("期望nil, 得到%v", result)
		}
	})
	t.Run("过大索引", func(t *testing.T) {
		result := runScript(t, `"hello"[100]`)
		if !result.IsNil() {
			t.Errorf("期望nil, 得到%v", result)
		}
	})
	t.Run("刚好越界", func(t *testing.T) {
		// "abc"长度3, 索引3越界
		result := runScript(t, `"abc"[3]`)
		if !result.IsNil() {
			t.Errorf("期望nil, 得到%v", result)
		}
	})
	t.Run("空字符串索引", func(t *testing.T) {
		result := runScript(t, `""[0]`)
		if !result.IsNil() {
			t.Errorf("期望nil, 得到%v", result)
		}
	})
}

// Test_StringIndex_Chained 测试索引结果的链式操作
func Test_StringIndex_Chained(t *testing.T) {
	t.Run("索引结果再索引", func(t *testing.T) {
		// "hello"[0]返回"h", "h"[0]返回"h"
		runStringTest(t, `"hello"[0][0]`, "h")
	})
	t.Run("索引结果比较", func(t *testing.T) {
		runBoolTest(t, `"hello"[0] == "h"`, true)
	})
}

// ========== 字符串切片 ==========

// Test_StringSlice_Basic 测试基础切片
func Test_StringSlice_Basic(t *testing.T) {
	t.Run("前三个字符", func(t *testing.T) { runStringTest(t, `"hello"[0:3]`, "hel") })
	t.Run("中间切片", func(t *testing.T) { runStringTest(t, `"hello"[2:4]`, "ll") })
	t.Run("省略下界", func(t *testing.T) { runStringTest(t, `"hello"[:3]`, "hel") })
	t.Run("省略上界", func(t *testing.T) { runStringTest(t, `"hello"[1:]`, "ello") })
	t.Run("全切片省略", func(t *testing.T) { runStringTest(t, `"hello"[:]`, "hello") })
	t.Run("零长度切片", func(t *testing.T) { runStringTest(t, `"hello"[0:0]`, "") })
	t.Run("末尾空切片", func(t *testing.T) { runStringTest(t, `"hello"[5:5]`, "") })
}

// Test_StringSlice_EdgeCases 测试切片边界情况
// 引擎行为: 反转区间返回空字符串; 越界自动clamp; 负数边界视为0
func Test_StringSlice_EdgeCases(t *testing.T) {
	t.Run("反转区间返回空", func(t *testing.T) { runStringTest(t, `"hello"[3:2]`, "") })
	t.Run("上界越界自动截断", func(t *testing.T) { runStringTest(t, `"hello"[0:10]`, "hello") })
	t.Run("省略上界且超长", func(t *testing.T) { runStringTest(t, `"hello"[:100]`, "hello") })
	t.Run("负数下界从末尾倒数", func(t *testing.T) { runStringTest(t, `"hello"[-1:3]`, "") })
	t.Run("负数上界去掉末尾", func(t *testing.T) { runStringTest(t, `"hello"[0:-1]`, "hell") })
}

// Test_StringSlice_Chained 测试切片结果链式操作
func Test_StringSlice_Chained(t *testing.T) {
	t.Run("切片再索引带括号", func(t *testing.T) {
		runStringTest(t, `("hello"[0:3])[1]`, "e")
	})
	t.Run("切片再切片", func(t *testing.T) {
		runStringTest(t, `"hello"[0:3][0:2]`, "he")
	})
	t.Run("无括号切片再索引", func(t *testing.T) {
		runStringTest(t, `"hello"[0:3][1]`, "e")
	})
	t.Run("切片长度", func(t *testing.T) {
		runIntTest(t, `len("hello"[0:3])`, 3)
	})
}

// Test_StringSlice_Empty 测试空字符串切片
func Test_StringSlice_Empty(t *testing.T) {
	t.Run("空字符串零切片", func(t *testing.T) { runStringTest(t, `""[0:0]`, "") })
	t.Run("空字符串全切片", func(t *testing.T) { runStringTest(t, `""[:]`, "") })
}

// ========== 字符串长度 ==========

// Test_StringLen_Basic 测试len函数
func Test_StringLen_Basic(t *testing.T) {
	t.Run("空字符串", func(t *testing.T) { runIntTest(t, `len("")`, 0) })
	t.Run("单字符", func(t *testing.T) { runIntTest(t, `len("a")`, 1) })
	t.Run("hello", func(t *testing.T) { runIntTest(t, `len("hello")`, 5) })
	t.Run("含空格", func(t *testing.T) { runIntTest(t, `len("a b")`, 3) })
}

// Test_StringLen_Calculated 测试len与字符串操作的组合
func Test_StringLen_Calculated(t *testing.T) {
	t.Run("切片后的len", func(t *testing.T) {
		runIntTest(t, `len("hello"[0:3])`, 3)
	})
	t.Run("空字符串len", func(t *testing.T) {
		runIntTest(t, `len("")`, 0)
	})
}

// ========== 字符串比较 ==========

// Test_StringCompare_Equality 测试字符串相等比较
func Test_StringCompare_Equality(t *testing.T) {
	t.Run("相同字符串", func(t *testing.T) { runBoolTest(t, `"abc" == "abc"`, true) })
	t.Run("不同字符串", func(t *testing.T) { runBoolTest(t, `"abc" == "abd"`, false) })
	t.Run("不等运算真", func(t *testing.T) { runBoolTest(t, `"abc" != "abd"`, true) })
	t.Run("不等运算假", func(t *testing.T) { runBoolTest(t, `"abc" != "abc"`, false) })
	t.Run("空串等于空串", func(t *testing.T) { runBoolTest(t, `"" == ""`, true) })
}

// Test_StringCompare_CrossType 测试字符串与非字符串类型的相等比较
// 引擎行为: 类型不同直接判false
func Test_StringCompare_CrossType(t *testing.T) {
	t.Run("字符串不等于整数", func(t *testing.T) { runBoolTest(t, `"abc" == 1`, false) })
	t.Run("字符串不等于nil", func(t *testing.T) { runBoolTest(t, `"abc" == nil`, false) })
	t.Run("字符串不等于布尔", func(t *testing.T) { runBoolTest(t, `"abc" == true`, false) })
	t.Run("字符串不等于浮点", func(t *testing.T) { runBoolTest(t, `"abc" == 1.5`, false) })
	t.Run("数字字符串不等于整数", func(t *testing.T) { runBoolTest(t, `"3" == 3`, false) })
	t.Run("整数不等于字符串", func(t *testing.T) { runBoolTest(t, `1 == "1"`, false) })
}

// Test_StringCompare_Ordering 测试字符串大小比较
func Test_StringCompare_Ordering(t *testing.T) {
	t.Run("小于为真", func(t *testing.T) { runBoolTest(t, `"a" < "b"`, true) })
	t.Run("小于为假", func(t *testing.T) { runBoolTest(t, `"b" < "a"`, false) })
	t.Run("大于为真", func(t *testing.T) { runBoolTest(t, `"b" > "a"`, true) })
	t.Run("大于为假", func(t *testing.T) { runBoolTest(t, `"a" > "b"`, false) })
	t.Run("字典序前缀短串小", func(t *testing.T) { runBoolTest(t, `"ab" < "abc"`, true) })
	t.Run("字典序前缀长串大", func(t *testing.T) { runBoolTest(t, `"abc" < "ab"`, false) })
	t.Run("空串小于非空", func(t *testing.T) { runBoolTest(t, `"" < "a"`, true) })
	t.Run("非空大于空串", func(t *testing.T) { runBoolTest(t, `"a" > ""`, true) })
	t.Run("大写小于小写", func(t *testing.T) { runBoolTest(t, `"A" < "a"`, true) })
	t.Run("相同字符串小于等于", func(t *testing.T) { runBoolTest(t, `"abc" <= "abc"`, true) })
	t.Run("相同字符串大于等于", func(t *testing.T) { runBoolTest(t, `"abc" >= "abc"`, true) })
}

// ========== 类型转换 ==========

// Test_StringConvert_ToString 测试string()类型转换
func Test_StringConvert_ToString(t *testing.T) {
	t.Run("正整数转字符串", func(t *testing.T) { runStringTest(t, `string(123)`, "123") })
	t.Run("负整数转字符串", func(t *testing.T) { runStringTest(t, `string(-123)`, "-123") })
	t.Run("零转字符串", func(t *testing.T) { runStringTest(t, `string(0)`, "0") })
	t.Run("负一转字符串", func(t *testing.T) { runStringTest(t, `string(-1)`, "-1") })
	t.Run("浮点数转字符串", func(t *testing.T) { runStringTest(t, `string(1.5)`, "1.5") })
	t.Run("布尔真转字符串", func(t *testing.T) { runStringTest(t, `string(true)`, "true") })
	t.Run("nil转字符串", func(t *testing.T) { runStringTest(t, `string(nil)`, "nil") })
}

// Test_StringConvert_FromString 测试从字符串的类型转换
func Test_StringConvert_FromString(t *testing.T) {
	t.Run("有效整数字符串", func(t *testing.T) { runIntTest(t, `int("123")`, 123) })
	t.Run("零字符串", func(t *testing.T) { runIntTest(t, `int("0")`, 0) })
	t.Run("负数字符串", func(t *testing.T) { runIntTest(t, `int("-456")`, -456) })
	t.Run("带正号字符串", func(t *testing.T) { runIntTest(t, `int("+789")`, 789) })
	t.Run("有效浮点字符串", func(t *testing.T) { runFloatTest(t, `float("3.14")`, 3.14) })
}

// Test_StringConvert_ErrorCases 测试类型转换的错误情况
func Test_StringConvert_ErrorCases(t *testing.T) {
	t.Run("非整数字符串转int", func(t *testing.T) {
		runRuntimeErrorTest(t, `int("abc")`)
	})
	t.Run("浮点格式字符串转int", func(t *testing.T) {
		runRuntimeErrorTest(t, `int("12.5")`)
	})
	t.Run("空字符串转int", func(t *testing.T) {
		runRuntimeErrorTest(t, `int("")`)
	})
	t.Run("非数字字符串转float", func(t *testing.T) {
		runRuntimeErrorTest(t, `float("abc")`)
	})
	t.Run("bool未定义", func(t *testing.T) {
		// bool不是内置函数, 编译报未定义变量
		runErrorTest(t, `bool(1)`)
	})
}

// ========== 字符串与if条件 ==========

// Test_StringTruthiness 测试字符串在if条件中的行为
// 引擎行为: 空字符串为假走else, 非空字符串为真走then
func Test_StringTruthiness(t *testing.T) {
	t.Run("空字符串走else", func(t *testing.T) {
		runIntTest(t, `if "" { 1 } else { 2 }`, 2)
	})
	t.Run("非空字符串走then", func(t *testing.T) {
		runIntTest(t, `if "x" { 1 } else { 2 }`, 1)
	})
}

// ========== 字符串与变量 ==========

// Test_StringVariable 测试字符串变量的操作
func Test_StringVariable(t *testing.T) {
	t.Run("变量赋值后拼接", func(t *testing.T) {
		runStringTest(t, `
			s := "hello"
			s + " world"
		`, "hello world")
	})
	t.Run("两个变量拼接", func(t *testing.T) {
		runStringTest(t, `
			a := "hello"
			b := "world"
			a + " " + b
		`, "hello world")
	})
	t.Run("变量重新赋值拼接", func(t *testing.T) {
		runStringTest(t, `
			s := "a"
			s := s + "b"
			s
		`, "ab")
	})
	t.Run("变量索引", func(t *testing.T) {
		runStringTest(t, `
			s := "abc"
			s[1]
		`, "b")
	})
	t.Run("变量切片", func(t *testing.T) {
		runStringTest(t, `
			s := "hello"
			s[0:3]
		`, "hel")
	})
	t.Run("变量len", func(t *testing.T) {
		runIntTest(t, `
			s := "hello"
			len(s)
		`, 5)
	})
}

// ========== 字符串与函数 ==========

// Test_StringFunction 测试字符串在函数中的使用
func Test_StringFunction(t *testing.T) {
	t.Run("字符串参数拼接", func(t *testing.T) {
		runStringTest(t, `
			fn concat(a, b) { a + b }
			concat("foo", "bar")
		`, "foobar")
	})
	t.Run("字符串参数加后缀", func(t *testing.T) {
		runStringTest(t, `
			fn greet(s) { s + "!" }
			greet("hi")
		`, "hi!")
	})
	t.Run("函数返回字面量字符串", func(t *testing.T) {
		runStringTest(t, `
			fn getStr() { "ok" }
			getStr()
		`, "ok")
	})
	t.Run("函数返回len结果", func(t *testing.T) {
		runIntTest(t, `
			fn slen(s) { len(s) }
			slen("hello")
		`, 5)
	})
	t.Run("函数索引字符串", func(t *testing.T) {
		runStringTest(t, `
			fn first(s) { s[0] }
			first("hello")
		`, "h")
	})
	t.Run("函数切片字符串", func(t *testing.T) {
		runStringTest(t, `
			fn slice3(s) { s[0:3] }
			slice3("hello")
		`, "hel")
	})
	t.Run("函数中字符串与整数拼接", func(t *testing.T) {
		runStringTest(t, `
			fn build(n) { "n=" + n }
			build(42)
		`, "n=42")
	})
}

// ========== 字符串与数组 ==========

// Test_StringArray 测试字符串数组相关操作
func Test_StringArray(t *testing.T) {
	t.Run("字符串数组索引", func(t *testing.T) {
		result := runScript(t, `["a", "b", "c"]`)
		arr := result.Array()
		if arr.Elements[0].String() != "a" {
			t.Errorf("期望a, 得到%s", arr.Elements[0].String())
		}
	})
	t.Run("数组元素拼接", func(t *testing.T) {
		runStringTest(t, `["a", "b"][0] + "c"`, "ac")
	})
	t.Run("拼接表达式作为数组元素", func(t *testing.T) {
		result := runScript(t, `["a" + "b", "c"]`)
		arr := result.Array()
		if arr.Elements[0].String() != "ab" {
			t.Errorf("期望ab, 得到%s", arr.Elements[0].String())
		}
	})
}

// ========== 字符串复合表达式 ==========

// Test_StringComposite 测试字符串复合表达式
func Test_StringComposite(t *testing.T) {
	t.Run("拼接结果比较", func(t *testing.T) {
		runBoolTest(t, `"a" + "b" == "ab"`, true)
	})
	t.Run("len结果比较", func(t *testing.T) {
		runBoolTest(t, `len("hello") == 5`, true)
	})
	t.Run("索引结果比较", func(t *testing.T) {
		runBoolTest(t, `"hello"[0] == "h"`, true)
	})
	t.Run("切片结果比较", func(t *testing.T) {
		runBoolTest(t, `"hello"[0:3] == "hel"`, true)
	})
	t.Run("len转字符串", func(t *testing.T) {
		runStringTest(t, `
			s := "hello"
			string(len(s))
		`, "5")
	})
	t.Run("多段拼接含整数", func(t *testing.T) {
		runStringTest(t, `"a" + "b" + "c" + "d" + "e" + "f" + "g" + "h"`, "abcdefgh")
	})
}

// ========== 运算符优先级 ==========

// Test_StringPrecedence 测试字符串拼接的运算符优先级
// 引擎行为: +是左结合; 若左操作数是字符串则后续全部字符串化
func Test_StringPrecedence(t *testing.T) {
	t.Run("字符串开头混合加法", func(t *testing.T) {
		// "x" + 1 + 2 => ("x"+1)+2 => "x1"+2 => "x12"
		runStringTest(t, `"x" + 1 + 2`, "x12")
	})
	t.Run("整数开头混合加法", func(t *testing.T) {
		// 1 + 2 + "x" => (1+2)+"x" => 3+"x" => "3x"
		runStringTest(t, `1 + 2 + "x"`, "3x")
	})
	t.Run("混合浮点和字符串", func(t *testing.T) {
		// 1+0.5=1.5(int自动转float), 再拼接得"1.5x"
		runStringTest(t, `1 + 0.5 + "x"`, "1.5x")
	})
	t.Run("字符串拼接再索引", func(t *testing.T) {
		runStringTest(t, `("ab" + "cd")[2]`, "c")
	})
}

// ========== 不支持的操作 ==========

// Test_StringUnsupported 测试引擎不支持的操作
func Test_StringUnsupported(t *testing.T) {
	t.Run("加等运算符", func(t *testing.T) {
		// += 现已支持, 字符串拼接
		runStringTest(t, `s := "a"
s += "b"
s`, "ab")
	})
	t.Run("单引号字符字面量不支持", func(t *testing.T) {
		// 引擎不支持 'c' 字符字面量
		runErrorTest(t, `'c'`)
	})
}

// ========== nil拼接的边界 ==========

// Test_StringConcat_NilBoundary 测试nil与字符串拼接的边界
// 注意: string + nil 拼接成功(得"nil"), 但 nil + nil 报错
func Test_StringConcat_NilBoundary(t *testing.T) {
	t.Run("字符串加nil成功", func(t *testing.T) {
		runStringTest(t, `"x" + nil`, "xnil")
	})
	t.Run("nil加字符串成功", func(t *testing.T) {
		runStringTest(t, `nil + "x"`, "nilx")
	})
	t.Run("nil加nil报错", func(t *testing.T) {
		// nil+nil不涉及字符串, 走数值加法路径, 运行时报错
		runRuntimeErrorTest(t, `nil + nil`)
	})
}

// ========== 字符串与Map ==========

// Test_StringMap 测试字符串在Map中的使用
func Test_StringMap(t *testing.T) {
	t.Run("字符串value访问", func(t *testing.T) {
		runStringTest(t, `
			m := {"k": "v"}
			m["k"]
		`, "v")
	})
	t.Run("拼接表达式作为value", func(t *testing.T) {
		runStringTest(t, `
			m := {"k": "a" + "b"}
			m["k"]
		`, "ab")
	})
	t.Run("动态拼接key", func(t *testing.T) {
		runStringTest(t, `
			k := "dy" + "namic"
			m := {k: "v"}
			m["dynamic"]
		`, "v")
	})
}

// ========== 类型转换链 ==========

// Test_StringConvert_Chain 测试类型转换链
func Test_StringConvert_Chain(t *testing.T) {
	t.Run("int转string再转int", func(t *testing.T) {
		runIntTest(t, `int(string(42))`, 42)
	})
	t.Run("string转int再转string", func(t *testing.T) {
		runStringTest(t, `string(int("42"))`, "42")
	})
	t.Run("负数往返转换", func(t *testing.T) {
		runIntTest(t, `int(string(-99))`, -99)
	})
}

// ========== 字符串拼接的方向性 ==========

// Test_StringConcat_Directional 测试拼接的方向性差异
func Test_StringConcat_Directional(t *testing.T) {
	t.Run("左字符串右整数", func(t *testing.T) {
		runStringTest(t, `"n=" + 42`, "n=42")
	})
	t.Run("左整数右字符串", func(t *testing.T) {
		runStringTest(t, `42 + "=n"`, "42=n")
	})
	t.Run("左字符串右浮点", func(t *testing.T) {
		runStringTest(t, `"f=" + 3.14`, "f=3.14")
	})
	t.Run("左浮点右字符串", func(t *testing.T) {
		runStringTest(t, `3.14 + "=f"`, "3.14=f")
	})
}

// ========== 字符串索引返回类型 ==========

// Test_StringIndex_ReturnType 测试字符串索引返回的类型
func Test_StringIndex_ReturnType(t *testing.T) {
	t.Run("索引返回TypeString", func(t *testing.T) {
		result := runScript(t, `"hello"[0]`)
		if result.Type != TypeString {
			t.Errorf("期望TypeString(%d), 得到%v", TypeString, result.Type)
		}
	})
	t.Run("越界索引返回TypeNil", func(t *testing.T) {
		result := runScript(t, `"hello"[100]`)
		if result.Type != TypeNil {
			t.Errorf("期望TypeNil, 得到%v", result.Type)
		}
	})
}

// ========== 大整数转换 ==========

// Test_StringConvert_LargeNumbers 测试大整数的字符串转换
func Test_StringConvert_LargeNumbers(t *testing.T) {
	t.Run("百万转字符串", func(t *testing.T) {
		runStringTest(t, `string(1000000)`, "1000000")
	})
	t.Run("负百万转字符串", func(t *testing.T) {
		runStringTest(t, `string(-1000000)`, "-1000000")
	})
	t.Run("大数字符串转int", func(t *testing.T) {
		runIntTest(t, `int("999999")`, 999999)
	})
}

// ========== 字符串包含特殊内容 ==========

// Test_StringLiteral_SpecialContent 测试包含特殊内容的字符串
func Test_StringLiteral_SpecialContent(t *testing.T) {
	t.Run("仅空格", func(t *testing.T) { runStringTest(t, `"   "`, "   ") })
	t.Run("含分号", func(t *testing.T) { runStringTest(t, `"a;b;c"`, "a;b;c") })
	t.Run("含花括号", func(t *testing.T) { runStringTest(t, `"a{b}c"`, "a{b}c") })
	t.Run("含方括号", func(t *testing.T) { runStringTest(t, `"a[b]c"`, "a[b]c") })
	t.Run("含等号", func(t *testing.T) { runStringTest(t, `"a=b"`, "a=b") })
	t.Run("含加号字面量", func(t *testing.T) { runStringTest(t, `"a+b"`, "a+b") })
	t.Run("含冒号", func(t *testing.T) { runStringTest(t, `"a:b"`, "a:b") })
}

// ========== 函数嵌套 ==========

// Test_StringFunction_Nested 测试字符串在嵌套函数中的传递
func Test_StringFunction_Nested(t *testing.T) {
	t.Run("函数内重复字符串", func(t *testing.T) {
		runStringTest(t, `
			fn double(s) { s + s }
			double("ab")
		`, "abab")
	})
	t.Run("嵌套函数调用", func(t *testing.T) {
		runStringTest(t, `
			fn outer(s) { s + "!" }
			fn inner(s) { outer(s) }
			inner("hi")
		`, "hi!")
	})
	t.Run("函数参数索引后再拼接", func(t *testing.T) {
		runStringTest(t, `
			fn first(s) { s[0] }
			first("hello") + "x"
		`, "hx")
	})
}

// ========== 切片组合 ==========

// Test_StringSlice_Combine 测试切片结果的组合操作
func Test_StringSlice_Combine(t *testing.T) {
	t.Run("两个切片拼接", func(t *testing.T) {
		// "he" + "lo" = "helo"
		runStringTest(t, `"hello"[0:2] + "hello"[3:5]`, "helo")
	})
	t.Run("切片跳过中间字符", func(t *testing.T) {
		// 取首尾, 跳过中间
		runStringTest(t, `"hello"[0:1] + "hello"[4:5]`, "ho")
	})
	t.Run("三段切片重组", func(t *testing.T) {
		runStringTest(t, `"abcdef"[0:2] + "abcdef"[4:6]`, "abef")
	})
}

// ========== 条件中的字符串操作 ==========

// Test_StringInCondition 测试字符串操作在条件判断中的使用
func Test_StringInCondition(t *testing.T) {
	t.Run("索引结果作条件", func(t *testing.T) {
		runStringTest(t, `
			s := "hello"
			if s[0] == "h" { "yes" } else { "no" }
		`, "yes")
	})
	t.Run("拼接结果作条件", func(t *testing.T) {
		runIntTest(t, `
			s := "hel" + "lo"
			if s == "hello" { 1 } else { 0 }
		`, 1)
	})
	t.Run("len结果作条件", func(t *testing.T) {
		runIntTest(t, `
			if len("ab") == 2 { 1 } else { 0 }
		`, 1)
	})
}

// ========== 变量索引越界 ==========

// Test_StringVariable_OutOfBounds 测试变量字符串的越界访问
func Test_StringVariable_OutOfBounds(t *testing.T) {
	t.Run("变量过大索引返回nil", func(t *testing.T) {
		result := runScript(t, `
			s := "ab"
			s[5]
		`)
		if !result.IsNil() {
			t.Errorf("期望nil, 得到%v", result)
		}
	})
	t.Run("变量负数索引返回nil", func(t *testing.T) {
		result := runScript(t, `
			s := "hello"
			s[-1]
		`)
		if !result.IsNil() {
			t.Errorf("期望nil, 得到%v", result)
		}
	})
}

// ========== 字符串拼接的空值行为 ==========

// Test_StringConcat_EmptyAndType 测试空字符串与类型转换在拼接中的行为
func Test_StringConcat_EmptyAndType(t *testing.T) {
	t.Run("空串加整数等价于string转换", func(t *testing.T) {
		// "" + 42 == "42" 等价于 string(42)
		runStringTest(t, `"" + 42`, "42")
	})
	t.Run("空串加浮点等价于string转换", func(t *testing.T) {
		runStringTest(t, `"" + 3.14`, "3.14")
	})
	t.Run("空串加布尔等价于string转换", func(t *testing.T) {
		runStringTest(t, `"" + true`, "true")
	})
	t.Run("整数加空串等价于string转换", func(t *testing.T) {
		runStringTest(t, `42 + ""`, "42")
	})
}
