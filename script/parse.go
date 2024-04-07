package script

import (
    "errors"
)

type _scriptType int8

const (
    typeCommend _scriptType = iota
    typeCompute
    typeIf
    typeFor
)

type simpleEngin struct {
    variable map[string]any
    tree     expression
}

type parseTree struct {
    offset     int
    end        int
    scriptType _scriptType
    script     []rune
    parent     *parseTree
    child      []*parseTree
}

type parser struct {
    engin   *simpleEngin
    tree    parseTree
    script  string
    start   int //启动位置
    offset  int //当前位置
    lOffset int //已经解析完成的位置
    parent  *parser
}

func (p *parser) parse(script []rune) (Engine, error) {
    //拆分为语句树
    for i := p.start; i < len(script); i++ {
        p.offset = i
        switch script[i] {
        case '/':
            //注释,只要判断后一个字符也是/,当前行为注释
            if i+1 < len(script) && script[i+1] == '/' {
                e := &parseTree{
                    offset:     i,
                    scriptType: typeCommend,
                    parent:     &p.tree,
                }
                for i < len(script) && script[i] != '\n' {
                    i++
                }
                i--
                p.lOffset = i
                e.end = i
                e.script = script[e.offset : e.end+1]
                p.tree.child = append(p.tree.child, e)
            }
        case '\n': //换行,换行符约等于;
            //判断前一个字符是否为 ',{('等非结束符
            if i-1 >= 0 && script[i-1] != ',' && script[i-1] != '{' && script[i-1] != '(' {
                e := &parseTree{
                    offset:     i,
                    scriptType: typeCompute,
                    parent:     &p.tree,
                }
                e.script = script[p.lOffset : e.end+1]
                e.end = i
                p.lOffset = i
                p.tree.child = append(p.tree.child, e)
            }
        case ';': //强制结束语句
            e := &parseTree{
                offset:     i,
                scriptType: typeCompute,
                parent:     &p.tree,
            }
            e.script = script[p.lOffset : e.end+1]
            e.end = i
            p.lOffset = i
            p.tree.child = append(p.tree.child, e)
        case '{': //子语句开始
            p2 := &parser{
                engin: p.engin,
                tree: parseTree{
                    offset:     i,
                    scriptType: typeCompute,
                    parent:     &p.tree,
                },
                script:  "",
                start:   i,
                offset:  i,
                lOffset: i,
                parent:  p,
            }
            _, err := p2.parse(script)
            if err != nil {
                return nil, err
            }
            p.tree.child = append(p.tree.child, &p2.tree)
        case '}': //子语句结束
            return p.engin, nil
        }
    }
    if p.lOffset < len(script)-1 {
        e := &parseTree{
            offset:     p.lOffset,
            scriptType: typeCompute,
            parent:     &p.tree,
        }
        e.script = script[p.lOffset:]
        e.end = len(script) - 1
        p.tree.child = append(p.tree.child, e)
    }
    //todo: 语句分析
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
