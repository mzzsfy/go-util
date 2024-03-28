package script

import (
    "errors"
)

type simpleEngin struct {
    variable map[string]any
    tree     expression
}

type expressionTree struct {
    offset int
    parent *expressionTree
    child  []*expression
}

type parser struct {
    engin   *simpleEngin
    tree    expressionTree
    script  string
    start   int //启动位置
    offset  int //当前位置
    lOffset int //已经解析完成的位置
    parent  *parser
}

func (p *parser) parse(script []rune) (Engine, error) {
    for i, r := range script {
        p.offset = i
        switch r {
        case '+': //加法,字符串拼接
        case '-': //减法,负数
        case '*': //乘法
        case '/': //除法,注释
        case '\n': //换行
        case '%': //取模
        case ';': //强制结束语句
        case ':': //a:=1
        case '?': //a?b?c?d:e:f => (a?(b?(c?x:d):e)
        case '{':
        case '}':
        case '[': //[a,b,c],a[b]
        case ']':
        case '(':
        case ')':
        case '`': //`a`
        case '"': //"a"
        case '\'': //'a'
        case '=': //a=b,a:=b,a==b
        }
    }
    return p.engin, nil
}

func (e *simpleEngin) Execute(m map[string]any) Result {
    return res{
        err: errors.New("not impl"),
    }
}

func (e *simpleEngin) Bind(m map[string]any) (Engine, error) {
    return e, nil
}

type opt struct {
}

type Opt func(*opt)

func Parse(script string, _ ...Opt) (Engine, error) {
    p := &parser{
        engin:   &simpleEngin{},
        script:  script,
        start:   0,
        offset:  0,
        lOffset: 0,
    }
    return p.parse([]rune(script))
}
