package main

import (
	"github.com/muchinfo/mtp2-common-lib/logger"

	"go.uber.org/zap"
)

func RunLoggerExample() {
	// 方式1: 使用自定义配置初始化
	config := &logger.Config{
		Level:      "debug",
		Format:     "console",
		OutputPath: "./logs/app.log",
		MaxAge:     7,  // 保留7天
		Rotation:   24, // 24小时切割一次
	}

	// 推荐外部项目使用 InitWithCallerSkip(config, 2)，保证日志调用栈定位到业务代码
	if err := logger.InitWithCallerSkip(config, 2); err != nil {
		panic("初始化日志失败: " + err.Error())
	}
	defer logger.Close()

	// 方式2: 使用开发环境快速初始化 (注释掉上面的代码使用这个)
	// if err := logger.InitDevelopment(); err != nil {
	// 	panic("初始化日志失败: " + err.Error())
	// }
	// defer logger.Close()

	// 方式3: 使用生产环境快速初始化 (注释掉上面的代码使用这个)
	// if err := logger.InitProduction(); err != nil {
	// 	panic("初始化日志失败: " + err.Error())
	// }
	// defer logger.Close()

	// 使用结构化日志记录
	logger.Info("应用启动",
		zap.String("version", "1.0.0"),
		zap.Int("port", 8080),
	)

	logger.Debug("调试信息",
		zap.String("user_id", "12345"),
		zap.String("action", "login"),
	)

	logger.Warn("警告信息",
		zap.String("message", "磁盘空间不足"),
		zap.Int("available_gb", 5),
	)

	logger.Error("错误信息",
		zap.String("error", "数据库连接失败"),
		zap.String("database", "mysql"),
	)

	// 使用糖化日志 (类似 printf 格式)
	logger.Infof("用户 %s 登录成功, IP: %s", "张三", "192.168.1.100")
	logger.Debugf("处理请求耗时: %dms", 150)
	logger.Warnf("缓存命中率较低: %.2f%%", 65.5)
	logger.Errorf("处理订单 %d 时发生错误: %s", 12345, "库存不足")

	// 直接使用全局 Logger 和 SugarLogger
	logger.Logger.Info("使用全局Logger",
		zap.String("module", "auth"),
		zap.Bool("success", true),
	)

	logger.SugarLogger.Infow("使用全局SugarLogger",
		"module", "payment",
		"amount", 99.99,
		"currency", "CNY",
	)

	logger.Info("应用退出")
}
