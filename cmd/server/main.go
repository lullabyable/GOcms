package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"

	"gocms/internal/config"
	"gocms/internal/database"
	"gocms/internal/middleware"
	"gocms/internal/router"
	"gocms/internal/session"
)

func main() {
	// 加载配置
	cfg, err := config.Load("")
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 初始化日志
	logger := initLogger(cfg.Log)
	defer logger.Sync()

	// 连接数据库
	db, err := database.Connect(cfg.DB)
	if err != nil {
		logger.Fatal("数据库连接失败", zap.Error(err))
	}
	sqlDB, _ := db.DB()
	defer sqlDB.Close()
	logger.Info("数据库连接成功")

	// 数据库迁移
	migrator := database.NewMigrator(db)
	if err := migrator.Migrate(); err != nil {
		logger.Warn("数据库迁移提示", zap.Error(err))
	}

	// 初始化 Session
	sm := session.NewManager(session.ManagerConfig{
		Type:      cfg.Session.Type,
		CookieName: cfg.Session.Name,
		MaxAge:    time.Duration(cfg.Session.MaxAge) * time.Second,
		Secret:    cfg.Session.Secret,
	})

	// 创建 Fiber 应用
	app := fiber.New(fiber.Config{
		AppName:      "GoCMS Go",
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
	})

	// 注册中间件
	middleware.Setup(app, logger)

	// 注册路由
	router.Setup(app, sm, db)

	// 静态文件
	app.Static("/static", "./web/static")
	app.Static("/uploads", "./web/uploads")

	// 优雅退出
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		logger.Info("收到退出信号，正在关闭...")
		app.Shutdown()
	}()

	// 启动服务
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	logger.Info("服务启动", zap.String("addr", addr))
	if err := app.Listen(addr); err != nil {
		logger.Fatal("服务启动失败", zap.Error(err))
	}
}

func initLogger(cfg config.LogConfig) *zap.Logger {
	rotator := &lumberjack.Logger{
		Filename:   cfg.File,
		MaxSize:    cfg.MaxSize,
		MaxBackups: cfg.MaxBackups,
		MaxAge:     cfg.MaxAge,
	}

	var level zapcore.Level
	switch cfg.Level {
	case "debug":
		level = zapcore.DebugLevel
	case "warn":
		level = zapcore.WarnLevel
	case "error":
		level = zapcore.ErrorLevel
	default:
		level = zapcore.InfoLevel
	}

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		zapcore.AddSync(rotator),
		level,
	)
	return zap.New(core, zap.AddCaller())
}
