package helper

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
)

// Md5 计算字符串的MD5哈希值,返回十六进制编码结果
// 注意: MD5不应用于密码存储等安全敏感场景,仅适用于非安全相关的哈希需求
func Md5(str string) string {
	h := md5.New()
	h.Write(StringToBytes(str))
	return hex.EncodeToString(h.Sum(nil))
}

// Md5Base64 计算字符串的MD5哈希值,返回Base64编码结果
// 注意: MD5不应用于密码存储等安全敏感场景,仅适用于非安全相关的哈希需求
func Md5Base64(str string) string {
	h := md5.New()
	h.Write(StringToBytes(str))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

// Base64 将字符串进行Base64编码
func Base64(str string) string {
	return base64.StdEncoding.EncodeToString(StringToBytes(str))
}

// DeBase64 将Base64编码的字符串解码
func DeBase64(str string) string {
	return BytesToString(DeBase64Byte(str))
}

// DeBase64Byte 将Base64编码的字符串解码为字节切片
// 解码失败时返回nil
func DeBase64Byte(str string) []byte {
	bytes, _ := base64.StdEncoding.DecodeString(str)
	return bytes
}
