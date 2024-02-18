package helper

import (
    "crypto/md5"
    "encoding/base64"
    "encoding/hex"
)

func Md5(str string) string {
    h := md5.New()
    h.Write([]byte(str))
    return hex.EncodeToString(h.Sum(nil))
}

func Md5Base64(str string) string {
    h := md5.New()
    h.Write([]byte(str))
    return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

func Base64(str string) string {
    return base64.StdEncoding.EncodeToString([]byte(str))
}

func DeBase64(str string) string {
    bytes, _ := base64.StdEncoding.DecodeString(str)
    return string(bytes)
}
