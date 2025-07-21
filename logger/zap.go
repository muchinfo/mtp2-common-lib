package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	rotatelogs "github.com/lestrrat/go-file-rotatelogs"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	// Logger 全局日志实例
	Logger *zap.Logger
	// SugarLogger 全局糖化日志实例
	SugarLogger *zap.SugaredLogger
)

// Config 日志配置结构
type Config struct {
	Level      string `json:"level" yaml:"level"`             // 日志级别: debug, info, warn, error
	Format     string `json:"format" yaml:"format"`           // 日志格式: json, console
	OutputPath string `json:"output_path" yaml:"output_path"` // 日志输出路径
	MaxAge     int    `json:"max_age" yaml:"max_age"`         // 日志保留天数; 0 表示永久保留
	Rotation   int    `json:"rotation" yaml:"rotation"`       // 日志切割时间(小时)
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		Level:      "info",
		Format:     "json",
		OutputPath: "./logs/app.log",
		MaxAge:     7, // 保留7天
		Rotation:   1, // 1小时切割一次
	}
}

// Init 初始化日志
func Init(config *Config) error {
	if config == nil {
		config = DefaultConfig()
	}

	// 创建日志目录
	logDir := filepath.Dir(config.OutputPath)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("创建日志目录失败: %w", err)
	}

	// 配置日志级别
	level := zapcore.InfoLevel
	switch config.Level {
	case "debug":
		level = zapcore.DebugLevel
	case "info":
		level = zapcore.InfoLevel
	case "warn":
		level = zapcore.WarnLevel
	case "error":
		level = zapcore.ErrorLevel
	default:
		level = zapcore.InfoLevel
	}

	// 配置编码器
	var encoder zapcore.Encoder
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05.000")
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	if config.Format == "console" {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	}

	// 配置日志轮转，按自然小时切割
	writer, err := rotatelogs.New(
		config.OutputPath+".%Y%m%d%H",
		rotatelogs.WithLinkName(config.OutputPath),
		rotatelogs.WithMaxAge(time.Duration(config.MaxAge)*24*time.Hour),
		rotatelogs.WithRotationTime(time.Hour), // 强制每个自然小时切割
		rotatelogs.WithClock(rotatelogs.Local), // 使用本地时钟，确保自然小时
	)
	if err != nil {
		return fmt.Errorf("配置日志轮转失败: %w", err)
	}

	// 创建 WriteSyncer
	fileWriteSyncer := zapcore.AddSync(writer)
	consoleWriteSyncer := zapcore.AddSync(os.Stdout)

	// 创建 Core
	core := zapcore.NewTee(
		zapcore.NewCore(encoder, fileWriteSyncer, level),
		zapcore.NewCore(zapcore.NewConsoleEncoder(encoderConfig), consoleWriteSyncer, level),
	)

	// 创建 Logger
	Logger = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
	SugarLogger = Logger.Sugar()

	return nil
}

// InitDevelopment 初始化开发环境日志 (简化配置)
func InitDevelopment() error {
	config := zap.NewDevelopmentConfig()
	config.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05")

	logger, err := config.Build(zap.AddCallerSkip(1))
	if err != nil {
		return fmt.Errorf("初始化开发环境日志失败: %w", err)
	}

	Logger = logger
	SugarLogger = logger.Sugar()
	return nil
}

// InitProduction 初始化生产环境日志 (简化配置)
func InitProduction() error {
	config := zap.NewProductionConfig()
	config.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05")

	logger, err := config.Build(zap.AddCallerSkip(1))
	if err != nil {
		return fmt.Errorf("初始化生产环境日志失败: %w", err)
	}

	Logger = logger
	SugarLogger = logger.Sugar()
	return nil
}

// Sync 同步日志
func Sync() {
	if Logger != nil {
		_ = Logger.Sync()
	}
	if SugarLogger != nil {
		_ = SugarLogger.Sync()
	}
}

// Close 关闭日志
func Close() {
	Sync()
}

// 便捷方法 - 使用结构化日志
func Debug(msg string, fields ...zap.Field) {
	if Logger != nil {
		Logger.Debug(msg, fields...)
	}
}

func Info(msg string, fields ...zap.Field) {
	if Logger != nil {
		Logger.Info(msg, fields...)
	}
}

func Warn(msg string, fields ...zap.Field) {
	if Logger != nil {
		Logger.Warn(msg, fields...)
	}
}

func Error(msg string, fields ...zap.Field) {
	if Logger != nil {
		Logger.Error(msg, fields...)
	}
}

func Fatal(msg string, fields ...zap.Field) {
	if Logger != nil {
		Logger.Fatal(msg, fields...)
	}
}

// 便捷方法 - 使用糖化日志 (类似 printf 格式)
func Debugf(template string, args ...interface{}) {
	if SugarLogger != nil {
		SugarLogger.Debugf(template, args...)
	}
}

func Infof(template string, args ...interface{}) {
	if SugarLogger != nil {
		SugarLogger.Infof(template, args...)
	}
}

func Warnf(template string, args ...interface{}) {
	if SugarLogger != nil {
		SugarLogger.Warnf(template, args...)
	}
}

func Errorf(template string, args ...interface{}) {
	if SugarLogger != nil {
		SugarLogger.Errorf(template, args...)
	}
}

func Fatalf(template string, args ...interface{}) {
	if SugarLogger != nil {
		SugarLogger.Fatalf(template, args...)
	}
}
