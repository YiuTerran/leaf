package log

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	rotate "github.com/lestrrat-go/file-rotatelogs"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	LenStackBuf = 4096
)

var (
	tracker    *zap.Logger
	logger     *zap.SugaredLogger
	infoWriter io.Writer
	warnWriter io.Writer
	debug      = zap.NewAtomicLevelAt(zap.DebugLevel)
	once       sync.Once
)

func Debug(format string, a ...interface{}) {
	logger.Debugf(format, a...)
}

func Info(format string, a ...interface{}) {
	logger.Infof(format, a...)
}

func Warn(format string, a ...interface{}) {
	logger.Warnf(format, a...)
}

func Error(format string, a ...interface{}) {
	logger.Errorf(format, a...)
}

func Fatal(format string, a ...interface{}) {
	logger.Fatalf(format, a...)
}

//输出json的track
func Track(msg string, fields ...zap.Field) {
	tracker.Info(msg, fields...)
}

//辅助函数，Track中使用zap.Any打印Object之前将其转换为原始json
func ToRawJson(obj interface{}) *json.RawMessage {
	bytes, err := json.Marshal(obj)
	if err != nil {
		return nil
	}
	p := json.RawMessage(bytes)
	return &p
}

func timeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02T15:04:05.000000"))
}

//激活debug等级
func EnableDebug(option bool) {
	if option {
		debug.SetLevel(zap.DebugLevel)
	} else {
		debug.SetLevel(zap.InfoLevel)
	}
}

// 判断所给路径文件/文件夹是否存在=>避免循环依赖fs
func exists(path string) bool {
	_, err := os.Stat(path) //os.Stat获取文件信息
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}

//path是一个文件夹路径，自动生成track.json, service.log, err.log
//其他的输出到标准输出和标准错误
func InitLogger(path string) {
	once.Do(func() {
		if !exists(path) && os.Mkdir(path, os.ModePerm) != nil {
			panic("fail to create log directory")
		}
		encoderCfg := zap.NewProductionEncoderConfig()
		encoderCfg.TimeKey = "@timestamp"
		encoderCfg.EncodeTime = timeEncoder
		tracker = zap.New(
			zapcore.NewCore(zapcore.NewJSONEncoder(encoderCfg),
				zapcore.AddSync(getWriter(filepath.Join(path, "track"))),
				zap.InfoLevel))
		//高优先级
		hp := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
			return lvl >= zapcore.WarnLevel
		})
		//所有
		all := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
			if debug.Enabled(zap.DebugLevel) {
				return true
			}
			return lvl > zap.DebugLevel
		})
		//都输出到标准输出，方便调试
		warnWriter = getWriter(filepath.Join(path, "error"))
		infoWriter = getWriter(filepath.Join(path, "service"))
		encoder := zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())
		core := zapcore.NewTee(
			// 将info及以下写入logPath,  warn及以上写入errPath
			zapcore.NewCore(encoder, zapcore.AddSync(infoWriter), all),
			zapcore.NewCore(encoder, zapcore.AddSync(warnWriter), hp),
			//同步到stdout，方便调试
			zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), all),
		)
		lg := zap.New(core)
		logger = lg.Sugar()
	})
}

func getWriter(filename string) io.Writer {
	// 生成rotatelogs的Logger 实际生成的文件名 demo.mmdd.log
	hook, err := rotate.New(
		filename+".%m%d.log",
		rotate.WithLinkName(filename),
		rotate.WithMaxAge(time.Hour*24*10),    // 保存10天
		rotate.WithRotationTime(time.Hour*24), //切割频率 24小时
	)
	if err != nil {
		panic(err)
	}
	return hook
}

func GetOriginLogger() *zap.SugaredLogger {
	return logger
}

//关闭服务器之前调用，同步缓冲区
func CloseLogger() {
	_ = logger.Sync()
	_ = tracker.Sync()
}
