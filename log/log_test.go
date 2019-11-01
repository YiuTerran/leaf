package log

import (
	"testing"

	"go.uber.org/zap"
)

func TestLogger(t *testing.T) {
	InitLogger("/tmp/")
	EnableDebug(true)
	Debug("this is a debug")
	Info("hello %v", "world")
	Error("fail to %v", "open the door")
	Track("track", zap.String("hello", "world"))
	EnableDebug(false)
	Debug("should not print")
	CloseLogger()
}