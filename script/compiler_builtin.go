package script

// ========== 内置函数编译 ==========
// 包含所有内置函数的编译逻辑和注册机制

// builtinCompiler 内置函数编译器类型
// 每个内置函数可以有自己的编译逻辑
type builtinCompiler func(c *Compiler, args []Expr) error

// builtinCompilers 内置函数编译器映射
// 将内置函数名映射到其编译器函数
var builtinCompilers map[string]builtinCompiler

// builtinSpec 内置函数规格
// 使用声明式配置统一管理内置函数
type builtinSpec struct {
	name     string
	compiler builtinCompiler
	op       OpCode
}

// registerBuiltins 注册所有内置函数
// 使用声明式配置降低复杂度，便于维护
func registerBuiltins() {
	// 定义所有内置函数规格
	// compiler非nil表示特殊处理，否则使用op创建单参数处理器
	specs := []builtinSpec{
		// 特殊处理函数
		{name: "print", compiler: makePrintCompiler(false)},
		{name: "println", compiler: makePrintCompiler(true)},
		{name: "delete", compiler: compileBuiltinDelete},
		{name: "push", compiler: compileBuiltinPush},
		{name: "getBindValue", compiler: compileBuiltinGetBindValue},
		// 单参数类型转换/操作函数
		{name: "len", op: OpLen},
		{name: "typeof", op: OpTypeOf},
		{name: "int", op: OpToInt},
		{name: "float", op: OpToFloat},
		{name: "string", op: OpToString},
	}

	builtinCompilers = make(map[string]builtinCompiler, len(specs))
	for _, spec := range specs {
		if spec.compiler != nil {
			builtinCompilers[spec.name] = spec.compiler
		} else {
			builtinCompilers[spec.name] = makeBuiltinUnary(spec.op, spec.name)
		}
	}
}

// init 初始化内置函数编译器映射
func init() {
	registerBuiltins()
}

// makePrintCompiler 创建print/println编译器
func makePrintCompiler(newline bool) builtinCompiler {
	return func(c *Compiler, args []Expr) error {
		if err := c.compileArgs(args); err != nil {
			return err
		}
		op := OpPrint
		if newline {
			op = OpPrintln
		}
		c.emit1(op, len(args))
		return nil
	}
}

// compileBuiltinPush 编译 push() 调用
func compileBuiltinPush(c *Compiler, args []Expr) error {
	// 检查参数数量
	if len(args) != 2 {
		if len(args) > 0 {
			return NewCompileErrorFromPos(args[0],
				"参数数量错误：push() 需要2个参数，但传入了%d个。\n"+
					"→ 正确用法：push(array, element)\n"+
					"→ 示例：\n"+
					"  - push([1, 2], 3)  // 正确\n"+
					"  - push(arr, 42)    // 正确\n"+
					"→ 功能：向数组末尾追加元素，返回新数组（不修改原数组）",
				len(args))
		}
		return NewCompileError(0, 0,
			"参数数量错误：push() 需要2个参数。\n"+
				"→ 正确用法：push(array, element)\n"+
				"→ 示例：push([1, 2], 3)")
	}

	// 编译数组参数
	if err := c.compileExpr(args[0]); err != nil {
		return err
	}
	// 编译元素参数
	if err := c.compileExpr(args[1]); err != nil {
		return err
	}
	c.emit(OpPush)
	return nil
}

// compileBuiltinDelete 编译 delete() 调用
func compileBuiltinDelete(c *Compiler, args []Expr) error {
	// 检查参数数量
	if len(args) != 2 {
		if len(args) > 0 {
			return NewCompileErrorFromPos(args[0],
				"参数数量错误：delete() 需要2个参数，但传入了%d个。\n"+
					"→ 正确用法：delete(map, key)\n"+
					"→ 示例：\n"+
					"  - delete(m, \"key\")  // 正确\n"+
					"  - delete(m, k)       // 正确\n"+
					"→ 功能：从map中删除指定的key",
				len(args))
		}
		return NewCompileError(0, 0,
			"参数数量错误：delete() 需要2个参数。\n"+
				"→ 正确用法：delete(map, key)\n"+
				"→ 示例：delete(m, \"key\")")
	}

	// 编译map参数
	if err := c.compileExpr(args[0]); err != nil {
		return err
	}
	// 编译key参数
	if err := c.compileExpr(args[1]); err != nil {
		return err
	}
	c.emit(OpDelete)
	return nil
}

// compileBuiltinGetBindValue 编译 getBindValue() 调用
// 强制要求必须在类型安全声明（:=>）中使用
func compileBuiltinGetBindValue(c *Compiler, args []Expr) error {
	// 检查是否在类型安全声明中
	if !c.inTypedDeclaration {
		// 使用第一个参数的位置信息（如果存在）
		if len(args) > 0 {
			return NewCompileErrorFromPos(args[0],
				"类型错误：getBindValue 必须在类型安全声明中使用。\n"+
					"→ 问题：未使用 :=> 语法指定类型。\n"+
					"→ 原因：getBindValue 从 Go 代码获取值，必须显式指定类型以确保类型安全。\n"+
					"→ 正确用法：x :=>int getBindValue(\"name\")\n"+
					"→ 正确用法：x :=>string getBindValue(\"name\")\n"+
					"→ 正确用法：x :=>any getBindValue(\"name\")\n"+
					"→ 错误用法：x := getBindValue(\"name\")\n"+
					"→ 建议：使用 :=>type 语法明确指定返回值类型")
		}
		return NewCompileError(0, 0,
			"类型错误：getBindValue 必须在类型安全声明中使用。\n"+
				"→ 正确用法：x :=>int getBindValue(\"name\")\n"+
				"→ 建议：使用 :=>type 语法明确指定返回值类型")
	}

	// 检查参数数量
	if len(args) != 1 {
		if len(args) > 1 {
			return NewCompileErrorFromPos(args[1],
				"参数数量错误：getBindValue() 需要1个参数，但传入了%d个。\n"+
					"→ 正确用法：getBindValue(\"name\")\n"+
					"→ 示例：\n"+
					"  - getBindValue(\"x\")  // 正确\n"+
					"  - getBindValue(\"x\", \"y\")  // 错误：参数过多\n"+
					"→ 功能：从 Go 绑定值中获取指定名称的值",
				len(args))
		}
		return NewCompileError(0, 0,
			"参数数量错误：getBindValue() 需要1个参数。\n"+
				"→ 正确用法：getBindValue(\"name\")\n"+
				"→ 示例：getBindValue(\"x\")")
	}

	// 编译参数
	if err := c.compileExpr(args[0]); err != nil {
		return err
	}

	// 发送加载绑定值指令
	c.emit(OpLoadBind)
	return nil
}

// makeBuiltinUnary 创建单参数内置函数编译器
func makeBuiltinUnary(op OpCode, name string) builtinCompiler {
	return func(c *Compiler, args []Expr) error {
		if len(args) != 1 {
			// 使用第一个参数的位置信息（如果存在）
			if len(args) > 0 {
				return NewCompileErrorFromPos(args[0],
					"参数数量错误：%s() 需要1个参数，但传入了%d个。\n"+
						"→ 正确用法：%s(value)\n"+
						"→ 示例：\n"+
						"  - %s(123)  // 正确\n"+
						"  - %s(x, y)  // 错误：参数过多\n"+
						"→ 建议：检查函数调用，确保只传入一个参数",
					name, len(args), name, name, name)
			}
			return NewCompileError(0, 0,
				"参数数量错误：%s() 需要1个参数，但没有传入任何参数。\n"+
					"→ 正确用法：%s(value)\n"+
					"→ 示例：%s(123)\n"+
					"→ 建议：在函数调用时传入一个参数",
				name, name, name)
		}
		if err := c.compileExpr(args[0]); err != nil {
			return err
		}
		c.emit(op)
		return nil
	}
}

// compileBuiltinCall 编译内置函数调用
func (c *Compiler) compileBuiltinCall(name string, args []Expr) error {
	compiler, ok := builtinCompilers[name]
	if !ok {
		// 使用第一个参数的位置信息（如果存在）
		if len(args) > 0 {
			return NewCompileErrorFromPos(args[0],
				"未知的内置函数：'%s'。\n"+
					"→ 问题：该函数名不是内置函数。\n"+
					"→ 可能的原因：\n"+
					"  - 函数名拼写错误\n"+
					"  - 试图调用未声明的函数\n"+
					"  - 使用了不支持的功能\n"+
					"→ 支持的内置函数：\n"+
					"  - len(value)：获取长度（数组/字符串/Map）\n"+
					"  - typeof(value)：获取类型名称\n"+
					"  - int(value)：转换为整数\n"+
					"  - float(value)：转换为浮点数\n"+
					"  - string(value)：转换为字符串\n"+
					"  - print(values...)：打印到标准输出\n"+
					"  - println(values...)：打印并换行\n"+
					"  - getBindValue(name)：获取绑定值\n"+
					"→ 外部函数：使用 #fn 声明后可通过 context.BindFunc() 绑定\n"+
					"→ 建议：检查函数名拼写或使用 #fn 声明外部函数",
				name)
		}
		return NewCompileError(0, 0,
			"未知的内置函数：'%s'。\n"+
				"→ 支持的内置函数：len, typeof, int, float, string, print, println, getBindValue\n"+
				"→ 建议：检查函数名拼写或使用 #fn 声明外部函数",
			name)
	}
	return compiler(c, args)
}

// isBuiltin 检查给定名称是否为内置函数
// 直接使用builtinCompilers映射表检查，避免维护两份数据
func (c *Compiler) isBuiltin(name string) bool {
	_, ok := builtinCompilers[name]
	return ok
}

// getExternalIndex 获取外部函数在列表中的索引
// 优先使用缓存进行O(1)查找
func (c *Compiler) getExternalIndex(name string) int {
	if idx, ok := c.externalCache[name]; ok {
		return idx
	}
	// 缓存未命中时遍历查找
	for i, ext := range c.externals {
		if ext.Name == name {
			c.externalCache[name] = i
			return i
		}
	}
	return -1
}
