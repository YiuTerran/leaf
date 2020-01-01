package netutil

//调整slice的长度到size，如果不足则右侧补0，超出则截断
func AdjustByteSlice(src []byte, size int) []byte {
	if len(src) == size {
		return src
	}
	if len(src) < size {
		arr := make([]byte, size)
		copy(arr, src)
		return arr
	}
	return src[:size]
}
