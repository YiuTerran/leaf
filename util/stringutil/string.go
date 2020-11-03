package stringutil

import (
	"fmt"
	"strings"
)

func Reverse(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

func CombineHttpUrl(host string, path ...interface{}) string {
	if len(path) == 0 {
		if strings.HasPrefix(host, "http") {
			return host
		}
		return fmt.Sprintf("http://%s", host)
	}
	suffix := ""
	for _, p := range path {
		ps := strings.Trim(fmt.Sprint(p), "/")
		if ps == "" {
			continue
		}
		//在最前面加上/
		suffix += "/" + ps
	}
	//最后一个/可能需要
	if fmt.Sprint(path[len(path)-1]) == "/" {
		suffix += "/"
	}
	host = strings.TrimSuffix(host, "/")
	if strings.HasPrefix(host, "http") {
		return fmt.Sprintf("%s%s", host, suffix)
	} else {
		return fmt.Sprintf("http://%s%s", host, suffix)
	}
}
