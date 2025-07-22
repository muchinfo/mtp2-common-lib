package logger

import (
	"testing"

	"go.uber.org/zap"
)

func TestInitWithCallerSkip(t *testing.T) {
	config := DefaultConfig()
	err := InitWithCallerSkip(config, 0)
	if err != nil {
		t.Fatalf("InitWithCallerSkip failed: %v", err)
	}
	defer Close()
	if Logger == nil {
		t.Error("Logger 未初始化")
	}
	if SugarLogger == nil {
		t.Error("SugarLogger 未初始化")
	}
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()
	if config.Level != "info" {
		t.Errorf("Expected level 'info', got '%s'", config.Level)
	}
	if config.Format != "json" {
		t.Errorf("Expected format 'json', got '%s'", config.Format)
	}
}

func TestInit(t *testing.T) {
	// 测试初始化功能，但不使用轮转以避免 Windows 文件锁定问题
	err := InitDevelopment()
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	defer Close()

	// 测试日志记录
	Info("测试信息", zap.String("test", "value"))
	Infof("测试格式化: %s", "value")

	// 验证全局变量是否初始化
	if Logger == nil {
		t.Error("Logger 未初始化")
	}
	if SugarLogger == nil {
		t.Error("SugarLogger 未初始化")
	}
}

func TestInitDevelopment(t *testing.T) {
	err := InitDevelopment()
	if err != nil {
		t.Fatalf("InitDevelopment failed: %v", err)
	}
	defer Close()

	if Logger == nil {
		t.Error("Logger 未初始化")
	}
	if SugarLogger == nil {
		t.Error("SugarLogger 未初始化")
	}
}

func TestInitProduction(t *testing.T) {
	err := InitProduction()
	if err != nil {
		t.Fatalf("InitProduction failed: %v", err)
	}
	defer Close()

	if Logger == nil {
		t.Error("Logger 未初始化")
	}
	if SugarLogger == nil {
		t.Error("SugarLogger 未初始化")
	}
}

func TestLoggerMethods(t *testing.T) {
	err := InitDevelopment()
	if err != nil {
		t.Fatalf("InitDevelopment failed: %v", err)
	}
	defer Close()

	// 测试结构化日志方法
	Debug("debug message", zap.String("key", "value"))
	Info("info message", zap.String("key", "value"))
	Warn("warn message", zap.String("key", "value"))
	Error("error message", zap.String("key", "value"))

	// 测试糖化日志方法
	Debugf("debug: %s", "formatted")
	Infof("info: %s", "formatted")
	Warnf("warn: %s", "formatted")
	Errorf("error: %s", "formatted")
}

func TestSafeLogging(t *testing.T) {
	// 测试在未初始化时调用日志方法不会 panic
	Logger = nil
	SugarLogger = nil

	Debug("test")
	Info("test")
	Warn("test")
	Error("test")

	Debugf("test %s", "format")
	Infof("test %s", "format")
	Warnf("test %s", "format")
	Errorf("test %s", "format")

	// 应该没有 panic
}

func BenchmarkStructuredLogging(b *testing.B) {
	err := InitProduction()
	if err != nil {
		b.Fatalf("InitProduction failed: %v", err)
	}
	defer Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Info("benchmark test",
			zap.Int("iteration", i),
			zap.String("type", "structured"),
		)
	}
}

func BenchmarkSugarLogging(b *testing.B) {
	err := InitProduction()
	if err != nil {
		b.Fatalf("InitProduction failed: %v", err)
	}
	defer Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Infof("benchmark test iteration: %d, type: %s", i, "sugar")
	}
}
