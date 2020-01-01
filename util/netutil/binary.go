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

func MergeBytes(bs [][]byte) []byte {
	if len(bs) == 0 {
		return []byte{}
	}
	if len(bs) == 1 {
		return bs[0]
	}
	l := 0
	for i := 0; i < len(bs); i++ {
		l += len(bs[i])
	}
	buffer := make([]byte, l)
	l = 0
	for i := 0; i < len(bs); i++ {
		copy(buffer[l:], bs[i])
		l += len(bs[i])
	}
	return buffer
}
