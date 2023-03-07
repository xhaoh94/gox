package xlog

// import (
// 	"fmt"
// 	"strings"

// 	"github.com/xhaoh94/gox/engine/app"
// 	"go.uber.org/zap"
// )

// var zlog *ZapLog

// func Init(conf app.LogConf) {
// 	if zlog != nil {
// 		return
// 	}
// 	zlog = new(conf)
// }
// func Destroy() {
// 	if zlog != nil {
// 		zlog.logger.Sync()
// 	}
// }

// func formatFileds(format string, args ...interface{}) (string, []zap.Field) {
// 	l := len(args)
// 	if l > 0 {
// 		return fmt.Sprintf(format, args[:l]...), []zap.Field{}
// 	}
// 	return format, []zap.Field{}
// }

// func Debug(format string, args ...interface{}) {
// 	toLog("debug", format, args...)
// }

// func Info(format string, args ...interface{}) {
// 	toLog("info", format, args...)
// }

// func Warn(format string, args ...interface{}) {
// 	toLog("warn", format, args...)
// }

// func Error(format string, args ...interface{}) {
// 	toLog("error", format, args...)
// }

// func Panic(format string, args ...interface{}) {
// 	toLog("panic", format, args...)
// }

// func Fatal(format string, args ...interface{}) {
// 	toLog("fatal", format, args...)
// }

// func toLog(level string, format string, args ...interface{}) {
// 	if zlog != nil {
// 		s, f := formatFileds(format, args...)
// 		switch strings.ToLower(level) {
// 		case "panic":
// 			zlog.logger.Panic(s, f...)
// 			break
// 		case "fatal":
// 			zlog.logger.Fatal(s, f...)
// 			break
// 		case "error":
// 			zlog.logger.Error(s, f...)
// 			break
// 		case "warn":
// 			zlog.logger.Warn(s, f...)
// 			break
// 		case "info":
// 			zlog.logger.Info(s, f...)
// 			break
// 		case "debug":
// 			zlog.logger.Debug(s, f...)
// 			break
// 		default:
// 			zlog.logger.Debug(s, f...)
// 			break
// 		}
// 	} else {
// 		fmt.Printf(format, args...)
// 	}
// }
