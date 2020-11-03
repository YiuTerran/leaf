package byteutil

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

func byteToString(bs []byte, delim string, formatter string) string {
	var sb strings.Builder
	for i, b := range bs {
		if i != 0 && delim != "" {
			sb.WriteString(delim)
		}
		sb.WriteString(fmt.Sprintf(formatter, b))
	}
	return sb.String()
}

//将字节数组转换成10进制组成的字符串（如IP地址转换）
func BytesToIntString(bs []byte, delim string) string {
	return byteToString(bs, delim, "%d")
}

//将字节数组转成16进制组成的字符串（如MAC地址转换），强制使用
func BytesToHexString(bs []byte, delim string) string {
	return byteToString(bs, delim, "%02X")
}

func stringToByte(s string, delim string, base int) []byte {
	ss := strings.Split(s, delim)
	result := make([]byte, 0)
	for _, s := range ss {
		i, _ := strconv.ParseInt(s, base, 9) //uint8是9位
		result = append(result, byte(i))
	}
	return result
}

//将整数组成的字符串转成字节数组
func IntStringToBytes(s string, delim string) []byte {
	return stringToByte(s, delim, 10)
}

//将16进制数组成的字符串转成字节数组
func HexStringToBytes(s string, delim string) []byte {
	return stringToByte(s, delim, 16)
}

func RemoveUUIDDash(uid string) string {
	return strings.Join(strings.Split(uid, "-"), "")
}

func Md5HexString(bs []byte) string {
	m := md5.New()
	m.Write(bs)
	return hex.EncodeToString(m.Sum(nil))
}

func UUID4() string {
	u, _ := uuid.NewRandom()
	return u.String()
}

func Md5(param ...interface{}) string {
	m := md5.New()
	var ss strings.Builder
	for _, p := range param {
		ss.WriteString(fmt.Sprint(p))
	}
	m.Write([]byte(ss.String()))
	return hex.EncodeToString(m.Sum(nil))
}
