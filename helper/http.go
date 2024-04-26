package helper

import (
    "encoding/json"
    "errors"
    "io"
    "net/http"
)

var (
    JsonUnmarshal = json.Unmarshal
)

func JsonResponseUnmarshal[T any](r *http.Response) (*T, error) {
    if r.Body == nil {
        return nil, errors.New("body is nil")
    }
    defer r.Body.Close()
    all, err := io.ReadAll(r.Body)
    if err != nil {
        return nil, err
    }
    t := new(T)
    err = JsonUnmarshal(all, t)
    return t, err
}

func JsonRequestUnmarshal[T any](req *http.Request) (*T, error) {
    r, err := http.DefaultClient.Do(req)
    if err != nil {
        return nil, err
    }
    return JsonResponseUnmarshal[T](r)
}
