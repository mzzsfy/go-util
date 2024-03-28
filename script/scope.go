package script

import "errors"

type globalScope struct {
    v map[string]any
}

func (s globalScope) Get(name string) (any, error) {
    a, ok := s.v[name]
    if !ok {
        return nil, errors.New("没有这个变量:" + name)
    }
    return a, nil
}

func (s globalScope) New(name string, a any) error {
    a, ok := s.v[name]
    if ok {
        return errors.New("重复声明变量:" + name)
    }
    s.v[name] = a
    return nil
}

func (s globalScope) Update(name string, a any) error {
    a, ok := s.v[name]
    if !ok {
        return errors.New("更新不存在的变量:" + name)
    }
    s.v[name] = a
    return nil
}

func NewChildScope(parent Scope) Scope {
    return &childScope{
        parent: parent,
        v:      make(map[string]any),
    }
}

type childScope struct {
    parent Scope
    v      map[string]any
}

func (s childScope) Get(name string) (any, error) {
    a, ok := s.v[name]
    if !ok {
        return s.parent.Get(name)
    }
    return a, nil
}

func (s childScope) New(name string, a any) error {
    s.v[name] = a
    return nil
}

func (s childScope) Update(name string, a any) error {
    a, ok := s.v[name]
    if ok {
        s.v[name] = a
    }
    return s.parent.Update(name, a)
}
