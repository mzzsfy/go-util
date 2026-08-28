package script

import (
	"math"
	"strconv"
	"testing"
)

// ========== 数值精度与边界测试 ==========
// 完整覆盖整数/浮点数的字面量边界、运算溢出、类型转换、位运算与数学模式
// int 对应 Go int(平台宽度), float 对应 Go float64
// 混合 int/float 运算: int 操作数自动转换为 float64 参与运算, 见 Test_Numeric_MixedIntFloat

// ---------- 整数字面量边界 ----------

// Test_Numeric_IntZero 整数零
func Test_Numeric_IntZero(t *testing.T) {
	t.Run("字面量0", func(t *testing.T) { runIntTest(t, `0`, 0) })
}

// Test_Numeric_IntUnits 整数基本单位值
func Test_Numeric_IntUnits(t *testing.T) {
	t.Run("正一", func(t *testing.T) { runIntTest(t, `1`, 1) })
	t.Run("负一", func(t *testing.T) { runIntTest(t, `-1`, -1) })
}

// Test_Numeric_IntMaxBoundary 平台int最大值边界
func Test_Numeric_IntMaxBoundary(t *testing.T) {
	maxInt := strconv.FormatInt(math.MaxInt, 10)
	t.Run("最大值", func(t *testing.T) { runIntTest(t, maxInt, math.MaxInt) })
	t.Run("最大值减一", func(t *testing.T) { runIntTest(t, maxInt+" - 1", math.MaxInt-1) })
}

// Test_Numeric_IntMinBoundary 平台int最小值边界
// 字面量最小值无法直接解析, 编译器先解析正数再取负, 正数溢出
func Test_Numeric_IntMinBoundary(t *testing.T) {
	const minInt64 = -9223372036854775808
	t.Run("最小值字面量编译错误", func(t *testing.T) {
		runErrorTest(t, `-9223372036854775808`)
	})
	maxInt := strconv.FormatInt(math.MaxInt, 10)
	t.Run("通过减法得到最小值", func(t *testing.T) {
		runIntTest(t, "-"+maxInt+" - 1", math.MinInt)
	})
	t.Run("最小值加一", func(t *testing.T) {
		runIntTest(t, "-"+maxInt+" - 1 + 1", math.MinInt+1)
	})
}

// Test_Numeric_IntLarge 大整数
func Test_Numeric_IntLarge(t *testing.T) {
	t.Run("一百万", func(t *testing.T) { runIntTest(t, `1000000`, 1000000) })
	t.Run("负大数", func(t *testing.T) { runIntTest(t, `-999999`, -999999) })
}

// Test_Numeric_HexLiterals 十六进制字面量
func Test_Numeric_HexLiterals(t *testing.T) {
	t.Run("0xFF", func(t *testing.T) { runIntTest(t, `0xFF`, 255) })
	t.Run("0x0", func(t *testing.T) { runIntTest(t, `0x0`, 0) })
	// 超过32位int的字面量与引擎同样按int截断, 期望值用运行时截断保持同步
	deadbeef := int64(0xDEADBEEF)
	t.Run("0xDEADBEEF", func(t *testing.T) { runIntTest(t, `0xDEADBEEF`, int(deadbeef)) })
	t.Run("0X大写前缀", func(t *testing.T) { runIntTest(t, `0XFF`, 255) })
}

// Test_Numeric_HexCaseInsensitive 十六进制大小写等价
func Test_Numeric_HexCaseInsensitive(t *testing.T) {
	t.Run("小写ff等于大写FF", func(t *testing.T) { runIntTest(t, `0xff`, 0xFF) })
	t.Run("混合大小写", func(t *testing.T) { runIntTest(t, `0xAbCdEf`, 0xABCDEF) })
}

// Test_Numeric_NumberSeparators 数字分隔符
func Test_Numeric_NumberSeparators(t *testing.T) {
	t.Run("百万分隔", func(t *testing.T) { runIntTest(t, `1_000_000`, 1000000) })
	t.Run("十六进制分隔", func(t *testing.T) { runIntTest(t, `0xFF_FF`, 0xFFFF) })
	t.Run("浮点分隔", func(t *testing.T) { runFloatTest(t, `1_000.5_5`, 1000.55) })
	t.Run("连续分隔", func(t *testing.T) { runIntTest(t, `1__000`, 1000) })
}

// ---------- 浮点数字面量精度 ----------

// Test_Numeric_FloatZero 浮点零值
func Test_Numeric_FloatZero(t *testing.T) {
	t.Run("正零", func(t *testing.T) { runFloatTest(t, `0.0`, 0.0) })
}

// Test_Numeric_FloatNegativeZero 浮点负零
func Test_Numeric_FloatNegativeZero(t *testing.T) {
	result := runScript(t, `-0.0`)
	if result.Float() != 0.0 {
		t.Errorf("期望 -0.0, 得到 %v", result.Float())
	}
	// 验证是负零: math.Signbit 检查符号位
	if !math.Signbit(result.Data.(float64)) {
		t.Errorf("期望负零(符号位为负), 得到正零")
	}
}

// Test_Numeric_FloatPrecision 浮点精度问题
func Test_Numeric_FloatPrecision(t *testing.T) {
	// 0.1 + 0.2 在 float64 下不精确等于 0.3
	t.Run("0.1+0.2精度", func(t *testing.T) {
		// 使用变量强制 float64 运算, 避免编译期高精度常量
		a, b := 0.1, 0.2
		runFloatTest(t, `0.1 + 0.2`, a+b)
	})
	t.Run("精度导致不等", func(t *testing.T) {
		runBoolTest(t, `0.1 + 0.2 == 0.3`, false)
	})
}

// Test_Numeric_FloatPi 圆周率精度
func Test_Numeric_FloatPi(t *testing.T) {
	runFloatTest(t, `3.14159265358979`, 3.14159265358979)
}

// Test_Numeric_FloatLongDecimal 多位小数精度
func Test_Numeric_FloatLongDecimal(t *testing.T) {
	runFloatTest(t, `1.123456789012345`, 1.123456789012345)
}

// Test_Numeric_FloatLargeInteger 大整数部分的浮点数
func Test_Numeric_FloatLargeInteger(t *testing.T) {
	runFloatTest(t, `1000000.5`, 1000000.5)
}

// Test_Numeric_FloatTiny 极小浮点值
func Test_Numeric_FloatTiny(t *testing.T) {
	t.Run("极小值0.00001", func(t *testing.T) { runFloatTest(t, `0.00001`, 0.00001) })
}

// Test_Numeric_NoScientificNotation 科学计数法不支持(编译错误)
func Test_Numeric_NoScientificNotation(t *testing.T) {
	t.Run("1e10编译错误", func(t *testing.T) {
		runErrorTest(t, `1e10`)
	})
}

// ---------- 整数运算 ----------

// Test_Numeric_IntAddBoundary 加法边界
func Test_Numeric_IntAddBoundary(t *testing.T) {
	maxInt := strconv.FormatInt(math.MaxInt, 10)
	t.Run("最大值加一溢出", func(t *testing.T) {
		// int 溢出回绕, 引擎不报错
		runIntTest(t, maxInt+" + 1", math.MinInt)
	})
	t.Run("大数加法", func(t *testing.T) {
		// 期望值与引擎同为int回绕语义
		a, b := 1000000000, 2000000000
		runIntTest(t, `1000000000 + 2000000000`, a+b)
	})
}

// Test_Numeric_IntSubBoundary 减法边界
func Test_Numeric_IntSubBoundary(t *testing.T) {
	maxInt := strconv.FormatInt(math.MaxInt, 10)
	t.Run("最小值减一溢出", func(t *testing.T) {
		// 最小值再减一溢出回绕到最大值
		runIntTest(t, "-"+maxInt+" - 1 - 1", math.MaxInt)
	})
	t.Run("大数减法", func(t *testing.T) {
		// 期望值与引擎同为int截断语义
		a := int64(3000000000)
		runIntTest(t, `3000000000 - 1000000000`, int(a)-1000000000)
	})
}

// Test_Numeric_IntMulLarge 乘法大数
func Test_Numeric_IntMulLarge(t *testing.T) {
	t.Run("百万乘百万", func(t *testing.T) {
		// 期望值与引擎同为int回绕语义
		a := 1000000
		runIntTest(t, `1000000 * 1000000`, a*a)
	})
}

// Test_Numeric_IntOverflow 整数溢出回绕
func Test_Numeric_IntOverflow(t *testing.T) {
	// 引擎对溢出不报错, 使用 Go int64 回绕语义
	t.Run("大数乘法溢出", func(t *testing.T) {
		result := runScript(t, `9223372036854775807 * 2`)
		// 不会 panic, 结果是溢出后的值
		if result.Type != TypeInt {
			t.Errorf("期望 int 类型, 得到 %d", result.Type)
		}
	})
}

// Test_Numeric_IntDivisionTruncation 整数除法截断
func Test_Numeric_IntDivisionTruncation(t *testing.T) {
	t.Run("7除以2", func(t *testing.T) { runIntTest(t, `7 / 2`, 3) })
	t.Run("1除以2", func(t *testing.T) { runIntTest(t, `1 / 2`, 0) })
}

// Test_Numeric_NegativeDivision 负数除法
func Test_Numeric_NegativeDivision(t *testing.T) {
	// Go 整数除法向零截断
	t.Run("-7除以2", func(t *testing.T) { runIntTest(t, `-7 / 2`, -3) })
	t.Run("-8除以2", func(t *testing.T) { runIntTest(t, `-8 / 2`, -4) })
	t.Run("7除以-2", func(t *testing.T) { runIntTest(t, `7 / -2`, -3) })
}

// Test_Numeric_NegativeModulo 负数取模
func Test_Numeric_NegativeModulo(t *testing.T) {
	// Go 取模结果符号跟随被除数
	t.Run("-7取模3", func(t *testing.T) { runIntTest(t, `-7 % 3`, -1) })
	t.Run("7取模-3", func(t *testing.T) { runIntTest(t, `7 % -3`, 1) })
	t.Run("-6取模3", func(t *testing.T) { runIntTest(t, `-6 % 3`, 0) })
}

// Test_Numeric_ZeroArithmetic 零的各种运算
func Test_Numeric_ZeroArithmetic(t *testing.T) {
	t.Run("0加0", func(t *testing.T) { runIntTest(t, `0 + 0`, 0) })
	t.Run("0减0", func(t *testing.T) { runIntTest(t, `0 - 0`, 0) })
	t.Run("0乘5", func(t *testing.T) { runIntTest(t, `0 * 5`, 0) })
	t.Run("0除以5", func(t *testing.T) { runIntTest(t, `0 / 5`, 0) })
	t.Run("5乘0", func(t *testing.T) { runIntTest(t, `5 * 0`, 0) })
}

// ---------- 浮点运算 ----------

// Test_Numeric_FloatAddSub 浮点加减法
func Test_Numeric_FloatAddSub(t *testing.T) {
	t.Run("浮点加", func(t *testing.T) { runFloatTest(t, `1.5 + 2.5`, 4.0) })
	t.Run("浮点减", func(t *testing.T) { runFloatTest(t, `3.14 - 0.14`, 3.0) })
}

// Test_Numeric_FloatMulDiv 浮点乘除法
func Test_Numeric_FloatMulDiv(t *testing.T) {
	t.Run("浮点乘", func(t *testing.T) { runFloatTest(t, `1.5 * 2.0`, 3.0) })
	t.Run("浮点除", func(t *testing.T) { runFloatTest(t, `6.0 / 3.0`, 2.0) })
}

// Test_Numeric_MixedIntFloat 混合 int/float 运算验证正确行为
// int 操作数自动转换为 float64 参与运算
func Test_Numeric_MixedIntFloat(t *testing.T) {
	// 1 + 1.5: int 1 转为 1.0, 1.0 + 1.5 = 2.5
	t.Run("int加float", func(t *testing.T) { runFloatTest(t, `1 + 1.5`, 2.5) })
	// 2 - 0.5: int 2 转为 2.0, 2.0 - 0.5 = 1.5
	t.Run("int减float", func(t *testing.T) { runFloatTest(t, `2 - 0.5`, 1.5) })
	// 2 * 1.5: int 2 转为 2.0, 2.0 * 1.5 = 3.0
	t.Run("int乘float", func(t *testing.T) { runFloatTest(t, `2 * 1.5`, 3.0) })
	// 5 / 2.0: int 5 转为 5.0, 5.0 / 2.0 = 2.5
	t.Run("int除float", func(t *testing.T) { runFloatTest(t, `5 / 2.0`, 2.5) })
	// 5.0 / 2: int 2 转为 2.0, 5.0 / 2.0 = 2.5
	t.Run("float除int", func(t *testing.T) { runFloatTest(t, `5.0 / 2`, 2.5) })
}

// ---------- 除零行为 ----------

// Test_Numeric_DivisionByZeroInt 整数除零
func Test_Numeric_DivisionByZeroInt(t *testing.T) {
	t.Run("5除以0", func(t *testing.T) { runRuntimeErrorTest(t, `5 / 0`) })
	t.Run("0除以0", func(t *testing.T) { runRuntimeErrorTest(t, `0 / 0`) })
	t.Run("负数除以0", func(t *testing.T) { runRuntimeErrorTest(t, `-5 / 0`) })
}

// Test_Numeric_DivisionByZeroFloat 浮点除零
func Test_Numeric_DivisionByZeroFloat(t *testing.T) {
	// 引擎对浮点除零也报运行时错误(不返回 Inf/NaN)
	t.Run("5.0除以0.0", func(t *testing.T) { runRuntimeErrorTest(t, `5.0 / 0.0`) })
	t.Run("0.0除以0.0", func(t *testing.T) { runRuntimeErrorTest(t, `0.0 / 0.0`) })
}

// Test_Numeric_ModuloByZero 取模零
func Test_Numeric_ModuloByZero(t *testing.T) {
	t.Run("5取模0", func(t *testing.T) { runRuntimeErrorTest(t, `5 % 0`) })
}

// ---------- 一元运算 ----------

// Test_Numeric_UnaryNegate 一元取负
func Test_Numeric_UnaryNegate(t *testing.T) {
	t.Run("正数取负", func(t *testing.T) { runIntTest(t, `-5`, -5) })
	t.Run("负数取负", func(t *testing.T) { runIntTest(t, `-(-5)`, 5) })
}

// Test_Numeric_UnaryNegateZero 零取负
func Test_Numeric_UnaryNegateZero(t *testing.T) {
	t.Run("整数零取负", func(t *testing.T) { runIntTest(t, `-0`, 0) })
}

// Test_Numeric_UnaryBitNot 位取反(^)
func Test_Numeric_UnaryBitNot(t *testing.T) {
	t.Run("5取反", func(t *testing.T) { runIntTest(t, `^5`, -6) })
	t.Run("0取反", func(t *testing.T) { runIntTest(t, `^0`, -1) })
	t.Run("-1取反", func(t *testing.T) { runIntTest(t, `^(-1)`, 0) })
}

// Test_Numeric_UnaryLogicalNot 逻辑非
func Test_Numeric_UnaryLogicalNot(t *testing.T) {
	t.Run("非true", func(t *testing.T) { runBoolTest(t, `!true`, false) })
	t.Run("非false", func(t *testing.T) { runBoolTest(t, `!false`, true) })
	t.Run("双重非", func(t *testing.T) { runBoolTest(t, `!!true`, true) })
}

// ---------- 位运算 ----------

// Test_Numeric_BitwiseAnd 位与运算
func Test_Numeric_BitwiseAnd(t *testing.T) {
	t.Run("5与3", func(t *testing.T) { runIntTest(t, `5 & 3`, 1) })
	t.Run("0与255", func(t *testing.T) { runIntTest(t, `0 & 255`, 0) })
	t.Run("255与255", func(t *testing.T) { runIntTest(t, `255 & 255`, 255) })
}

// Test_Numeric_BitwiseOr 位或运算
func Test_Numeric_BitwiseOr(t *testing.T) {
	t.Run("5或3", func(t *testing.T) { runIntTest(t, `5 | 3`, 7) })
	t.Run("0或255", func(t *testing.T) { runIntTest(t, `0 | 255`, 255) })
}

// Test_Numeric_BitwiseXor 位异或运算
func Test_Numeric_BitwiseXor(t *testing.T) {
	t.Run("5异或3", func(t *testing.T) { runIntTest(t, `5 ^ 3`, 6) })
	t.Run("0异或255", func(t *testing.T) { runIntTest(t, `0 ^ 255`, 255) })
	t.Run("自异或", func(t *testing.T) { runIntTest(t, `255 ^ 255`, 0) })
}

// Test_Numeric_BitwiseShift 位移运算
func Test_Numeric_BitwiseShift(t *testing.T) {
	t.Run("左移4位", func(t *testing.T) { runIntTest(t, `1 << 4`, 16) })
	t.Run("右移4位", func(t *testing.T) { runIntTest(t, `256 >> 4`, 16) })
	t.Run("左移0位", func(t *testing.T) { runIntTest(t, `1 << 0`, 1) })
	t.Run("大数左移", func(t *testing.T) { runIntTest(t, `1 << 20`, 1048576) })
}

// Test_Numeric_BitwiseNegative 负数位运算
func Test_Numeric_BitwiseNegative(t *testing.T) {
	// -5 在二进制补码下: ...11111011, & 3 = 3
	t.Run("负数与", func(t *testing.T) { runIntTest(t, `-5 & 3`, 3) })
	t.Run("负数或", func(t *testing.T) { runIntTest(t, `-1 | 0`, -1) })
}

// Test_Numeric_BitwiseOnFloat 浮点位运算报错
func Test_Numeric_BitwiseOnFloat(t *testing.T) {
	t.Run("浮点位与", func(t *testing.T) { runRuntimeErrorTest(t, `1.5 & 1`) })
	t.Run("浮点位或", func(t *testing.T) { runRuntimeErrorTest(t, `1.0 | 2.0`) })
}

// ---------- 类型转换 ----------

// Test_Numeric_IntConversion int()类型转换
func Test_Numeric_IntConversion(t *testing.T) {
	t.Run("浮点截断正", func(t *testing.T) { runIntTest(t, `int(3.14)`, 3) })
	t.Run("浮点截断接近整数", func(t *testing.T) { runIntTest(t, `int(3.99)`, 3) })
	t.Run("浮点截断负", func(t *testing.T) { runIntTest(t, `int(-3.14)`, -3) })
	t.Run("字符串转int", func(t *testing.T) { runIntTest(t, `int("123")`, 123) })
	t.Run("true转1", func(t *testing.T) { runIntTest(t, `int(true)`, 1) })
	t.Run("false转0", func(t *testing.T) { runIntTest(t, `int(false)`, 0) })
}

// Test_Numeric_IntConversionError int()转换错误
func Test_Numeric_IntConversionError(t *testing.T) {
	t.Run("非数字字符串", func(t *testing.T) { runRuntimeErrorTest(t, `int("abc")`) })
	t.Run("浮点字符串", func(t *testing.T) { runRuntimeErrorTest(t, `int("3.14")`) })
	t.Run("空字符串", func(t *testing.T) { runRuntimeErrorTest(t, `int("")`) })
	t.Run("nil转换", func(t *testing.T) { runRuntimeErrorTest(t, `int(nil)`) })
}

// Test_Numeric_FloatConversion float()类型转换
func Test_Numeric_FloatConversion(t *testing.T) {
	t.Run("int转float", func(t *testing.T) { runFloatTest(t, `float(123)`, 123.0) })
	t.Run("字符串转float", func(t *testing.T) { runFloatTest(t, `float("3.14")`, 3.14) })
	t.Run("true转1.0", func(t *testing.T) { runFloatTest(t, `float(true)`, 1.0) })
	t.Run("负int转float", func(t *testing.T) { runFloatTest(t, `float(-7)`, -7.0) })
}

// Test_Numeric_StringConversion string()类型转换
func Test_Numeric_StringConversion(t *testing.T) {
	t.Run("int转字符串", func(t *testing.T) { runStringTest(t, `string(123)`, "123") })
	t.Run("float转字符串", func(t *testing.T) { runStringTest(t, `string(3.14)`, "3.14") })
	t.Run("bool转字符串", func(t *testing.T) { runStringTest(t, `string(true)`, "true") })
}

// Test_Numeric_NoBoolConversion 无bool()内置函数
func Test_Numeric_NoBoolConversion(t *testing.T) {
	// 引擎没有 bool() 转换函数, 编译时报未定义变量
	t.Run("bool(1)编译错误", func(t *testing.T) { runErrorTest(t, `bool(1)`) })
}

// ---------- 混合类型比较 ----------

// Test_Numeric_CrossTypeEqual 跨类型相等比较
// 数值类型(int/float)自动提升比较, 1 == 1.0 为 true
func Test_Numeric_CrossTypeEqual(t *testing.T) {
	t.Run("int与float相等", func(t *testing.T) { runBoolTest(t, `1 == 1.0`, true) })
	t.Run("int与float等号", func(t *testing.T) { runBoolTest(t, `1 != 1.0`, false) })
}

// Test_Numeric_MixedComparison 混合 int/float 大小比较验证正确行为
// int 与 float 比较时自动转换, 正确返回比较结果
func Test_Numeric_MixedComparison(t *testing.T) {
	t.Run("int小于float", func(t *testing.T) { runBoolTest(t, `1 < 1.5`, true) })
	t.Run("float大于int", func(t *testing.T) { runBoolTest(t, `1.5 > 1`, true) })
	t.Run("int小于等于float", func(t *testing.T) { runBoolTest(t, `1 <= 1.5`, true) })
}

// ---------- 数值比较 ----------

// Test_Numeric_IntComparison 整数比较
func Test_Numeric_IntComparison(t *testing.T) {
	t.Run("相等", func(t *testing.T) { runBoolTest(t, `5 == 5`, true) })
	t.Run("不等", func(t *testing.T) { runBoolTest(t, `5 != 6`, true) })
	t.Run("小于", func(t *testing.T) { runBoolTest(t, `3 < 5`, true) })
	t.Run("大于", func(t *testing.T) { runBoolTest(t, `5 > 3`, true) })
	t.Run("小于等于", func(t *testing.T) { runBoolTest(t, `5 <= 5`, true) })
	t.Run("大于等于", func(t *testing.T) { runBoolTest(t, `5 >= 5`, true) })
}

// Test_Numeric_FloatComparison 浮点数比较
func Test_Numeric_FloatComparison(t *testing.T) {
	t.Run("相等", func(t *testing.T) { runBoolTest(t, `3.14 == 3.14`, true) })
	t.Run("小于", func(t *testing.T) { runBoolTest(t, `1.5 < 2.5`, true) })
	t.Run("大于", func(t *testing.T) { runBoolTest(t, `2.5 > 1.5`, true) })
}

// Test_Numeric_ChainedComparison 链式比较
func Test_Numeric_ChainedComparison(t *testing.T) {
	t.Run("逻辑与链", func(t *testing.T) { runBoolTest(t, `1 < 2 && 2 < 3`, true) })
	t.Run("逻辑或链", func(t *testing.T) { runBoolTest(t, `5 > 3 || 1 > 2`, true) })
}

// ---------- 特殊值在算术中 ----------

// Test_Numeric_SpecialValueArithmetic 特殊值参与算术运算
func Test_Numeric_SpecialValueArithmetic(t *testing.T) {
	t.Run("nil加int报错", func(t *testing.T) { runRuntimeErrorTest(t, `nil + 1`) })
	t.Run("bool加int报错", func(t *testing.T) { runRuntimeErrorTest(t, `true + 1`) })
	t.Run("nil减int报错", func(t *testing.T) { runRuntimeErrorTest(t, `nil - 1`) })
}

// Test_Numeric_StringConcatWithNumber 字符串与数字拼接
func Test_Numeric_StringConcatWithNumber(t *testing.T) {
	// string + 基本类型自动拼接
	t.Run("字符串加int", func(t *testing.T) { runStringTest(t, `"a" + 1`, "a1") })
	t.Run("int加字符串", func(t *testing.T) { runStringTest(t, `1 + "a"`, "1a") })
	t.Run("字符串加float", func(t *testing.T) { runStringTest(t, `"v" + 3.14`, "v3.14") })
}

// ---------- 运算优先级 ----------

// Test_Numeric_OperatorPrecedence 运算优先级
func Test_Numeric_OperatorPrecedence(t *testing.T) {
	t.Run("乘法优先于加法", func(t *testing.T) { runIntTest(t, `1 + 2 * 3`, 7) })
	t.Run("括号改变优先级", func(t *testing.T) { runIntTest(t, `(1 + 2) * 3`, 9) })
	t.Run("取模优先级", func(t *testing.T) { runIntTest(t, `2 + 7 % 3`, 3) })
	t.Run("链式加法", func(t *testing.T) { runIntTest(t, `1 + 2 + 3 + 4`, 10) })
	t.Run("链式乘法", func(t *testing.T) { runIntTest(t, `2 * 3 * 4`, 24) })
	t.Run("括号内取负", func(t *testing.T) { runIntTest(t, `-(3 + 4)`, -7) })
}

// ---------- 数学模式 ----------

// Test_Numeric_MathFactorial 阶乘
func Test_Numeric_MathFactorial(t *testing.T) {
	// 5! = 120
	runIntTest(t, `1 * 2 * 3 * 4 * 5`, 120)
}

// Test_Numeric_MathFactorialFunction 阶乘函数实现
func Test_Numeric_MathFactorialFunction(t *testing.T) {
	runIntTest(t, `
fn fact(n) {
    if n <= 1 { return 1 }
    return n * fact(n - 1)
}
fact(6)
`, 720)
}

// Test_Numeric_MathFibonacci 斐波那契
func Test_Numeric_MathFibonacci(t *testing.T) {
	runIntTest(t, `
fn fib(n) {
    if n < 2 { return n }
    return fib(n - 1) + fib(n - 2)
}
fib(10)
`, 55)
}

// Test_Numeric_MathGCD 最大公约数
func Test_Numeric_MathGCD(t *testing.T) {
	runIntTest(t, `
fn gcd(a, b) {
    for b != 0 {
        t := b
        b = a % b
        a = t
    }
    return a
}
gcd(48, 18)
`, 6)
}

// Test_Numeric_MathLCM 最小公倍数
func Test_Numeric_MathLCM(t *testing.T) {
	runIntTest(t, `
fn gcd(a, b) {
    for b != 0 {
        t := b
        b = a % b
        a = t
    }
    return a
}
fn lcm(a, b) {
    return a * b / gcd(a, b)
}
lcm(4, 6)
`, 12)
}

// Test_Numeric_MathIsPrime 素数判定
func Test_Numeric_MathIsPrime(t *testing.T) {
	runBoolTest(t, `
fn isPrime(n) {
    if n < 2 { return false }
    i := 2
    for i * i <= n {
        if n % i == 0 { return false }
        i = i + 1
    }
    return true
}
isPrime(17)
`, true)
}

// Test_Numeric_FloatModuloError 浮点取模报错
func Test_Numeric_FloatModuloError(t *testing.T) {
	// 取模运算只支持整数
	t.Run("float取模", func(t *testing.T) { runRuntimeErrorTest(t, `5.5 % 2.0`) })
}

// Test_Numeric_ModuloChain 取模链式运算
func Test_Numeric_ModuloChain(t *testing.T) {
	// 17 % 5 = 2, 2 % 3 = 2
	runIntTest(t, `17 % 5 % 3`, 2)
}
