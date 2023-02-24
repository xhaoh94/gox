package logger

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/rs/zerolog"
	"gopkg.in/natefinch/lumberjack.v2"
)

func parseLevel(lv string) zerolog.Level {
	switch strings.ToLower(lv) {
	case "panic":
		return zerolog.PanicLevel
	case "fatal":
		return zerolog.FatalLevel
	case "error":
		return zerolog.ErrorLevel
	case "warn":
		return zerolog.WarnLevel
	case "info":
		return zerolog.InfoLevel
	case "debug":
		return zerolog.DebugLevel
	default:
		return zerolog.NoLevel
	}
}

func toLevelStr(lv zerolog.Level) string {
	switch lv {
	case zerolog.PanicLevel:
		return "panic"
	case zerolog.FatalLevel:
		return "fatal"
	case zerolog.ErrorLevel:
		return "error"
	case zerolog.WarnLevel:
		return "warn"
	case zerolog.InfoLevel:
		return "info"
	case zerolog.DebugLevel:
		return "debug"
	}
	return ""
}

// 创建分割日志的writer
func getWriter(Filename string, MaxSize int, MaxAge int, MaxBackups int) *lumberjack.Logger {
	err := os.MkdirAll(Filename, os.ModePerm)
	if err != nil {
		panic(err)
	}
	return &lumberjack.Logger{
		Filename:   Filename,
		MaxSize:    MaxSize,
		MaxAge:     MaxAge,
		MaxBackups: MaxBackups,
		LocalTime:  true,
		Compress:   true,
	}
}

// levelFileWriter 指定级别的日志写文件
type levelFileWriter struct {
	lw io.Writer
	lv zerolog.Level
}

func (w *levelFileWriter) Write(p []byte) (n int, err error) {
	return w.lw.Write(p)
}

func (w *levelFileWriter) WriteLevel(l zerolog.Level, p []byte) (n int, err error) {
	if l == w.lv {
		return w.lw.Write(p)
	}
	return len(p), nil
}
func newLevelFileWriter(lv zerolog.Level, fileName string) io.Writer {
	fileName = fileName + "_" + toLevelStr(lv) + ".log"
	lw := getWriter(fileName, 10, 30, 10)
	return &levelFileWriter{lw: lw, lv: lv}
}

// levelWriter 指定级别的日志写文件
type levelWriter struct {
	lv zerolog.Level
}

func (w *levelWriter) Write(p []byte) (n int, err error) {
	return os.Stderr.Write(p)
}

func (w *levelWriter) WriteLevel(l zerolog.Level, p []byte) (n int, err error) {
	if l >= w.lv {
		return w.Write(p)
	}
	return len(p), nil
}
func newLevelWriter(lv zerolog.Level) io.Writer {
	return &levelWriter{lv: lv}
}

// consoleWriter 控制台日志
type consoleWriter struct {
	lw zerolog.ConsoleWriter
	lv zerolog.Level
}

func (w *consoleWriter) Write(p []byte) (n int, err error) {
	return w.lw.Write(p)
}

func (w *consoleWriter) WriteLevel(l zerolog.Level, p []byte) (n int, err error) {
	if l >= w.lv {
		return w.lw.Write(p)
	}
	return len(p), nil
}
func newConsoleWriter(lv zerolog.Level, nocolor bool, timeFormat string) io.Writer {
	lw := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: timeFormat}
	lw.NoColor = nocolor
	lw.FormatLevel = func(i interface{}) string {
		return strings.ToUpper(fmt.Sprintf("| %-6s|", i))
	}
	lw.FormatMessage = func(i interface{}) string {
		return fmt.Sprintf("%s", i)
	}
	lw.FormatFieldName = func(i interface{}) string {
		return fmt.Sprintf("%s:", i)
	}
	lw.FormatFieldValue = func(i interface{}) string {
		return fmt.Sprintf("%s;", i)
	}
	return &consoleWriter{
		lw: lw,
		lv: lv,
	}
}

// func InitLogger() {
// 	errorFile, _ := os.OpenFile("error.log", os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
// 	warnFile, _ := os.OpenFile("warn.log", os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
// 	Log = zerolog.New(zerolog.MultiLevelWriter(
// 		// Warn 日志写入 warn.log
// 		&levelFileWriter{lw: warnFile, lv: zerolog.WarnLevel},
// 		// Error 日志写入 error.log
// 		&levelFileWriter{lw: errorFile, lv: zerolog.ErrorLevel},
// 		// Debug, Error 日志显示在控制台
// 		&consoleWriter{
// 			lw: zerolog.ConsoleWriter{Out: os.Stdout},
// 			lv: zerolog.DebugLevel,
// 		},
// 	)).With().Timestamp().Caller().Logger()
// }

// func main() {
// 	InitLogger()
// 	Log.Trace().Msg("test TRACE")
// 	Log.Debug().Msg("test DEBUG")
// 	Log.Info().Msg("test INFO")
// 	Log.Warn().Msg("test WARN")
// 	Log.Error().Msg("test ERROR")
// 	Log.Fatal().Msg("test FATAL")
// }
