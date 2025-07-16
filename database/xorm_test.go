package database

import (
	"os"
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestOracleEngineAndCRUD(t *testing.T) {
	dsn := os.Getenv("ORACLE_DSN")
	if dsn == "" {
		t.Skip("未设置 ORACLE_DSN，跳过测试")
	}
	logger, _ := zap.NewDevelopment()
	breaker := NewCircuitBreaker(2)
	engine, err := NewOracleEngine(dsn, logger, 10*time.Millisecond, breaker)
	if err != nil {
		t.Fatalf("连接 Oracle 失败: %v", err)
	}
	defer engine.Close()

	if err := AutoMigrate(engine); err != nil {
		t.Fatalf("自动建表失败: %v", err)
	}

	user := &User{Name: "张三", Age: 20}
	_, err = engine.Insert(user)
	if err != nil {
		t.Fatalf("插入失败: %v", err)
	}

	var got User
	has, err := engine.ID(user.Id).Get(&got)
	if err != nil || !has {
		t.Fatalf("查询失败: %v", err)
	}
	if got.Name != user.Name || got.Age != user.Age {
		t.Errorf("查询结果不符: %+v", got)
	}

	user.Age = 21
	_, err = engine.ID(user.Id).Update(user)
	if err != nil {
		t.Fatalf("更新失败: %v", err)
	}

	_, err = engine.ID(user.Id).Delete(new(User))
	if err != nil {
		t.Fatalf("删除失败: %v", err)
	}

	// 测试慢SQL统计
	// 如果 Logger 支持设置慢SQL阈值，则调用其方法
	if setter, ok := engine.Logger().(interface{ SetSlowThreshold(time.Duration) }); ok {
		setter.SetSlowThreshold(1 * time.Nanosecond)
	}
	engine.Query("select * from user where 1=0")

	// 测试熔断
	breaker.Fail()
	breaker.Fail()
	if !breaker.IsOpen() {
		t.Error("熔断器未打开")
	}
	if err := breaker.Check(); err == nil {
		t.Error("熔断器未报错")
	}
}
