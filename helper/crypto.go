package helper

import (
    "crypto/md5"
    "encoding/base64"
    "encoding/hex"
)

func Md5(str string) string {
    h := md5.New()
    h.Write(StringToBytes(str))
    return hex.EncodeToString(h.Sum(nil))
}

func Md5Base64(str string) string {
    h := md5.New()
    h.Write(StringToBytes(str))
    return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

func Base64(str string) string {
    return base64.StdEncoding.EncodeToString(StringToBytes(str))
}

func DeBase64(str string) string {
    return BytesToString(DeBase64Byte(str))
}

func DeBase64Byte(str string) []byte {
    bytes, _ := base64.StdEncoding.DecodeString(str)
    return bytes
}
