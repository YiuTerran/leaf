package udp

import (
	"golang.org/x/xerrors"
)

const (
	MaxPacketSize     = 65507
	DefaultPacketSize = 1024
)

var (
	ChanFullError = xerrors.New("write chan full")
	InitError     = xerrors.New("fail to init")
)


func MergeBytes(bs [][]byte) []byte {
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
