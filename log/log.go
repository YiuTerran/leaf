package log

import (
	"encoding/json"
	"fmt"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	tracker *zap.Logger
	logger  *zap.SugaredLogger
	cfg     zap.Config
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
func Jsonize(obj interface{}) *json.RawMessage {
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
		cfg.Level.SetLevel(zap.DebugLevel)
	} else {
		cfg.Level.SetLevel(zap.InfoLevel)
	}
}

//path是一个文件夹路径，自动生成track.json
//其他的输出到标准输出和标准错误
func InitLogger(path string) {
	if logger != nil && tracker != nil {
		return
	}
	//track logger
	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.TimeKey = "@timestamp"
	encoderCfg.EncodeTime = timeEncoder
	rawJSON := []byte(fmt.Sprintf(`{
	  "level": "info",
	  "encoding": "json",
	  "outputPaths": ["%s/track.json"],
	  "encoderConfig": {
	    "levelEncoder": "uppercase"
	  },
      "disableCaller": true
	}`, path))
	var tc zap.Config
	if err := json.Unmarshal(rawJSON, &tc); err != nil {
		panic(err)
	}
	tc.EncoderConfig = encoderCfg
	tracker, _ = tc.Build()

	rawJSON = []byte(`{
		"level": "debug",
		"outputPaths": ["stdout"],
		"errorOutputPaths": ["stderr"],
		"encoding": "console",
	}`)
	if err := json.Unmarshal(rawJSON, &cfg); err != nil {
		panic(err)
	}
	lg, _ := cfg.Build()
	logger = lg.Sugar()
}

//关闭服务器之前调用，同步缓冲区
func CloseLogger() {
	_ = logger.Sync()
	_ = tracker.Sync()
}
