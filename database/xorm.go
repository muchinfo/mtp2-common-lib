package database

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync/atomic"
	"time"

	_ "github.com/godror/godror"
	"go.uber.org/zap"
	"xorm.io/xorm"
	"xorm.io/xorm/log"
)

// NewOracleEngine 创建 Oracle 数据库引擎，支持 zap.Logger 注入、连接监控、慢 SQL 统计、熔断
func NewOracleEngine(dsn string, logger *zap.Logger, slowThreshold time.Duration, breaker *CircuitBreaker) (*xorm.Engine, error) {
	engine, err := xorm.NewEngine("oracle", dsn)
	if err != nil {
		return nil, err
	}
	// 连接池配置
	engine.SetMaxOpenConns(20)
	engine.SetMaxIdleConns(5)
	engine.SetConnMaxLifetime(30 * time.Minute)

	// 日志对接 zap，慢 SQL 统计
	if logger != nil {
		engine.SetLogger(&ZapXormLogger{
			logger:        logger,
			slowThreshold: slowThreshold,
		})
	} else {
		engine.SetLogger(log.NewSimpleLogger(os.Stdout))
	}
	engine.ShowSQL(true)

	// 健康检查
	if err := engine.Ping(); err != nil {
		for i := 0; i < 3; i++ {
			time.Sleep(time.Second * 2)
			if err = engine.Ping(); err == nil {
				break
			}
		}
		if err != nil {
			return nil, err
		}
	}

	// 连接监控与熔断
	if breaker != nil {
		go monitorConnection(engine, breaker, logger)
	}

	return engine, nil
}

// monitorConnection 定时健康检查，触发熔断
func monitorConnection(engine *xorm.Engine, breaker *CircuitBreaker, logger *zap.Logger) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		err := engine.Ping()
		if err != nil {
			breaker.Fail()
			if logger != nil {
				logger.Error("[XORM] 数据库健康检查失败，熔断计数+1", zap.Error(err))
			}
		} else {
			breaker.Success()
		}
	}
}

// CircuitBreaker 简单熔断器
type CircuitBreaker struct {
	failCount int32
	threshold int32
	open      int32
}

func NewCircuitBreaker(threshold int32) *CircuitBreaker {
	return &CircuitBreaker{threshold: threshold}
}

func (b *CircuitBreaker) Fail() {
	atomic.AddInt32(&b.failCount, 1)
	if atomic.LoadInt32(&b.failCount) >= b.threshold {
		atomic.StoreInt32(&b.open, 1)
	}
}

func (b *CircuitBreaker) Success() {
	atomic.StoreInt32(&b.failCount, 0)
	atomic.StoreInt32(&b.open, 0)
}

func (b *CircuitBreaker) IsOpen() bool {
	return atomic.LoadInt32(&b.open) == 1
}

func (b *CircuitBreaker) Check() error {
	if b.IsOpen() {
		return errors.New("circuit breaker open: 数据库连接异常")
	}
	return nil
}

// ZapXormLogger 实现 xorm log.Logger 接口，输出到 zap

type ZapXormLogger struct {
	logger        *zap.Logger
	level         log.LogLevel
	showSQL       bool
	slowThreshold time.Duration
}

func (z *ZapXormLogger) Debug(v ...interface{}) { z.logger.Debug(sprint(v...)) }
func (z *ZapXormLogger) Debugf(format string, v ...interface{}) {
	z.logger.Debug(fmt.Sprintf(format, v...))
}
func (z *ZapXormLogger) Error(v ...interface{}) { z.logger.Error(sprint(v...)) }
func (z *ZapXormLogger) Errorf(format string, v ...interface{}) {
	z.logger.Error(fmt.Sprintf(format, v...))
}
func (z *ZapXormLogger) Info(v ...interface{}) { z.logger.Info(sprint(v...)) }
func (z *ZapXormLogger) Infof(format string, v ...interface{}) {
	z.logger.Info(fmt.Sprintf(format, v...))
}
func (z *ZapXormLogger) Warn(v ...interface{}) { z.logger.Warn(sprint(v...)) }
func (z *ZapXormLogger) Warnf(format string, v ...interface{}) {
	z.logger.Warn(fmt.Sprintf(format, v...))
}
func (z *ZapXormLogger) Level() log.LogLevel     { return z.level }
func (z *ZapXormLogger) SetLevel(l log.LogLevel) { z.level = l }
func (z *ZapXormLogger) ShowSQL(show ...bool) {
	if len(show) > 0 {
		z.showSQL = show[0]
	}
}
func (z *ZapXormLogger) IsShowSQL() bool { return z.showSQL }

// 慢 SQL 统计
func (z *ZapXormLogger) BeforeSQL(ctx context.Context, sql string, args ...interface{}) context.Context {
	return context.WithValue(ctx, "start", time.Now())
}
func (z *ZapXormLogger) AfterSQL(ctx context.Context, sql string, args ...interface{}) {
	start, _ := ctx.Value("start").(time.Time)
	cost := time.Since(start)
	if z.slowThreshold > 0 && cost > z.slowThreshold {
		z.logger.Warn("[XORM] 慢SQL", zap.String("sql", sql), zap.Duration("cost", cost))
	}
}

func sprint(v ...interface{}) string {
	return fmt.Sprint(v...)
}

// Ping 检查数据库连接健康
func Ping(engine *xorm.Engine) error {
	return engine.Ping()
}

// TryReconnect 尝试断线重连
func TryReconnect(engine **xorm.Engine, dsn string, maxRetry int) error {
	var err error
	for i := 0; i < maxRetry; i++ {
		if *engine != nil {
			(*engine).Close()
		}
		*engine, err = xorm.NewEngine("oracle", dsn)
		if err == nil && (*engine).Ping() == nil {
			return nil
		}
		time.Sleep(time.Second * 2)
	}
	return err
}

// User 示例实体
type User struct {
	Id   int64  `xorm:"pk autoincr"`
	Name string `xorm:"varchar(100)"`
	Age  int    `xorm:"number(3)"`
}

// AutoMigrate 自动建表
func AutoMigrate(engine *xorm.Engine) error {
	return engine.Sync2(new(User))
}
