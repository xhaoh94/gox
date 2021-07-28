package xlog

import (
	"fmt"

	"go.uber.org/zap"
)

var log *ZapLog

func Init() {
	if log != nil {
		return
	}
	log = new()
}
func Destroy() {
	log.logger.Sync()
}

func formatFileds(format string, args ...interface{}) (string, []zap.Field) {
	l := len(args)
	if l > 0 {
		return fmt.Sprintf(format, args[:l]...), []zap.Field{}
	}
	return format, []zap.Field{}
}

func Debug(format string, args ...interface{}) {
	s, f := formatFileds(format, args...)
	log.logger.Debug(s, f...)
}

func Info(format string, args ...interface{}) {
	s, f := formatFileds(format, args...)
	log.logger.Info(s, f...)
}

func Warn(format string, args ...interface{}) {
	s, f := formatFileds(format, args...)
	log.logger.Warn(s, f...)
}

func Error(format string, args ...interface{}) {
	s, f := formatFileds(format, args...)
	log.logger.Error(s, f...)
}

func Panic(format string, args ...interface{}) {
	s, f := formatFileds(format, args...)
	log.logger.Panic(s, f...)
}

func Fatal(format string, args ...interface{}) {
	s, f := formatFileds(format, args...)
	log.logger.Fatal(s, f...)
}
