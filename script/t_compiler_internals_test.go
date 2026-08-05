package script

import (
	"testing"
)

// ========== 编译器内部机制测试 ==========
// 深度测试编译器内部机制:常量池、局部变量、跳转回填、操作码生成等

// ========== 测试辅助函数 ==========

func ciHasOp(insts []Instruction, op OpCode) bool {
	for _, inst := range insts {
		if inst.Op == op {
			return true
		}
	}
	return false
}

func ciCountOp(insts []Instruction, op OpCode) int {
	count := 0
	for _, inst := range insts {
		if inst.Op == op {
			count++
		}
	}
	return count
}

func ciCountConstantByInt(constants []Value, val int) int {
	count := 0
	for _, c := range constants {
		if c.Type == TypeInt && c.Int() == val {
			count++
		}
	}
	return count
}

func ciCountConstantByString(constants []Value, val string) int {
	count := 0
	for _, c := range constants {
		if c.Type == TypeString && c.String() == val {
			count++
		}
	}
	return count
}

func ciConstantHasInt(constants []Value, val int) bool {
	return ciCountConstantByInt(constants, val) > 0
}

func ciConstantHasString(constants []Value, val string) bool {
	return ciCountConstantByString(constants, val) > 0
}

func opName(op OpCode) string {
	switch op {
	case OpAdd:
		return "OpAdd"
	case OpSub:
		return "OpSub"
	case OpMul:
		return "OpMul"
	case OpDiv:
		return "OpDiv"
	case OpMod:
		return "OpMod"
	case OpLess:
		return "OpLess"
	case OpGreater:
		return "OpGreater"
	case OpEqual:
		return "OpEqual"
	case OpNotEqual:
		return "OpNotEqual"
	case OpLessEq:
		return "OpLessEq"
	case OpGreaterEq:
		return "OpGreaterEq"
	case OpBitAnd:
		return "OpBitAnd"
	case OpBitOr:
		return "OpBitOr"
	case OpBitXor:
		return "OpBitXor"
	case OpLShift:
		return "OpLShift"
	case OpRShift:
		return "OpRShift"
	case OpNot:
		return "OpNot"
	case OpNeg:
		return "OpNeg"
	case OpBitNot:
		return "OpBitNot"
	}
	return "unknown"
}

// ========== 常量池管理测试 ==========

// Test_CompilerInternals_ConstantPool_Int 整数常量被收集到常量池
func Test_CompilerInternals_ConstantPool_Int(t *testing.T) {
	t.Run("单个整数", func(t *testing.T) {
		script := compileScript(t, "42")
		if !ciConstantHasInt(script.Constants, 42) {
			t.Errorf("常量池应包含整数42")
		}
	})
	t.Run("多个不同整数", func(t *testing.T) {
		script := compileScript(t, "1 + 2 + 3")
		if !ciConstantHasInt(script.Constants, 1) {
			t.Error("常量池应包含整数1")
		}
		if !ciConstantHasInt(script.Constants, 2) {
			t.Error("常量池应包含整数2")
		}
		if !ciConstantHasInt(script.Constants, 3) {
			t.Error("常量池应包含整数3")
		}
	})
}

// Test_CompilerInternals_ConstantPool_String 字符串常量被收集到常量池
func Test_CompilerInternals_ConstantPool_String(t *testing.T) {
	t.Run("单个字符串", func(t *testing.T) {
		script := compileScript(t, "\"hello\"")
		if !ciConstantHasString(script.Constants, "hello") {
			t.Errorf("常量池应包含字符串hello")
		}
	})
	t.Run("多个不同字符串", func(t *testing.T) {
		script := compileScript(t, "\"a\" + \"b\"")
		if !ciConstantHasString(script.Constants, "a") {
			t.Error("常量池应包含字符串a")
		}
		if !ciConstantHasString(script.Constants, "b") {
			t.Error("常量池应包含字符串b")
		}
	})
}

// Test_CompilerInternals_ConstantPool_Float 浮点常量被收集到常量池
func Test_CompilerInternals_ConstantPool_Float(t *testing.T) {
	script := compileScript(t, "3.14")
	found := false
	for _, c := range script.Constants {
		if c.Type == TypeFloat && c.Float() == 3.14 {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("常量池应包含浮点数3.14")
	}
}

// Test_CompilerInternals_ConstantPool_Bool 布尔常量被收集到常量池
func Test_CompilerInternals_ConstantPool_Bool(t *testing.T) {
	t.Run("true", func(t *testing.T) {
		script := compileScript(t, "true")
		found := false
		for _, c := range script.Constants {
			if c.Type == TypeBool && c.Bool() {
				found = true
				break
			}
		}
		if !found {
			t.Error("常量池应包含布尔值true")
		}
	})
	t.Run("false", func(t *testing.T) {
		script := compileScript(t, "false")
		found := false
		for _, c := range script.Constants {
			if c.Type == TypeBool && !c.Bool() {
				found = true
				break
			}
		}
		if !found {
			t.Error("常量池应包含布尔值false")
		}
	})
}

// Test_CompilerInternals_ConstantPool_Nil nil常量被收集到常量池
func Test_CompilerInternals_ConstantPool_Nil(t *testing.T) {
	script := compileScript(t, "nil")
	hasNilConst := false
	for _, c := range script.Constants {
		if c.IsNil() {
			hasNilConst = true
			break
		}
	}
	if !hasNilConst {
		t.Error("常量池应包含nil值")
	}
}

// Test_CompilerInternals_ConstantPool_ArrayElements 数组常量中的元素常量
func Test_CompilerInternals_ConstantPool_ArrayElements(t *testing.T) {
	script := compileScript(t, "[10, 20, 30]")
	if !ciConstantHasInt(script.Constants, 10) {
		t.Error("常量池应包含数组元素10")
	}
	if !ciConstantHasInt(script.Constants, 20) {
		t.Error("常量池应包含数组元素20")
	}
	if !ciConstantHasInt(script.Constants, 30) {
		t.Error("常量池应包含数组元素30")
	}
}

// Test_CompilerInternals_ConstantPool_MapKeyValue Map常量中的键值常量
func Test_CompilerInternals_ConstantPool_MapKeyValue(t *testing.T) {
	script := compileScript(t, "{\"key1\": 100, \"key2\": 200}")
	if !ciConstantHasString(script.Constants, "key1") {
		t.Error("常量池应包含Map键key1")
	}
	if !ciConstantHasString(script.Constants, "key2") {
		t.Error("常量池应包含Map键key2")
	}
	if !ciConstantHasInt(script.Constants, 100) {
		t.Error("常量池应包含Map值100")
	}
	if !ciConstantHasInt(script.Constants, 200) {
		t.Error("常量池应包含Map值200")
	}
}

// Test_CompilerInternals_ConstantPool_Deduplication 相同常量值复用索引
func Test_CompilerInternals_ConstantPool_Deduplication(t *testing.T) {
	t.Run("整数去重", func(t *testing.T) {
		script := compileScript(t, "1 + 1 + 1")
		count := ciCountConstantByInt(script.Constants, 1)
		if count != 1 {
			t.Errorf("整数1应去重为1次, 实际%d次", count)
		}
	})
	t.Run("字符串去重", func(t *testing.T) {
		script := compileScript(t, "\"x\" + \"x\" + \"x\"")
		count := ciCountConstantByString(script.Constants, "x")
		if count != 1 {
			t.Errorf("字符串x应去重为1次, 实际%d次", count)
		}
	})
	t.Run("布尔去重", func(t *testing.T) {
		script := compileScript(t, "true && true")
		count := 0
		for _, c := range script.Constants {
			if c.Type == TypeBool && c.Bool() {
				count++
			}
		}
		if count != 1 {
			t.Errorf("布尔true应去重为1次, 实际%d次", count)
		}
	})
}

// Test_CompilerInternals_ConstantPool_Predefined 预定义常量验证
func Test_CompilerInternals_ConstantPool_Predefined(t *testing.T) {
	script := compileScript(t, "42")
	if !ciConstantHasInt(script.Constants, 0) {
		t.Error("常量池应包含预定义常量0(切片起始索引)")
	}
	if !ciConstantHasInt(script.Constants, SliceEndDefault) {
		t.Error("常量池应包含预定义切片结束索引常量")
	}
}

// Test_CompilerInternals_ConstantPool_Size 常量池大小与源码常量数量关系
func Test_CompilerInternals_ConstantPool_Size(t *testing.T) {
	script := compileScript(t, "1 + 2 + 3")
	if len(script.Constants) < 5 {
		t.Errorf("常量池大小应>=5(2预定义+3源码), 实际%d", len(script.Constants))
	}
}

// ========== 局部变量分配测试 ==========

// Test_CompilerInternals_LocalVar_SingleAlloc 单个变量声明分配索引
func Test_CompilerInternals_LocalVar_SingleAlloc(t *testing.T) {
	script := compileScript(t, "x := 10\nx")
	hasStoreLocal0 := false
	for _, inst := range script.Main.Instructions {
		if inst.Op == OpStoreLocal && inst.ArgCount >= 1 && inst.Args[0] == 0 {
			hasStoreLocal0 = true
			break
		}
	}
	if !hasStoreLocal0 {
		t.Error("单个变量应分配索引0")
	}
	if script.Main.NumLocals < 1 {
		t.Errorf("NumLocals应>=1, 实际%d", script.Main.NumLocals)
	}
}

// Test_CompilerInternals_LocalVar_MultipleAlloc 多个变量声明递增索引
func Test_CompilerInternals_LocalVar_MultipleAlloc(t *testing.T) {
	script := compileScript(t, "a := 1\nb := 2\nc := 3\na + b + c")
	storeIndices := make(map[int]bool)
	for _, inst := range script.Main.Instructions {
		if inst.Op == OpStoreLocal && inst.ArgCount >= 1 {
			storeIndices[inst.Args[0]] = true
		}
	}
	if !storeIndices[0] || !storeIndices[1] || !storeIndices[2] {
		t.Errorf("三个变量应分配索引0,1,2, StoreLocal索引集: %v", storeIndices)
	}
	if script.Main.NumLocals < 3 {
		t.Errorf("NumLocals应>=3, 实际%d", script.Main.NumLocals)
	}
}

// Test_CompilerInternals_LocalVar_Redefinition 变量重定义复用索引
func Test_CompilerInternals_LocalVar_Redefinition(t *testing.T) {
	script := compileScript(t, "x := 1\nx := 2\nx")
	storeCount := 0
	storeIdx := -1
	for _, inst := range script.Main.Instructions {
		if inst.Op == OpStoreLocal && inst.ArgCount >= 1 {
			storeCount++
			storeIdx = inst.Args[0]
		}
	}
	if storeCount != 2 {
		t.Errorf("应有2次StoreLocal, 实际%d次", storeCount)
	}
	if storeIdx != 0 {
		t.Errorf("重定义应复用索引0, 实际%d", storeIdx)
	}
	if script.Main.NumLocals != 1 {
		t.Errorf("重定义后NumLocals应为1, 实际%d", script.Main.NumLocals)
	}
}

// Test_CompilerInternals_LocalVar_ForRangeTemp for-range的临时变量分配
func Test_CompilerInternals_LocalVar_ForRangeTemp(t *testing.T) {
	script := compileScript(t, "for v := range [1, 2, 3] {\nprint(v)\n}\n0")
	if script.Main.NumLocals < 2 {
		t.Errorf("for-range应分配临时变量, NumLocals应>=2, 实际%d", script.Main.NumLocals)
	}
}

// Test_CompilerInternals_LocalVar_FunctionParams 函数参数作为局部变量
func Test_CompilerInternals_LocalVar_FunctionParams(t *testing.T) {
	script := compileScript(t, "fn f(a, b, c) {\nreturn a + b + c\n}\n0")
	if len(script.Functions) == 0 {
		t.Fatal("应有函数定义")
	}
	fn := script.Functions[0]
	if fn.NumParams != 3 {
		t.Errorf("参数数应为3, 实际%d", fn.NumParams)
	}
	if fn.NumLocals < fn.NumParams {
		t.Errorf("NumLocals(%d)应>=NumParams(%d)", fn.NumLocals, fn.NumParams)
	}
}

// Test_CompilerInternals_LocalVar_NumLocalsEmpty 空脚本的NumLocals
func Test_CompilerInternals_LocalVar_NumLocalsEmpty(t *testing.T) {
	script := compileScript(t, "42")
	if script.Main.NumLocals != 0 {
		t.Errorf("无变量声明时NumLocals应为0, 实际%d", script.Main.NumLocals)
	}
}

// ========== 表达式编译测试 ==========

// Test_CompilerInternals_Expr_OpConst 常量加载生成OpConst
func Test_CompilerInternals_Expr_OpConst(t *testing.T) {
	script := compileScript(t, "123")
	if !ciHasOp(script.Main.Instructions, OpConst) {
		t.Error("字面量123应生成OpConst指令")
	}
}

// Test_CompilerInternals_Expr_OpLoadLocal 变量加载生成OpLoadLocal
func Test_CompilerInternals_Expr_OpLoadLocal(t *testing.T) {
	script := compileScript(t, "x := 10\nx")
	if !ciHasOp(script.Main.Instructions, OpLoadLocal) {
		t.Error("变量引用x应生成OpLoadLocal指令")
	}
}

// Test_CompilerInternals_Expr_Arithmetic 算术运算操作码生成
func Test_CompilerInternals_Expr_Arithmetic(t *testing.T) {
	tests := []struct {
		input string
		op    OpCode
		name  string
	}{
		{"1 + 2", OpAdd, "加法"},
		{"1 - 2", OpSub, "减法"},
		{"1 * 2", OpMul, "乘法"},
		{"1 / 2", OpDiv, "除法"},
		{"1 % 2", OpMod, "取模"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			script := compileScript(t, tt.input)
			if !ciHasOp(script.Main.Instructions, tt.op) {
				t.Errorf("%s应生成%s", tt.input, opName(tt.op))
			}
		})
	}
}

// Test_CompilerInternals_Expr_Comparison 比较运算操作码生成
func Test_CompilerInternals_Expr_Comparison(t *testing.T) {
	tests := []struct {
		input string
		op    OpCode
		name  string
	}{
		{"1 < 2", OpLess, "小于"},
		{"1 > 2", OpGreater, "大于"},
		{"1 == 2", OpEqual, "等于"},
		{"1 != 2", OpNotEqual, "不等于"},
		{"1 <= 2", OpLessEq, "小于等于"},
		{"1 >= 2", OpGreaterEq, "大于等于"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			script := compileScript(t, tt.input)
			if !ciHasOp(script.Main.Instructions, tt.op) {
				t.Errorf("%s应生成%s", tt.input, opName(tt.op))
			}
		})
	}
}

// Test_CompilerInternals_Expr_Logic 逻辑运算操作码生成
func Test_CompilerInternals_Expr_Logic(t *testing.T) {
	t.Run("逻辑非", func(t *testing.T) {
		script := compileScript(t, "!true")
		if !ciHasOp(script.Main.Instructions, OpNot) {
			t.Error("!true应生成OpNot指令")
		}
	})
	t.Run("短路AND", func(t *testing.T) {
		script := compileScript(t, "true && false")
		if !ciHasOp(script.Main.Instructions, OpJumpIfFalse) {
			t.Error("&&应生成OpJumpIfFalse")
		}
		if !ciHasOp(script.Main.Instructions, OpDup) {
			t.Error("短路求值应使用OpDup")
		}
	})
	t.Run("短路OR", func(t *testing.T) {
		script := compileScript(t, "true || false")
		if !ciHasOp(script.Main.Instructions, OpJumpIfTrue) {
			t.Error("||应生成OpJumpIfTrue")
		}
	})
}

// Test_CompilerInternals_Expr_Bitwise 位运算操作码生成
func Test_CompilerInternals_Expr_Bitwise(t *testing.T) {
	tests := []struct {
		input string
		op    OpCode
		name  string
	}{
		{"1 & 2", OpBitAnd, "位与"},
		{"1 | 2", OpBitOr, "位或"},
		{"1 ^ 2", OpBitXor, "位异或"},
		{"1 << 2", OpLShift, "左移"},
		{"1 >> 2", OpRShift, "右移"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			script := compileScript(t, tt.input)
			if !ciHasOp(script.Main.Instructions, tt.op) {
				t.Errorf("%s应生成%s", tt.input, opName(tt.op))
			}
		})
	}
}

// Test_CompilerInternals_Expr_Unary 一元运算操作码生成
func Test_CompilerInternals_Expr_Unary(t *testing.T) {
	t.Run("取负", func(t *testing.T) {
		script := compileScript(t, "-5")
		if !ciHasOp(script.Main.Instructions, OpNeg) {
			t.Error("-5应生成OpNeg指令")
		}
	})
	t.Run("位取反", func(t *testing.T) {
		script := compileScript(t, "^5")
		if !ciHasOp(script.Main.Instructions, OpBitNot) {
			t.Error("^5应生成OpBitNot指令")
		}
	})
}

// Test_CompilerInternals_Expr_IndexAccess 索引访问生成OpIndex
func Test_CompilerInternals_Expr_IndexAccess(t *testing.T) {
	script := compileScript(t, "arr := [1, 2, 3]\narr[1]")
	if !ciHasOp(script.Main.Instructions, OpIndex) {
		t.Error("arr[1]应生成OpIndex指令")
	}
}

// Test_CompilerInternals_Expr_Slice 切片生成OpSlice
func Test_CompilerInternals_Expr_Slice(t *testing.T) {
	t.Run("完整切片", func(t *testing.T) {
		script := compileScript(t, "arr := [1, 2, 3, 4]\narr[1:3]")
		if !ciHasOp(script.Main.Instructions, OpSlice) {
			t.Error("arr[1:3]应生成OpSlice指令")
		}
	})
	t.Run("省略结束切片", func(t *testing.T) {
		script := compileScript(t, "arr := [1, 2, 3]\narr[1:]")
		if !ciHasOp(script.Main.Instructions, OpSlice) {
			t.Error("arr[1:]应生成OpSlice指令")
		}
		if !ciConstantHasInt(script.Constants, SliceEndDefault) {
			t.Error("省略切片结束应使用预定义切片结束索引常量")
		}
	})
	t.Run("省略起始切片", func(t *testing.T) {
		script := compileScript(t, "arr := [1, 2, 3]\narr[:2]")
		if !ciHasOp(script.Main.Instructions, OpSlice) {
			t.Error("arr[:2]应生成OpSlice指令")
		}
	})
}

// Test_CompilerInternals_Expr_Assignment 赋值生成OpStoreLocal
func Test_CompilerInternals_Expr_Assignment(t *testing.T) {
	script := compileScript(t, "x := 10\nx")
	storeCount := ciCountOp(script.Main.Instructions, OpStoreLocal)
	if storeCount < 1 {
		t.Error("变量声明应生成OpStoreLocal指令")
	}
}

// Test_CompilerInternals_Expr_IndexAssignment 索引赋值生成OpStoreIndex
func Test_CompilerInternals_Expr_IndexAssignment(t *testing.T) {
	script := compileScript(t, "arr := [1, 2, 3]\narr[0] = 100\n0")
	if !ciHasOp(script.Main.Instructions, OpStoreIndex) {
		t.Error("arr[0] = 100应生成OpStoreIndex指令")
	}
}

// ========== 控制流编译测试 ==========

// Test_CompilerInternals_ControlFlow_IfJump if语句的条件跳转
func Test_CompilerInternals_ControlFlow_IfJump(t *testing.T) {
	script := compileScript(t, "if true { 1 }")
	if !ciHasOp(script.Main.Instructions, OpJumpIfFalse) {
		t.Error("if语句应生成OpJumpIfFalse指令")
	}
}

// Test_CompilerInternals_ControlFlow_IfElseJump if-else的双分支跳转
func Test_CompilerInternals_ControlFlow_IfElseJump(t *testing.T) {
	script := compileScript(t, "if true { 1 } else { 2 }")
	if !ciHasOp(script.Main.Instructions, OpJumpIfFalse) {
		t.Error("if-else应生成OpJumpIfFalse指令")
	}
	jumpCount := ciCountOp(script.Main.Instructions, OpJump)
	if jumpCount < 1 {
		t.Error("if-else的then分支后应有OpJump跳过else分支")
	}
}

// Test_CompilerInternals_ControlFlow_ForBackJump for循环的回跳指令
func Test_CompilerInternals_ControlFlow_ForBackJump(t *testing.T) {
	script := compileScript(t, "x := 0\nfor x < 3 {\nx = x + 1\n}\n0")
	if !ciHasOp(script.Main.Instructions, OpJump) {
		t.Error("for循环应生成OpJump回跳指令")
	}
}

// Test_CompilerInternals_ControlFlow_ForRangeIter for-range的迭代指令
func Test_CompilerInternals_ControlFlow_ForRangeIter(t *testing.T) {
	script := compileScript(t, "for v := range [1, 2, 3] {\nprint(v)\n}\n0")
	if !ciHasOp(script.Main.Instructions, OpLen) {
		t.Error("for-range应生成OpLen获取集合长度")
	}
	if !ciHasOp(script.Main.Instructions, OpIndex) {
		t.Error("for-range应生成OpIndex获取元素")
	}
}

// Test_CompilerInternals_ControlFlow_BreakBackfill break的跳转回填
func Test_CompilerInternals_ControlFlow_BreakBackfill(t *testing.T) {
	script := compileScript(t, "for true {\nbreak\n}\n0")
	jumpCount := ciCountOp(script.Main.Instructions, OpJump)
	if jumpCount < 2 {
		t.Errorf("含break的循环应至少有2个OpJump, 实际%d", jumpCount)
	}
}

// Test_CompilerInternals_ControlFlow_ContinueBackfill continue的跳转回填
func Test_CompilerInternals_ControlFlow_ContinueBackfill(t *testing.T) {
	script := compileScript(t, "for true {\ncontinue\n}\n0")
	jumpCount := ciCountOp(script.Main.Instructions, OpJump)
	if jumpCount < 2 {
		t.Errorf("含continue的循环应至少有2个OpJump, 实际%d", jumpCount)
	}
}

// Test_CompilerInternals_ControlFlow_NestedLoopBreakContinue 嵌套循环break/continue
func Test_CompilerInternals_ControlFlow_NestedLoopBreakContinue(t *testing.T) {
	script := compileScript(t, "for i := 0; i < 3; i = i + 1 {\nfor j := 0; j < 3; j = j + 1 {\nif j == 1 {\nbreak\n}\n}\n}\n0")
	if !ciHasOp(script.Main.Instructions, OpJumpIfFalse) {
		t.Error("嵌套循环应包含条件跳转")
	}
	jumpCount := ciCountOp(script.Main.Instructions, OpJump)
	if jumpCount < 3 {
		t.Errorf("嵌套循环应有多条OpJump, 实际%d", jumpCount)
	}
}

// ========== 函数编译测试 ==========

// Test_CompilerInternals_Function_ReturnGen 函数定义生成OpReturn
func Test_CompilerInternals_Function_ReturnGen(t *testing.T) {
	script := compileScript(t, "fn f() {\nreturn 1\n}\n0")
	if len(script.Functions) == 0 {
		t.Fatal("应有函数定义")
	}
	fn := script.Functions[0]
	retCount := ciCountOp(fn.Instructions, OpReturn)
	if retCount < 1 {
		t.Error("函数体应包含OpReturn指令")
	}
}

// Test_CompilerInternals_Function_ImplicitReturn 无显式return的隐式OpReturn
func Test_CompilerInternals_Function_ImplicitReturn(t *testing.T) {
	script := compileScript(t, "fn f(x) {\nx + 1\n}\n0")
	fn := script.Functions[0]
	retCount := ciCountOp(fn.Instructions, OpReturn)
	if retCount < 1 {
		t.Error("即使无显式return,函数体末尾也应有隐式OpReturn")
	}
}

// Test_CompilerInternals_Function_ParamBinding 函数参数绑定到局部变量
func Test_CompilerInternals_Function_ParamBinding(t *testing.T) {
	script := compileScript(t, "fn f(a, b) {\nreturn a + b\n}\n0")
	fn := script.Functions[0]
	if !ciHasOp(fn.Instructions, OpLoadLocal) {
		t.Error("函数体应通过OpLoadLocal加载参数")
	}
}

// Test_CompilerInternals_Function_CallGen 函数调用生成OpCall
func Test_CompilerInternals_Function_CallGen(t *testing.T) {
	script := compileScript(t, "fn add(a, b) {\nreturn a + b\n}\nadd(1, 2)")
	if !ciHasOp(script.Main.Instructions, OpCall) {
		t.Error("函数调用add(1,2)应生成OpCall指令")
	}
}

// Test_CompilerInternals_Function_NumLocalsVsParams 函数NumLocals与NumParams
func Test_CompilerInternals_Function_NumLocalsVsParams(t *testing.T) {
	script := compileScript(t, "fn f(a, b) {\nc := a + b\nreturn c\n}\n0")
	fn := script.Functions[0]
	if fn.NumLocals < 3 {
		t.Errorf("NumLocals应>=3(2参数+1局部), 实际%d", fn.NumLocals)
	}
}

// Test_CompilerInternals_Function_Recursive 递归函数的编译
func Test_CompilerInternals_Function_Recursive(t *testing.T) {
	script := compileScript(t, "fn fib(n) {\nif n <= 1 {\nreturn n\n}\nreturn fib(n - 1) + fib(n - 2)\n}\n0")
	fn := script.Functions[0]
	if !ciHasOp(fn.Instructions, OpCall) {
		t.Error("递归函数体内应包含OpCall指令")
	}
}

// Test_CompilerInternals_Function_ExternalFn 外部函数#fn声明编译
func Test_CompilerInternals_Function_ExternalFn(t *testing.T) {
	script := compileScript(t, "#fn myFunc(x)=>int\n0")
	if len(script.Externals) != 1 {
		t.Errorf("应有1个外部函数声明, 实际%d", len(script.Externals))
	}
	if script.Externals[0].Name != "myFunc" {
		t.Errorf("外部函数名应为myFunc, 实际%s", script.Externals[0].Name)
	}
}

// Test_CompilerInternals_Function_ExternalCall 外部函数调用生成OpCallBind
func Test_CompilerInternals_Function_ExternalCall(t *testing.T) {
	script := compileScript(t, "#fn myFunc(x)=>int\nmyFunc(42)\n0")
	if !ciHasOp(script.Main.Instructions, OpCallBind) {
		t.Error("外部函数调用应生成OpCallBind指令")
	}
}

// ========== 数组和Map编译测试 ==========

// Test_CompilerInternals_Array_LiteralNew 数组字面量生成OpArrayNew
func Test_CompilerInternals_Array_LiteralNew(t *testing.T) {
	script := compileScript(t, "[1, 2, 3]")
	if !ciHasOp(script.Main.Instructions, OpArrayNew) {
		t.Error("数组字面量应生成OpArrayNew指令")
	}
}

// Test_CompilerInternals_Array_EmptyNew 空数组的OpArrayNew
func Test_CompilerInternals_Array_EmptyNew(t *testing.T) {
	script := compileScript(t, "[]")
	if !ciHasOp(script.Main.Instructions, OpArrayNew) {
		t.Error("空数组应生成OpArrayNew(0)指令")
	}
	for _, inst := range script.Main.Instructions {
		if inst.Op == OpArrayNew {
			if inst.ArgCount < 1 || inst.Args[0] != 0 {
				t.Errorf("空数组OpArrayNew参数应为0")
			}
		}
	}
}

// Test_CompilerInternals_Map_LiteralNew Map字面量生成OpMapNew
func Test_CompilerInternals_Map_LiteralNew(t *testing.T) {
	script := compileScript(t, "{\"a\": 1}")
	if !ciHasOp(script.Main.Instructions, OpMapNew) {
		t.Error("Map字面量应生成OpMapNew指令")
	}
}

// Test_CompilerInternals_Builtin_PushComp push内置函数编译
func Test_CompilerInternals_Builtin_PushComp(t *testing.T) {
	script := compileScript(t, "arr := [1, 2]\npush(arr, 3)\n0")
	if !ciHasOp(script.Main.Instructions, OpPush) {
		t.Error("push(arr, 3)应生成OpPush指令")
	}
}

// Test_CompilerInternals_Builtin_DeleteComp delete内置函数编译
func Test_CompilerInternals_Builtin_DeleteComp(t *testing.T) {
	script := compileScript(t, "m := {\"a\": 1}\ndelete(m, \"a\")\n0")
	if !ciHasOp(script.Main.Instructions, OpDelete) {
		t.Error("delete应生成OpDelete指令")
	}
}

// Test_CompilerInternals_Map_RangeKeys Map range的OpMapKeys
func Test_CompilerInternals_Map_RangeKeys(t *testing.T) {
	script := compileScript(t, "for k := range {\"a\": 1, \"b\": 2} {\nprint(k)\n}\n0")
	if !ciHasOp(script.Main.Instructions, OpMapKeys) {
		t.Error("Map range应生成OpMapKeys指令")
	}
}

// ========== 编译优化与行为测试 ==========

// Test_CompilerInternals_NoConstFolding 无常量折叠
func Test_CompilerInternals_NoConstFolding(t *testing.T) {
	script := compileScript(t, "1 + 2")
	if !ciHasOp(script.Main.Instructions, OpAdd) {
		t.Error("编译器未做常量折叠, 1+2应保留OpAdd指令")
	}
}

// Test_CompilerInternals_NoDeadCodeElimination 无死代码消除
func Test_CompilerInternals_NoDeadCodeElimination(t *testing.T) {
	script := compileScript(t, "fn f() {\nreturn 1\n}\n0")
	fn := script.Functions[0]
	retCount := ciCountOp(fn.Instructions, OpReturn)
	if retCount < 2 {
		t.Errorf("无死代码消除, OpReturn应>=2, 实际%d", retCount)
	}
}

// Test_CompilerInternals_LiteralOpcodes nil/true/false使用OpConst而非专用操作码
func Test_CompilerInternals_LiteralOpcodes(t *testing.T) {
	t.Run("nil字面量使用OpConst", func(t *testing.T) {
		script := compileScript(t, "nil")
		if !ciHasOp(script.Main.Instructions, OpConst) {
			t.Error("nil字面量应通过OpConst加载")
		}
	})
	t.Run("true字面量使用OpConst", func(t *testing.T) {
		script := compileScript(t, "true")
		if !ciHasOp(script.Main.Instructions, OpConst) {
			t.Error("true字面量应通过OpConst加载")
		}
	})
	t.Run("false字面量使用OpConst", func(t *testing.T) {
		script := compileScript(t, "false")
		if !ciHasOp(script.Main.Instructions, OpConst) {
			t.Error("false字面量应通过OpConst加载")
		}
	})
}

// Test_CompilerInternals_ImplicitNilPush 非表达式语句末尾隐式压入nil
func Test_CompilerInternals_ImplicitNilPush(t *testing.T) {
	script := compileScript(t, "x := 10")
	if !ciHasOp(script.Main.Instructions, OpNil) {
		t.Error("非表达式语句作为最后语句应隐式生成OpNil")
	}
}

// ========== 编译错误测试 ==========

// Test_CompilerInternals_Error_UndefinedVar 未定义变量
func Test_CompilerInternals_Error_UndefinedVar(t *testing.T) {
	runErrorTest(t, "undefinedVar")
}

// Test_CompilerInternals_Error_BreakOutsideLoop break在循环外
func Test_CompilerInternals_Error_BreakOutsideLoop(t *testing.T) {
	runErrorTest(t, "break")
}

// Test_CompilerInternals_Error_ContinueOutsideLoop continue在循环外
func Test_CompilerInternals_Error_ContinueOutsideLoop(t *testing.T) {
	runErrorTest(t, "continue")
}

// Test_CompilerInternals_Error_UnknownBuiltin 未知内置函数
func Test_CompilerInternals_Error_UnknownBuiltin(t *testing.T) {
	runErrorTest(t, "unknownFunc(1)")
}

// Test_CompilerInternals_Error_BuiltinWrongArgs 内置函数参数数量错误
func Test_CompilerInternals_Error_BuiltinWrongArgs(t *testing.T) {
	t.Run("len参数过多", func(t *testing.T) {
		runErrorTest(t, "len(1, 2)")
	})
	t.Run("push参数过少", func(t *testing.T) {
		runErrorTest(t, "push([1])")
	})
	t.Run("delete参数过多", func(t *testing.T) {
		runErrorTest(t, "delete({\"a\": 1}, \"a\", \"b\")")
	})
}

// ========== 指令序列验证测试 ==========

// Test_CompilerInternals_InstSeq_SimpleScript 简单脚本的指令数量合理
func Test_CompilerInternals_InstSeq_SimpleScript(t *testing.T) {
	script := compileScript(t, "42")
	constCount := ciCountOp(script.Main.Instructions, OpConst)
	if constCount != 1 {
		t.Errorf("单字面量42应只有1条OpConst, 实际%d", constCount)
	}
}

// Test_CompilerInternals_InstSeq_ExprStmtPop 表达式语句生成OpPop
func Test_CompilerInternals_InstSeq_ExprStmtPop(t *testing.T) {
	script := compileScript(t, "1 + 2\n3")
	popCount := ciCountOp(script.Main.Instructions, OpPop)
	if popCount < 1 {
		t.Error("非最后的表达式语句应生成OpPop弹出结果")
	}
}

// Test_CompilerInternals_InstSeq_LastExprNoPop 最后一个表达式语句不生成OpPop
func Test_CompilerInternals_InstSeq_LastExprNoPop(t *testing.T) {
	script := compileScript(t, "42")
	popCount := ciCountOp(script.Main.Instructions, OpPop)
	if popCount != 0 {
		t.Errorf("最后的表达式语句不应生成OpPop, 实际%d", popCount)
	}
}

// Test_CompilerInternals_InstSeq_FunctionMainInstructions 函数定义生成主函数指令
func Test_CompilerInternals_InstSeq_FunctionMainInstructions(t *testing.T) {
	script := compileScript(t, "fn f() {\n1\n}\n0")
	if !ciHasOp(script.Main.Instructions, OpStoreLocal) {
		t.Error("函数定义应在主函数中生成OpStoreLocal")
	}
}

// Test_CompilerInternals_InstSeq_IfConditionOrder 条件分支的指令顺序
func Test_CompilerInternals_InstSeq_IfConditionOrder(t *testing.T) {
	script := compileScript(t, "if true { 1 } else { 2 }")
	insts := script.Main.Instructions
	if len(insts) < 5 {
		t.Fatalf("if-else应至少5条指令, 实际%d", len(insts))
	}
	if insts[1].Op != OpJumpIfFalse {
		t.Errorf("第2条指令应为OpJumpIfFalse, 实际%v", insts[1].Op)
	}
}

// ========== NumLocals验证测试 ==========

// Test_CompilerInternals_NumLocals_NoVar 无变量时NumLocals为0
func Test_CompilerInternals_NumLocals_NoVar(t *testing.T) {
	script := compileScript(t, "1 + 2")
	if script.Main.NumLocals != 0 {
		t.Errorf("无变量声明NumLocals应为0, 实际%d", script.Main.NumLocals)
	}
}

// Test_CompilerInternals_NumLocals_SingleVar 单个变量的NumLocals
func Test_CompilerInternals_NumLocals_SingleVar(t *testing.T) {
	script := compileScript(t, "x := 1\nx")
	if script.Main.NumLocals != 1 {
		t.Errorf("单个变量NumLocals应为1, 实际%d", script.Main.NumLocals)
	}
}

// Test_CompilerInternals_NumLocals_StandardFor 标准for循环增加NumLocals
func Test_CompilerInternals_NumLocals_StandardFor(t *testing.T) {
	script := compileScript(t, "for i := 0; i < 3; i = i + 1 {\nprint(i)\n}\n0")
	if script.Main.NumLocals < 1 {
		t.Errorf("标准for循环NumLocals应>=1, 实际%d", script.Main.NumLocals)
	}
}

// Test_CompilerInternals_NumLocals_FunctionLocals 函数内局部变量
func Test_CompilerInternals_NumLocals_FunctionLocals(t *testing.T) {
	script := compileScript(t, "fn f(a) {\nb := a + 1\nc := b * 2\nreturn c\n}\n0")
	fn := script.Functions[0]
	if fn.NumLocals < 3 {
		t.Errorf("函数NumLocals应>=3, 实际%d", fn.NumLocals)
	}
}

// Test_CompilerInternals_NumLocals_CountFor 数组range增加临时变量
func Test_CompilerInternals_NumLocals_CountFor(t *testing.T) {
	script := compileScript(t, "for v := range [1, 2, 3] {\nprint(v)\n}\n0")
	if script.Main.NumLocals < 4 {
		t.Errorf("数组range应分配>=4个局部变量, 实际%d", script.Main.NumLocals)
	}
}

// Test_CompilerInternals_FunctionCount 多函数编译验证
func Test_CompilerInternals_FunctionCount(t *testing.T) {
	script := compileScript(t, "fn f1() {\n1\n}\nfn f2() {\n2\n}\nfn f3() {\n3\n}\n0")
	if len(script.Functions) != 3 {
		t.Errorf("应编译3个函数, 实际%d", len(script.Functions))
	}
}

// Test_CompilerInternals_ExternalCount 多外部函数声明验证
func Test_CompilerInternals_ExternalCount(t *testing.T) {
	script := compileScript(t, "#fn f1(x)=>int\n#fn f2(x)=>string\n#fn f3()=>bool\n0")
	if len(script.Externals) != 3 {
		t.Errorf("应声明3个外部函数, 实际%d", len(script.Externals))
	}
}
