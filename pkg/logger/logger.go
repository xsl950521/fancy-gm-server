package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	// Logger 全局日志实例
	Logger *zap.Logger
	// SugaredLogger 全局sugared日志实例（更易用）
	SugaredLogger *zap.SugaredLogger
)

// Config 日志配置
type Config struct {
	Level      string `yaml:"level"`       // debug, info, warn, error
	Encoding   string `yaml:"encoding"`    // json, console
	OutputPath string `yaml:"output_path"` // 输出路径，空则输出到stdout
	ErrorPath  string `yaml:"error_path"`  // 错误日志路径，空则输出到stderr
}

// Init 初始化日志
func Init(config Config) error {
	var level zapcore.Level
	if err := level.UnmarshalText([]byte(config.Level)); err != nil {
		level = zapcore.InfoLevel
	}

	// 设置编码器配置
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.LowercaseLevelEncoder

	var encoder zapcore.Encoder
	if config.Encoding == "console" {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	}

	// 设置输出
	var core zapcore.Core
	if config.OutputPath == "" {
		core = zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), level)
	} else {
		file, err := os.OpenFile(config.OutputPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return err
		}
		core = zapcore.NewCore(encoder, zapcore.AddSync(file), level)
	}

	// 如果设置了错误日志路径，添加错误日志输出
	if config.ErrorPath != "" {
		errorFile, err := os.OpenFile(config.ErrorPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return err
		}
		errorCore := zapcore.NewCore(encoder, zapcore.AddSync(errorFile), zapcore.ErrorLevel)
		core = zapcore.NewTee(core, errorCore)
	}

	// 创建Logger
	Logger = zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
	SugaredLogger = Logger.Sugar()

	return nil
}

// DefaultConfig 返回默认配置
func DefaultConfig() Config {
	return Config{
		Level:      "info",
		Encoding:   "json",
		OutputPath: "",
		ErrorPath:  "",
	}
}

// Sync 同步日志缓冲区
func Sync() {
	if Logger != nil {
		_ = Logger.Sync()
	}
}
