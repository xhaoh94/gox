package xlog

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/xhaoh94/gox/app"

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
func newHook(id uint, str string) *lumberjack.Logger {
	logCfg := app.GetAppCfg().Log
	path := logCfg.LogPath
	if err := os.MkdirAll(path, 0766); err != nil {
		panic(err)
	}
	logName := fmt.Sprintf("server_%d_%s.log", id, str)

	return &lumberjack.Logger{
		Filename:   filepath.Join(path, logName),
		MaxSize:    logCfg.LogMaxSize,
		MaxAge:     logCfg.LogMaxAge,
		MaxBackups: logCfg.MaxBackups,
		// LocalTime:  true,
		Compress: false,
	}
}
func new(id uint) *ZapLog {
	logger := newZapLogger(id)
	zap.RedirectStdLog(logger)
	// logger = logger.With(zap.Uint("sid", id)) //额外需要加的数据
	return &ZapLog{logger: logger}
}
func newZapLogger(id uint) *zap.Logger {
	logCfg := app.GetAppCfg().Log
	encCfg := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "lev",
		NameKey:        "app",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeCaller:   zapcore.ShortCallerEncoder,
		EncodeDuration: zapcore.MillisDurationEncoder,
		EncodeLevel:    zapcore.CapitalColorLevelEncoder,
		EncodeTime: func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendString(t.Format("2006-01-02 15:04:05.000"))
		},
		// EncodeTime:zapcore.ISO8601TimeEncoder,
	}

	var encoder zapcore.Encoder
	switch logCfg.Console {
	case "console":
		encoder = zapcore.NewConsoleEncoder(encCfg)
		break
	case "json":
		encoder = zapcore.NewJSONEncoder(encCfg)
		break
	default:
		encoder = zapcore.NewConsoleEncoder(encCfg)
		break
	}
	var coreArr []zapcore.Core

	zLev := parseLevel(logCfg.LogLevel)
	split := logCfg.Split && zLev < zap.ErrorLevel

	levWriters := []zapcore.WriteSyncer{zapcore.AddSync(os.Stdout)}
	if logCfg.IsWriteLog {
		levWriters = append(levWriters, zapcore.AddSync(newHook(id, logCfg.LogLevel)))
	}
	levWriteSyncer := zapcore.NewMultiWriteSyncer(levWriters...)
	levPriority := zap.LevelEnablerFunc(func(lev zapcore.Level) bool {
		if split {
			return lev < zap.ErrorLevel && lev >= zLev
		}
		return lev >= zLev
	})
	levCore := zapcore.NewCore(encoder, levWriteSyncer, levPriority)
	coreArr = append(coreArr, levCore)

	if split {
		errWriters := []zapcore.WriteSyncer{zapcore.AddSync(os.Stdout)}
		if logCfg.IsWriteLog {
			errWriters = append(errWriters, zapcore.AddSync(newHook(id, "error")))
		}
		errWriteSyncer := zapcore.NewMultiWriteSyncer(errWriters...)
		errPriority := zap.LevelEnablerFunc(func(lev zapcore.Level) bool {
			return lev >= zap.ErrorLevel
		})
		errCore := zapcore.NewCore(encoder, errWriteSyncer, errPriority)
		coreArr = append(coreArr, errCore)
	}

	ops := []zap.Option{zap.AddCaller(), zap.AddStacktrace(parseLevel(logCfg.Stacktrace)), zap.AddCallerSkip(logCfg.Skip)}
	if logCfg.Development {
		ops = append(ops, zap.Development())
	}

	return zap.New(zapcore.NewTee(coreArr...), ops...)
}
