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
    t := new(T)
    if r.Body == nil {
        return nil, errors.New("body is nil")
    }
    defer r.Body.Close()
    all, err := io.ReadAll(r.Body)
    if err != nil {
        return nil, err
    }
    err = JsonUnmarshal(all, t)
    return nil, err
}

func JsonRequestUnmarshal[T any](req *http.Request) (*T, error) {
    r, err := http.DefaultClient.Do(req)
    if err != nil {
        return nil, err
    }
    return JsonResponseUnmarshal[T](r)
}
