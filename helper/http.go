package helper

import (
    "errors"
    "io"
    "net/http"
)

var (
    // JsonUnmarshal 需要手动设置 JsonUnmarshal=func(r,v){return json.NewDecoder(r).Decode(v)}
    JsonUnmarshal func(io.Reader, any) error
)

func JsonResponseUnmarshal[T any](r *http.Response) (*T, error) {
    if r.Body == nil {
        return nil, errors.New("body is nil")
    }
    defer r.Body.Close()
    t := new(T)
    return t, JsonUnmarshal(r.Body, &t)
}

func JsonRequestUnmarshal[T any](req *http.Request) (*T, error) {
    r, err := http.DefaultClient.Do(req)
    if err != nil {
        return nil, err
    }
    return JsonResponseUnmarshal[T](r)
}
