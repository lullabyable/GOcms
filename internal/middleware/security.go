package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
)

// RateLimiter 限流器
type RateLimiter struct {
	visitors map[string]*visitor
	mu       sync.RWMutex
	rate     int           // 每窗口请求数
	window   time.Duration // 时间窗口
}

type visitor struct {
	count    int
	lastSeen time.Time
}

// NewRateLimiter 创建限流器
func NewRateLimiter(rate int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*visitor),
		rate:     rate,
		window:   window,
	}
	// 定期清理
	go rl.cleanup()
	return rl
}

func (rl *RateLimiter) cleanup() {
	for {
		time.Sleep(rl.window * 2)
		rl.mu.Lock()
		for ip, v := range rl.visitors {
			if time.Since(v.lastSeen) > rl.window*2 {
				delete(rl.visitors, ip)
			}
		}
		rl.mu.Unlock()
	}
}

// Allow 检查是否允许请求
func (rl *RateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	v, exists := rl.visitors[key]
	if !exists || time.Since(v.lastSeen) > rl.window {
		rl.visitors[key] = &visitor{count: 1, lastSeen: time.Now()}
		return true
	}

	if v.count >= rl.rate {
		return false
	}

	v.count++
	v.lastSeen = time.Now()
	return true
}

// RateLimitMiddleware HTTP 限流中间件
func RateLimitMiddleware(rate int, window time.Duration) fiber.Handler {
	limiter := NewRateLimiter(rate, window)
	return func(c *fiber.Ctx) error {
		ip := c.IP()
		if !limiter.Allow(ip) {
			return c.Status(http.StatusTooManyRequests).JSON(fiber.Map{
				"code": 429,
				"msg":  "请求过于频繁，请稍后再试",
			})
		}
		return c.Next()
	}
}

// CSRFProtection CSRF 防护中间件
func CSRFProtection() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// 只检查 POST/PUT/DELETE 请求
		method := c.Method()
		if method == "POST" || method == "PUT" || method == "DELETE" {
			// 检查 Referer
			referer := c.Get("Referer")
			origin := c.Get("Origin")
			host := c.Hostname()

			if referer != "" && !strings.Contains(referer, host) {
				return c.Status(http.StatusForbidden).JSON(fiber.Map{
					"code": 403,
					"msg":  "CSRF 验证失败",
				})
			}
			if origin != "" && !strings.Contains(origin, host) {
				return c.Status(http.StatusForbidden).JSON(fiber.Map{
					"code": 403,
					"msg":  "CSRF 验证失败",
				})
			}
		}
		return c.Next()
	}
}

// XSSProtection XSS 防护中间件
func XSSProtection() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// 设置安全头
		c.Set("X-Content-Type-Options", "nosniff")
		c.Set("X-Frame-Options", "SAMEORIGIN")
		c.Set("X-XSS-Protection", "1; mode=block")
		c.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Set("Content-Security-Policy", "default-src 'self' 'unsafe-inline' 'unsafe-eval' https: http: data: blob:")
		return c.Next()
	}
}

// SQLInjectionCheck SQL 注入检测中间件
func SQLInjectionCheck() fiber.Handler {
	sqlPatterns := []string{
		"'", "\"", ";", "--", "/*", "*/", "xp_", "exec ", "execute ",
		"insert ", "delete ", "update ", "drop ", "truncate ", "union ",
		"select ", "from ", "where ", "or 1=1", "or '1'='1",
	}

	return func(c *fiber.Ctx) error {
		// 检查查询参数
		query := c.Request().URI().QueryString()
		if containsSQLInjection(string(query), sqlPatterns) {
			return c.Status(http.StatusForbidden).JSON(fiber.Map{
				"code": 403,
				"msg":  "检测到异常请求",
			})
		}

		// 检查表单参数
		if c.Method() == "POST" {
			body := string(c.Body())
			if containsSQLInjection(body, sqlPatterns) {
				return c.Status(http.StatusForbidden).JSON(fiber.Map{
					"code": 403,
					"msg":  "检测到异常请求",
				})
			}
		}

		return c.Next()
	}
}

func containsSQLInjection(input string, patterns []string) bool {
	lower := strings.ToLower(input)
	for _, p := range patterns {
		if strings.Contains(lower, strings.ToLower(p)) {
			return true
		}
	}
	return false
}

// IPWhitelist IP 白名单中间件
func IPWhitelist(whitelist []string) fiber.Handler {
	allowed := make(map[string]bool)
	for _, ip := range whitelist {
		allowed[ip] = true
	}

	return func(c *fiber.Ctx) error {
		if len(allowed) > 0 && !allowed[c.IP()] {
			return c.Status(http.StatusForbidden).JSON(fiber.Map{
				"code": 403,
				"msg":  "IP 不在白名单中",
			})
		}
		return c.Next()
	}
}

// RequestSizeLimit 请求体大小限制
func RequestSizeLimit(maxSize int) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if c.Request().Header.ContentLength() > maxSize {
			return c.Status(http.StatusRequestEntityTooLarge).JSON(fiber.Map{
				"code": 413,
				"msg":  fmt.Sprintf("请求体超过限制 (%d bytes)", maxSize),
			})
		}
		return c.Next()
	}
}

// SecurityHeaders 安全响应头
func SecurityHeaders() fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Set("X-Content-Type-Options", "nosniff")
		c.Set("X-Frame-Options", "SAMEORIGIN")
		c.Set("X-XSS-Protection", "1; mode=block")
		c.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		c.Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		return c.Next()
	}
}
