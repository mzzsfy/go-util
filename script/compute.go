package script

import (
    "fmt"
    "github.com/mzzsfy/go-util/helper"
    "strconv"
)

type compute interface {
    expression
}

func operatorPriority(r rune) uint8 {
    switch r {
    case '(', ')':
        return 1
    //case '!':
    //    return 11
    case '*', '/', '%':
        return 21
    case '+', '-':
        return 31
    //case '^', '|':
    //    return 41
    //case '&':
    //    return 44
    default:
        return 0
    }
}

type numberNode struct {
    offset int
    value  variable
    expr   byte //运算符
    next   *numberNode
}

func (p numberNode) String() string {
    if p.expr > 0 {
        return string(p.expr)
    }
    return fmt.Sprint(p.value)
}

type numberCompute struct {
    offset int
    expr   numberNode // number or variable
}

func (n *numberCompute) compute(scope Scope) (r any) {
    node := n.expr
    defer func() { recoverRunTimeError(recover(), helper.Default(node.offset, n.offset)) }()
    if node.next == nil {
        return node.value.value(scope)
    }
    stack := make([]variable, 0, 3)
    for {
        if node.next == nil {
            if len(stack) != 1 {
                panic("计算结果应该为单个值")
            }
            r = stack[0].value(scope)
            //if f, ok := r.(float64); ok {
            //if math.Trunc(f) != f {
            //    //保留9位小数
            //    n10 := math.Pow10(9)
            //    return math.Trunc((f+0.5/n10)*n10) / n10
            //}
            //}
            return
        }
        //计算
        if node.expr != 0 {
            if len(stack) < 2 {
                panic("表达式错误,参数不足")
            }
            left := stack[len(stack)-2].value(scope)
            right := stack[len(stack)-1].value(scope)
            stack = n.compute1(left, right, &node, stack[:len(stack)-2])
        } else {
            stack = append(stack, node.value)
        }
        node = *node.next
    }
}

// 优化,计算静态表达式,避免重复计算
func (n *numberCompute) optimization() {
    node := n.expr
    stack := make([]variable, 0, 3)
    offsetStack := make([]int, 0, 3)
    defer func() {
        recoverRunTimeError(recover(), helper.Default(node.offset, n.offset))
        for i := len(stack) - 1; i >= 0; i-- {
            n1 := node
            node = numberNode{
                offset: offsetStack[i],
                value:  stack[i],
                next:   &n1,
            }
        }
        n.expr = node
    }()
    for {
        if node.next == nil {
            return
        }
        //计算
        if node.expr != 0 {
            if len(stack) < 2 {
                panic("表达式错误,参数不足")
            }
            //todo: 当遇到非固定值时,跳过参数,尝试继续计算
            if _, ok := stack[len(stack)-2].(fixedValue); !ok {
                return
            }
            if _, ok := stack[len(stack)-1].(fixedValue); !ok {
                return
            }
            left := stack[len(stack)-2].(fixedValue).v
            right := stack[len(stack)-1].(fixedValue).v
            stack = n.compute1(left, right, &node, stack[:len(stack)-2])
            offsetStack = offsetStack[:len(offsetStack)-1]
        } else {
            stack = append(stack, node.value)
            offsetStack = append(offsetStack, node.offset)
        }
        node.next, node = nil, *node.next
    }
}

func (n *numberCompute) compute1(left any, right any, node *numberNode, stack []variable) []variable {
    //1:整数模式
    //2:浮点数模式
    var mode int
    var lf float64
    var rf float64
    var li int
    var ri int
    if _, ok := left.(int); ok {
        li = left.(int)
        if _, ok := right.(int); ok {
            ri = right.(int)
            mode = 1
        } else {
            rf = right.(float64)
            lf = float64(li)
        }
    } else {
        lf = left.(float64)
        if _, ok := right.(int); ok {
            ri = right.(int)
            rf = float64(ri)
        } else {
            rf = right.(float64)
        }
    }
    if mode == 0 {
        mode = 2
    }
    switch node.expr {
    case '*':
        switch mode {
        case 1:
            stack = append(stack, fixedValue{v: li * ri})
        case 2:
            stack = append(stack, fixedValue{v: lf * rf})
        default:
            panic("计算模式错误")
        }
    case '/':
        switch mode {
        case 1:
            //todo:提供选项,整数转换为浮点数后计算除法
            stack = append(stack, fixedValue{v: li / ri})
        case 2:
            stack = append(stack, fixedValue{v: lf / rf})
        default:
            panic("计算模式错误")
        }
    case '%':
        switch mode {
        case 1:
            stack = append(stack, fixedValue{v: li % ri})
        case 2:
            //todo:提供选项,小数转换为整数后取模
            panic("小数不支持模运算")
        default:
            panic("计算模式错误")
        }
    case '+':
        switch mode {
        case 1:
            stack = append(stack, fixedValue{v: li + ri})
        case 2:
            stack = append(stack, fixedValue{v: lf + rf})
        default:
            panic("计算模式错误")
        }
    case '-':
        switch mode {
        case 1:
            stack = append(stack, fixedValue{v: li - ri})
        case 2:
            stack = append(stack, fixedValue{v: lf - rf})
        default:
            panic("计算模式错误")
        }
    default:
        panic("不支持的表达式")
    }
    return stack
}

func parseNumberExpr(expr []rune, offset int) *numberCompute {
    c := &numberCompute{expr: numberNode{offset: offset}}
    parseNumberExpr1(expr, offset, &c.expr)
    return c
}

type parseError struct {
    expr   []rune
    offset int
    err    string
}

func (p parseError) String() string {
    if p.offset > 10 {
        return "parse error: " + string(p.expr[p.offset-10:helper.Min(p.offset+10, len(p.expr))]) + " " + p.err
    }
    return "parse error: " + string(p.expr[:helper.Min(p.offset+10, len(p.expr))]) + " " + p.err
}

func parseNumberExpr1(expr []rune, offset int, c *numberNode) int {
    isVar := false
    decimal := false
    brackets := false
    numbers := make([]rune, 0, 2)
    lastOperator := rune(0)
    var lowOperator []rune
    defer func() {
        if len(numbers) > 0 {
            c = parseNumber(expr, isVar, numbers, c, c.offset, decimal)
            c.next = &numberNode{}
            c = c.next
        }
        if lastOperator != 0 {
            c.offset = len(expr) - 1
            c.expr = byte(lastOperator)
            c.next = &numberNode{}
            c = c.next
        }
        for _, o := range lowOperator {
            c.expr = byte(o)
            c.next = &numberNode{}
            c = c.next
        }
    }()
    for i := offset; i < len(expr); i++ {
        priority := operatorPriority(expr[i])
        if priority > 0 {
            //负数
            if expr[i] == '-' {
                if len(numbers) == 0 && !brackets {
                    numbers = append(numbers, '-')
                    continue
                }
            }
            if expr[i] != '^' && expr[i] != '(' && len(numbers) == 0 && lastOperator == 0 && !brackets {
                panic(parseError{
                    expr:   expr,
                    offset: i,
                    err:    "不正确的表达式: " + string(expr[i]),
                })
            }
            if len(numbers) > 0 {
                //为-(x)格式,指定取负数,强制指定为0-x
                if len(numbers) == 1 && numbers[0] == '-' {
                    c.value = fixedValue{v: 0}
                    isVar = false
                    decimal = false
                    numbers = make([]rune, 0, 2)
                    c.offset = i
                    c.next = &numberNode{}
                    c = c.next
                    if lastOperator != 0 {
                        c.offset = i
                        c.expr = byte(lastOperator)
                        c.next = &numberNode{}
                        c = c.next
                    }
                    lastOperator = '-'
                } else {
                    c = parseNumber(expr, isVar, numbers, c, i, decimal)
                    c.next = &numberNode{}
                    c = c.next
                    decimal = false
                    isVar = false
                    numbers = make([]rune, 0, 2)
                }
            }
            if expr[i] == '(' {
                node := &numberNode{}
                i = parseNumberExpr1(expr, i+1, node)
                c.offset = node.offset
                c.expr = node.expr
                c.value = node.value
                c.next = node.next
                for {
                    c = c.next
                    if c.next == nil {
                        break
                    }
                }
                brackets = true
                continue
            }
            if expr[i] == ')' {
                if lastOperator != 0 {
                    c.offset = i - 1
                    c.expr = byte(lastOperator)
                    c.next = &numberNode{}
                    c = c.next
                }
                lastOperator = 0
                return i
            } else {
                //有以前的运算符
                if lastOperator != 0 {
                    if operatorPriority(expr[i]) < operatorPriority(lastOperator) {
                        lowOperator = append(lowOperator, lastOperator)
                    } else {
                        c.offset = i - 1
                        c.expr = byte(lastOperator)
                        c.next = &numberNode{}
                        c = c.next
                    }
                }
                if len(lowOperator) > 0 && operatorPriority(expr[i]) >= operatorPriority(lowOperator[0]) {
                    c.expr = byte(lowOperator[len(lowOperator)-1])
                    lowOperator = lowOperator[:len(lowOperator)-1]
                    c.next = &numberNode{}
                    c = c.next
                }
            }
            lastOperator = expr[i]
            brackets = false
        } else {
            brackets = false
            if expr[i] == ' ' {
                //下一个字符非空格或运算符
                if len(numbers) != 0 && len(expr) >= i && expr[i+1] != ' ' && operatorPriority(expr[i+1]) == 0 {
                    panic(parseError{
                        expr:   expr,
                        offset: i,
                        err:    "无法处理表达式中间的空格: " + string(expr[helper.Max(i-3, 0):helper.Min(i+3, len(expr))]),
                    })
                }
                continue
            }
            if !isVar {
                if expr[i] == '.' {
                    decimal = true
                    if len(numbers) == 0 {
                        numbers = append(numbers, '0')
                    }
                    numbers = append(numbers, '.')
                } else if expr[i] >= '0' && expr[i] <= '9' {
                    numbers = append(numbers, expr[i])
                } else if ((len(numbers) == 0 || (len(numbers) == 1 && numbers[0] == '-')) && expr[i] >= 'a' && expr[i] <= 'z') || (expr[i] >= 'A' && expr[i] <= 'Z') || expr[i] == '$' || expr[i] == '_' {
                    isVar = true
                    numbers = append(numbers, expr[i])
                } else {
                    panic(parseError{
                        expr:   expr,
                        offset: i,
                        err:    "无法识别的表达式: " + string(expr[helper.Max(i-3, 0):helper.Min(i+3, len(expr))]),
                    })
                }
            } else if (expr[i] >= '0' && expr[i] <= '9') || (expr[i] >= 'a' && expr[i] <= 'z') || (expr[i] >= 'A' && expr[i] <= 'Z') || expr[i] == '$' || expr[i] == '_' {
                numbers = append(numbers, expr[i])
            } else {
                panic(parseError{
                    expr:   expr,
                    offset: i,
                    err:    "无法识别的表达式: " + string(expr[helper.Max(i-3, 0):helper.Min(i+3, len(expr))]),
                })
            }
        }
    }
    return len(expr) - 1
}

func parseNumber(expr []rune, isVar bool, numbers []rune, c *numberNode, i int, decimal bool) *numberNode {
    var v any
    if isVar {
        if numbers[0] == '-' {
            c.offset = i - 1
            c.expr = byte(numbers[0])
            c.next = &numberNode{}
            c = c.next
            c.value = &varValue{
                name:     string(numbers[1:]),
                negative: true,
            }
        } else {
            c.value = &varValue{name: string(numbers)}
        }
    } else {
        if decimal {
            float, err := strconv.ParseFloat(string(numbers), 64)
            if err != nil {
                panic(parseError{
                    expr:   expr,
                    offset: i,
                    err:    "不正确的表达式:" + err.Error(),
                })
            }
            v = float
        } else {
            Int, err := strconv.ParseInt(string(numbers), 10, 64)
            if err != nil {
                panic(parseError{
                    expr:   expr,
                    offset: i,
                    err:    "不正确的表达式:" + err.Error(),
                })
            }
            v = int(Int)
        }
        c.value = fixedValue{v: v}
    }
    c.offset = i - len(numbers)
    return c
}
