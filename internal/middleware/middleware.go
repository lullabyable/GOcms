package middleware

import (
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"go.uber.org/zap"
	"gocms/internal/session"
)

// Setup 注册全局中间件
func Setup(app *fiber.App, logger *zap.Logger) {
	app.Use(recover.New())
	app.Use(requestid.New())
	app.Use(cors.New())
	app.Use(compress.New())
	app.Use(ZapLogger(logger))
}

// ZapLogger 请求日志中间件
func ZapLogger(logger *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next()
		latency := time.Since(start)

		logger.Info("request",
			zap.String("method", c.Method()),
			zap.String("path", c.Path()),
			zap.Int("status", c.Response().StatusCode()),
			zap.Duration("latency", latency),
			zap.String("ip", c.IP()),
		)
		return err
	}
}

// AdminAuth 后台认证中间件
func AdminAuth(sm *session.Manager) fiber.Handler {
	return func(c *fiber.Ctx) error {
		path := c.Path()
		// 登录页面跳过
		if strings.HasSuffix(path, "/login") || strings.HasSuffix(path, "/captcha") || strings.HasSuffix(path, "/verify") {
			return c.Next()
		}

		sess := sm.Get(c)
		adminID := sess.Get("admin_id")
		if adminID == "" {
			if c.Get("Accept") == "application/json" || c.Get("X-Requested-With") == "XMLHttpRequest" {
				return c.Status(401).JSON(fiber.Map{"code": 401, "msg": "请先登录"})
			}
			return c.Redirect("/admin/login")
		}

		c.Locals("admin_id", adminID)
		c.Locals("admin_name", sess.Get("admin_name"))
		c.Locals("admin_role", sess.Get("admin_role"))
		return c.Next()
	}
}
