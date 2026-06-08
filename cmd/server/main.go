package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"gorm.io/gorm"

	"gocms/internal/config"
	"gocms/internal/database"
	"gocms/internal/handler/admin"
	"gocms/internal/middleware"
	"gocms/internal/router"
	"gocms/internal/session"
	"gocms/internal/template"
)

func main() {
	// 加载配置（如果不存在则使用空配置，等待 /install 安装）
	cfg, err := config.Load("")
	if err != nil {
		// 配置文件不存在，使用默认配置
		cfg = &config.Config{}
		cfg.Server.Host = "0.0.0.0"
		cfg.Server.Port = 8080
		cfg.Server.ReadTimeout = 30
		cfg.Server.WriteTimeout = 30
		cfg.Template.Dir = "./web/templates"
		cfg.Template.Theme = "default"
		cfg.Cache.Type = "file"
		cfg.Cache.FileDir = "./runtime/cache"
		cfg.Log.Level = "info"
		cfg.Log.File = "./runtime/logs/gocms.log"
	}

	// 初始化日志
	logger := initLogger(cfg.Log)
	defer logger.Sync()

	// 连接数据库（失败不退出，等待 /install 安装）
	var db *gorm.DB
	installH := admin.NewInstallHandler()

	db, err = database.Connect(cfg.DB)
	if err != nil {
		logger.Warn("数据库未就绪，等待安装", zap.Error(err))
	} else {
		sqlDB, _ := db.DB()
		defer sqlDB.Close()
		logger.Info("数据库连接成功")

		// 数据库迁移
		migrator := database.NewMigrator(db)
		if err := migrator.Migrate(); err != nil {
			logger.Warn("数据库迁移提示", zap.Error(err))
		}

		// 检查是否已安装
		if db.Migrator().HasTable("mac_admin") {
			installH.SetInstalled(true)
		}
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

	// 初始化模板引擎
	tplEngine := template.NewEngine(cfg.Template.Dir, cfg.Template.Theme)
	if err := tplEngine.Load(); err != nil {
		logger.Warn("模板加载提示", zap.Error(err))
	} else {
		logger.Info("模板加载成功",
			zap.String("theme", tplEngine.Theme()),
			zap.String("dir", tplEngine.ThemeDir()))
	}

	// 注册中间件
	middleware.Setup(app, logger)

	// 注册路由
	router.Setup(app, sm, db, tplEngine, installH)

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
