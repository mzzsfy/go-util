package helper

import "sync"

type Node[T any] interface {
    // Next 获取子节点,len表示消耗的字节长度, over 表示是否是一个结束节点
    Next([]byte) (node Node[T])
    AddNext([]byte, T)
    Value() T
}

type ByteNode[T any] struct {
    v    T
    l    int
    next map[string]Node[T]
}

func (b *ByteNode[T]) Next(bytes []byte) Node[T] {
    if len(bytes) == 0 {
        return b.next[""]
    }
    if len(bytes) <= b.l {
        node1, b1 := b.next[BytesToString(bytes)]
        if !b1 {
            return nil
        }
        return node1.Next(nil)
    }
    node1, b1 := b.next[BytesToString(bytes[:b.l])]
    if !b1 {
        return nil
    }
    return node1.Next(bytes[b.l:])
}

func (b *ByteNode[T]) AddNext(bytes []byte, v T) {
    if len(bytes) == 0 {
        b.next[""] = &ByteNode[T]{
            l:    b.l,
            next: make(map[string]Node[T]),
            v:    v,
        }
        return
    }
    if len(bytes) <= b.l {
        s := BytesToString(bytes)
        if n, ok := b.next[s]; ok {
            n.AddNext(nil, v)
        } else {
            b2 := &ByteNode[T]{
                l:    b.l,
                next: make(map[string]Node[T]),
            }
            b2.AddNext(nil, v)
            b.next[s] = b2
        }
        return
    }
    var (
        n  Node[T]
        ok bool
    )
    s := BytesToString(bytes[:b.l])
    if n, ok = b.next[s]; !ok {
        n = &ByteNode[T]{
            l:    b.l,
            next: make(map[string]Node[T]),
        }
        b.next[s] = n
    }
    n.AddNext(bytes[b.l:], v)
}

func (b *ByteNode[T]) Value() T {
    return b.v
}

// MakeNewDfsNode 返回一个创建一个新的节点的方法, size 表示每个节点消耗的字节长度,建议2~8之间
func MakeNewDfsNode[T any](size int) func([]byte, T) Node[T] {
    if size <= 0 {
        panic("size must be greater than 0")
    }
    return func(data []byte, v T) Node[T] {
        b := &ByteNode[T]{
            l:    size,
            next: make(map[string]Node[T]),
        }
        if len(data) > size {
            b.AddNext(data[size:], v)
        } else {
            b.v = v
        }
        return b
    }
}

//Dfa Deterministic Finite Automaton
type Dfa[T any] struct {
    lock sync.Mutex
    root Node[T]
}

func (d *Dfa[T]) Add(data []byte, v T) {
    d.lock.Lock()
    defer d.lock.Unlock()
    if len(data) == 0 {
        panic("data length must be greater than 0")
    }
    d.root.AddNext(data, v)
}
func (d *Dfa[T]) AddSimple(data []byte) {
    var v T
    d.Add(data, v)
}

func (d *Dfa[T]) AddStr(data string, v T) {
    d.Add(StringToBytes(data), v)
}

func (d *Dfa[T]) AddStrSimple(data string) {
    var v T
    d.AddStr(data, v)
}

func (d *Dfa[T]) Test(data []byte) Node[T] {
    if len(data) == 0 {
        return d.root
    }
    return d.root.Next(data)
}

func (d *Dfa[T]) TestStr(data string) Node[T] {
    if len(data) == 0 {
        return d.root
    }
    return d.root.Next(StringToBytes(data))
}

// NewDfa 创建一个Dfa,配合 MakeNewDfsNode 使用
func NewDfa[T any](newNode func([]byte, T) Node[T]) *Dfa[T] {
    var v T
    return &Dfa[T]{
        root: newNode(nil, v),
    }
}
