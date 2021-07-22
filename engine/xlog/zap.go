package xlog

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/xhaoh94/gox/app"
	"github.com/xhaoh94/gox/engine/conf"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type ZapLog struct {
	logger *zap.Logger
}

func parseLevel(lvl string) zapcore.Level {
	switch strings.ToLower(lvl) {
	case "panic":
		return zapcore.PanicLevel
	case "fatal":
		return zapcore.FatalLevel
	case "error":
		return zapcore.ErrorLevel
	case "warn":
		return zapcore.WarnLevel
	case "info":
		return zapcore.InfoLevel
	case "debug":
		return zapcore.DebugLevel
	default:
		return zapcore.DebugLevel
	}
}

//创建分割日志的writer
func newHook() *lumberjack.Logger {

	path := conf.AppCfg.Log.LogPath
	if err := os.MkdirAll(path, 0766); err != nil {
		panic(err)
	}
	now := time.Now()
	module := fmt.Sprintf("%s_%d_%02d_%02d_%02d_%02d_%02d.log", app.ServiceType, now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), now.Second())

	return &lumberjack.Logger{
		Filename:   filepath.Join(path, module),
		MaxSize:    conf.AppCfg.Log.LogMaxSize,
		MaxAge:     conf.AppCfg.Log.LogMaxAge,
		MaxBackups: conf.AppCfg.Log.MaxBackups,
		LocalTime:  true,
		Compress:   false,
	}
}
func new() *ZapLog {
	logger := newZapLogger()
	zap.RedirectStdLog(logger)
	// logger = logger.With(zap.String("service", app.ServiceID)) //额外需要加的数据
	return &ZapLog{logger: logger}
}
func newZapLogger() *zap.Logger {
	encCfg := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "app",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeCaller:   zapcore.ShortCallerEncoder,
		EncodeDuration: zapcore.NanosDurationEncoder,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime: func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendString(t.Format("2006-01-02 15:04:05.000"))
		},
		// EncodeTime:zapcore.ISO8601TimeEncoder,
	}

	atomicLevel := zap.NewAtomicLevel()
	atomicLevel.SetLevel(parseLevel(conf.AppCfg.Log.LogLevel))

	writers := []zapcore.WriteSyncer{zapcore.AddSync(os.Stdout)}
	if conf.AppCfg.Log.IsWriteLog {
		hook := newHook()
		writers = append(writers, zapcore.AddSync(hook))
	}

	core := zapcore.NewCore(zapcore.NewJSONEncoder(encCfg), zapcore.NewMultiWriteSyncer(writers...), atomicLevel)
	ops := []zap.Option{zap.AddCaller(), zap.AddStacktrace(parseLevel(conf.AppCfg.Log.Stacktrace)), zap.AddCallerSkip(1)}
	if conf.AppCfg.Log.Development {
		ops = append(ops, zap.Development())
	}
	// // 设置初始化字段
	// filed := zap.Fields(zap.String("serviceName", "serviceName"))
	return zap.New(core, ops...)
}
