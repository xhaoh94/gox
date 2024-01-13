package logger

import (
	"io"
	"os"
	"path/filepath"
	"strconv"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v2"
)

type (
	logConf struct {
		LogLevel   string `yaml:"log_level"`
		LogMaxSize int    `yaml:"log_max_size"`
		MaxBackups int    `yaml:"log_max_backups"`
		LogMaxAge  int    `yaml:"log_max_age"`
		Console    bool   `yaml:"log_console"`
		Skip       int    `yaml:"log_callerskip"`
	}
)

var (
	development bool
	Logger      zerolog.Logger
	// LogSampled zerolog.Logger
)

func init() {
	timeFormat := "2006-01-02 15:04:05"
	zerolog.TimeFieldFormat = timeFormat
	// 路径脱敏, 日志格式规范, 避免与自定义字段名冲突: {"E":"is Err(error)","error":"is Str(error)"}
	zerolog.TimestampFieldName = "T"
	zerolog.LevelFieldName = "L"
	zerolog.MessageFieldName = "M"
	zerolog.ErrorFieldName = "E"
	zerolog.CallerFieldName = "C"
	zerolog.ErrorStackFieldName = "S"
	zerolog.DurationFieldInteger = true
}

// 配置热加载等场景调用, 重载日志环境
func Init(path string, dev bool) error {
	development = dev
	if err := LogConfig(path); err != nil {
		return err
	}

	// // 抽样的日志记录器
	// sampler := &zerolog.BurstSampler{
	// 	Burst:  3,
	// 	Period: time.Second,
	// }
	// LogSampled = Log.Sample(&zerolog.LevelSampler{
	// 	TraceSampler: sampler,
	// 	DebugSampler: sampler,
	// 	InfoSampler:  sampler,
	// 	WarnSampler:  sampler,
	// 	ErrorSampler: sampler,
	// })

	return nil
}

// LogConfig 加载日志配置
func LogConfig(path string) error {
	var logCfg logConf
	if path == "" {
		logCfg = logConf{
			LogLevel:   "debug",
			LogMaxSize: 100,
			MaxBackups: 30,
			LogMaxAge:  7,
			Console:    true,
			Skip:       2,
		}
	} else {
		bytes, err := os.ReadFile(path)
		if err != nil {
			log.Fatal().Err(err).Str("path", path).Msg("LogConfig")
			return err
		}

		err = yaml.Unmarshal(bytes, &logCfg)
		if err != nil {
			log.Fatal().Err(err).Str("path", path).Msg("LogConfig")
			return err
		}
	}
	if logCfg.LogLevel == "" {
		log.Fatal().Str("path", path).Msg("日志等级为空")
	}
	zerolog.CallerMarshalFunc = func(pc uintptr, file string, line int) string {
		return filepath.Base(file) + ":" + strconv.Itoa(line)
	}
	var (
		writers []io.Writer = make([]io.Writer, 0)
	)
	lv := parseLevel(logCfg.LogLevel)

	if !development {
		// 2. 生产环境时, 日志输出到文件
		for i := zerolog.DebugLevel; i < zerolog.PanicLevel; i++ {
			if i >= lv {
				writers = append(writers, newLevelFileWriter(i, "tt"))
			}
		}
	}
	if logCfg.Console {
		// 1. 开发环境时, 日志高亮输出到控制台
		writers = []io.Writer{newConsoleWriter(lv, !development, "2006-01-02 15:04:05")}
	} else {
		writers = append(writers, newLevelWriter(lv))
	}
	Logger = zerolog.New(zerolog.MultiLevelWriter(writers...)).Level(lv).With().Timestamp().
		Caller().Logger()

	return nil
}
