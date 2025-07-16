package main

import (
	"fmt"
	"mtp2-common-lib/database"
	"os"
	"time"

	"go.uber.org/zap"
)

func RunXormOracleExample() {
	dsn := os.Getenv("ORACLE_DSN")
	if dsn == "" {
		fmt.Println("请设置 ORACLE_DSN 环境变量，如 user/password@host:port/sid")
		return
	}
	logger, _ := zap.NewDevelopment()
	breaker := database.NewCircuitBreaker(3)
	engine, err := database.NewOracleEngine(dsn, logger, 10*time.Millisecond, breaker)
	if err != nil {
		panic(err)
	}
	defer engine.Close()

	err = database.AutoMigrate(engine)
	if err != nil {
		panic(err)
	}

	user := &database.User{Name: "李四", Age: 30}
	_, err = engine.Insert(user)
	if err != nil {
		panic(err)
	}
	fmt.Printf("插入用户: %+v\n", user)

	var got database.User
	has, err := engine.ID(user.Id).Get(&got)
	if err != nil || !has {
		panic("查询失败")
	}
	fmt.Printf("查询到用户: %+v\n", got)

	user.Age = 31
	_, err = engine.ID(user.Id).Update(user)
	if err != nil {
		panic(err)
	}
	fmt.Println("更新用户年龄为 31")

	_, err = engine.ID(user.Id).Delete(new(database.User))
	if err != nil {
		panic(err)
	}
	fmt.Println("删除用户成功")

	// 演示慢SQL统计
	if setter, ok := engine.Logger().(interface{ SetSlowThreshold(time.Duration) }); ok {
		setter.SetSlowThreshold(1 * time.Nanosecond)
	}
	engine.Query("select * from user where 1=0")

	// 演示熔断
	breaker.Fail()
	breaker.Fail()
	breaker.Fail()
	if breaker.IsOpen() {
		fmt.Println("熔断器已打开，数据库连接异常")
	}
}
