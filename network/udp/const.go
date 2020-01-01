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
