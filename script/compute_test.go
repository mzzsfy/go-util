package script

import (
    "fmt"
    "github.com/mzzsfy/go-util/seq"
    "sync/atomic"
    "testing"
)

func TestParse1(t *testing.T) {
    exprs := []string{
        "3+3",
        "3+3-3",
        "3+3*3",
        "(3+3)*3",
        "(3+3)*3+3",
        "(3+3)*(3+3)",
        "((3+3)+(3+3))*3",
        "3*((3+3)+(3+3))",
        "3-((3+3)+(3+3))",
        "((3+3)+(3+3))-3",
        "((3+3)+(3+3)-3)+3",
        "3+((3+3)+(3+3)-3)",
        "3+((3+3)+(3+3)-3)*3",
        "3+((3+3)+(3+3)-3)*3",
        "3+3+3+(3-(3+3)+(3+3)-3)*3",
        "3*3+3+(3-(3+3)+(3+3)-3)*3",
        "3*3+3*(3-(3+3)+(3+3)-3)*3",
        "3*3*3+((3+3)+(3+3)-3)*3",
        "3*3+3*((3+3)+(3+3)-3)*3+3*3",
        "3*3+3*((3+3)+(3+3)-3)*(3+3)*3",
    }
    idGen := int32(0)
    exprs = append(exprs, fmt.Sprintf(
        "-%d+%d+((-(%d*%d))+(%d/((%d+%d)*%d))*(%d/%d+%d%%%d*(%d*%d*%d-%d+%d-%d%%%d-%d-%d)+%d+(%d-%d)*%d)-%d)+%d*%d/%d+%d+%d",
        atomic.AddInt32(&idGen, 1),
        atomic.AddInt32(&idGen, 1),
        atomic.AddInt32(&idGen, 1),
        atomic.AddInt32(&idGen, 1),
        atomic.AddInt32(&idGen, 1),
        atomic.AddInt32(&idGen, 1),
        atomic.AddInt32(&idGen, 1),
        atomic.AddInt32(&idGen, 1),
        atomic.AddInt32(&idGen, 1),
        atomic.AddInt32(&idGen, 1),
        atomic.AddInt32(&idGen, 1),
        atomic.AddInt32(&idGen, 1),
        atomic.AddInt32(&idGen, 1),
        atomic.AddInt32(&idGen, 1),
        atomic.AddInt32(&idGen, 1),
        atomic.AddInt32(&idGen, 1),
        atomic.AddInt32(&idGen, 1),
        atomic.AddInt32(&idGen, 1),
        atomic.AddInt32(&idGen, 1),
        atomic.AddInt32(&idGen, 1),
        atomic.AddInt32(&idGen, 1),
        atomic.AddInt32(&idGen, 1),
        atomic.AddInt32(&idGen, 1),
        atomic.AddInt32(&idGen, 1),
        atomic.AddInt32(&idGen, 1),
        atomic.AddInt32(&idGen, 1),
        atomic.AddInt32(&idGen, 1),
        atomic.AddInt32(&idGen, 1),
        atomic.AddInt32(&idGen, 1),
        atomic.AddInt32(&idGen, 1),
        atomic.AddInt32(&idGen, 1),
    ))
    exprs = append(exprs, fmt.Sprintf(
        "-%d+%d+(-(%d/(%d+%d)*%d)*(%d/%d+%d%%%d*(%d*%d*%d-%d+%d-%d%%%d-%d-%d)+%d+(%d-%d)*%d)-%d)+%d*%d/%d+%d+%d",
        atomic.AddInt32(&idGen, 1),
        atomic.AddInt32(&idGen, 1),
        atomic.AddInt32(&idGen, 1),
        atomic.AddInt32(&idGen, 1),
        atomic.AddInt32(&idGen, 1),
        atomic.AddInt32(&idGen, 1),
        atomic.AddInt32(&idGen, 1),
        atomic.AddInt32(&idGen, 1),
        atomic.AddInt32(&idGen, 1),
        atomic.AddInt32(&idGen, 1),
        atomic.AddInt32(&idGen, 1),
        atomic.AddInt32(&idGen, 1),
        atomic.AddInt32(&idGen, 1),
        atomic.AddInt32(&idGen, 1),
        atomic.AddInt32(&idGen, 1),
        atomic.AddInt32(&idGen, 1),
        atomic.AddInt32(&idGen, 1),
        atomic.AddInt32(&idGen, 1),
        atomic.AddInt32(&idGen, 1),
        atomic.AddInt32(&idGen, 1),
        atomic.AddInt32(&idGen, 1),
        atomic.AddInt32(&idGen, 1),
        atomic.AddInt32(&idGen, 1),
        atomic.AddInt32(&idGen, 1),
        atomic.AddInt32(&idGen, 1),
        atomic.AddInt32(&idGen, 1),
        atomic.AddInt32(&idGen, 1),
        atomic.AddInt32(&idGen, 1),
        atomic.AddInt32(&idGen, 1),
    ))
    for _, exprStr := range exprs {
        t.Log(exprStr)
        s1 := fmt.Sprint(seq.From(func(t func(any)) {
            node := parseNumberExpr([]rune(exprStr), 0).stack
            for {
                if node.next == nil {
                    break
                }
                t(node)
                node = node.next
            }
        }).ToSlice())
        s2 := fmt.Sprint(NifixToRPN([]byte(exprStr)))
        t.Log(s1)
        t.Log(s2)
        //if s1 != s2 {
        //    t.Error("not equal")
        //}
    }
}

//与NifixToRPN 递归消除括号,返回对应后括号的位置
func barket(a []byte, first int) ([]string, int) {
    last := 0  //对应后括号的位置
    count := 0 //计数左括号
    for i := first; i < len(a)-1; i++ {
        if a[i+1] == '(' {
            count++
        }
        if a[i+1] == ')' {
            if count == 0 {
                last = i + 1
                break
            } else {
                count--
            }
        }
    }
    NewA := a[first+1 : last]
    return NifixToRPN(NewA), last //递归
}

//中缀表达式转后缀表达式
//仅仅支持含 + - * / ( ) 与 整数 组合的表达式(可以支持负数如(-2))
func NifixToRPN(nifix []byte) []string {
    operatorStack := make([]byte, 0, 20) //存放运算符的栈(切片模拟栈)
    ret := make([]string, 0, len(nifix)) //返回的逆波兰表达式
    //把'+' '-'看成一类，优先级低于 '*' '/'
    priority := map[byte]int{
        '+': 0,
        '-': 0,
        '*': 1,
        '/': 1,
        '%': 1,
    }
    i := 0
    //处理第一个数为负数的情况
    if nifix[0] == '-' {
        i = 1
        tempI := i
        //处理'-'后的数字加入ret
        for tempI+1 < len(nifix) && nifix[tempI+1] >= '0' && nifix[tempI+1] <= '9' {
            tempI++
        }
        ret = append(ret, "-"+string(nifix[i:tempI+1]))
        i = tempI + 1
    }
    for ; i < len(nifix); i++ {
        //将连续的数字写入ret
        if nifix[i] >= '0' && nifix[i] <= '9' {
            tempI := i
            for tempI+1 < len(nifix) && nifix[tempI+1] >= '0' && nifix[tempI+1] <= '9' {
                tempI++
            }
            ret = append(ret, string(nifix[i:tempI+1]))
            i = tempI
            continue
        }
        //把括号里面的内容当成一个数字，将得到的逆波兰表达式直接写入ret
        if nifix[i] == '(' {
            //调用barket消除括号
            newNifix, last := barket(nifix, i)
            ret = append(ret, newNifix...)
            i = last
            continue
        }
        //栈中没有运算符，直接将当前运算符入栈
        if len(operatorStack) == 0 {
            operatorStack = append(operatorStack, nifix[i])
            continue
        }
        //如果当前符号的优先级 小于 上一个未出栈符号优先级，则将栈清空，当前符号入栈
        if priority[nifix[i]] < priority[operatorStack[len(operatorStack)-1]] {
            //将栈清空，写入ret
            for len(operatorStack) != 0 {
                ret = append(ret, string(operatorStack[len(operatorStack)-1]))
                operatorStack = operatorStack[0 : len(operatorStack)-1]
            }
            //当前符号入栈
            operatorStack = append(operatorStack, nifix[i])
            continue
        }
        //如果当前符号的优先级 等于 上一个未出栈符号优先级，栈中优先级相同的连续符号全部出栈，当前符号入栈
        if priority[nifix[i]] == priority[operatorStack[len(operatorStack)-1]] {
            for len(operatorStack) != 0 && priority[nifix[i]] == priority[operatorStack[len(operatorStack)-1]] {
                ret = append(ret, string(operatorStack[len(operatorStack)-1]))
                operatorStack = operatorStack[0 : len(operatorStack)-1]
            }
            //当前符号入栈
            operatorStack = append(operatorStack, nifix[i])
            continue
        }
        //此时栈中还有符号，且当前符号优先级高于栈中符号,将当前符号入栈
        operatorStack = append(operatorStack, nifix[i])
    }
    //还有未出栈的符号，依次出栈
    if len(operatorStack) != 0 {
        if len(operatorStack) == 2 {
            ret = append(ret, string(operatorStack[1]))
        }
        ret = append(ret, string(operatorStack[0]))
    }
    return ret
}
