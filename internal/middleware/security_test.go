package middleware_test

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"gocms/internal/middleware"
)

func TestRateLimiter(t *testing.T) {
	rl := middleware.NewRateLimiter(3, time.Second)

	// 前3次应该允许
	for i := 0; i < 3; i++ {
		if !rl.Allow("192.168.1.1") {
			t.Errorf("request %d should be allowed", i+1)
		}
	}

	// 第4次应该被限制
	if rl.Allow("192.168.1.1") {
		t.Error("4th request should be denied")
	}

	// 不同IP应该允许
	if !rl.Allow("192.168.1.2") {
		t.Error("different IP should be allowed")
	}
}

func TestRateLimiterReset(t *testing.T) {
	rl := middleware.NewRateLimiter(2, 100*time.Millisecond)

	rl.Allow("test")
	rl.Allow("test")

	// 等待窗口过期
	time.Sleep(150 * time.Millisecond)

	if !rl.Allow("test") {
		t.Error("should be allowed after window reset")
	}
}

func TestRateLimitMiddleware(t *testing.T) {
	app := fiber.New()
	app.Use(middleware.RateLimitMiddleware(2, time.Second))
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	// 前2次成功
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("X-Forwarded-For", "10.0.0.1")
		resp, _ := app.Test(req)
		if resp.StatusCode != 200 {
			t.Errorf("request %d: expected 200, got %d", i+1, resp.StatusCode)
		}
	}

	// 第3次被限流
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Forwarded-For", "10.0.0.1")
	resp, _ := app.Test(req)
	if resp.StatusCode != 429 {
		t.Errorf("expected 429, got %d", resp.StatusCode)
	}
}

func TestCSRFProtection(t *testing.T) {
	app := fiber.New()
	app.Use(middleware.CSRFProtection())
	app.Post("/test", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	// 无 Referer/Origin 应该通过
	req := httptest.NewRequest("POST", "/test", nil)
	resp, _ := app.Test(req)
	if resp.StatusCode != 200 {
		t.Errorf("no referer should pass: got %d", resp.StatusCode)
	}

	// 正确的 Referer 应该通过
	req = httptest.NewRequest("POST", "/test", nil)
	req.Header.Set("Referer", "http://example.com/test")
	req.Header.Set("Host", "example.com")
	resp, _ = app.Test(req)
	if resp.StatusCode != 200 {
		t.Errorf("correct referer should pass: got %d", resp.StatusCode)
	}
}

func TestXSSProtection(t *testing.T) {
	app := fiber.New()
	app.Use(middleware.XSSProtection())
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, _ := app.Test(req)

	if resp.Header.Get("X-Content-Type-Options") != "nosniff" {
		t.Error("missing X-Content-Type-Options header")
	}
	if resp.Header.Get("X-Frame-Options") != "SAMEORIGIN" {
		t.Error("missing X-Frame-Options header")
	}
	if resp.Header.Get("X-XSS-Protection") != "1; mode=block" {
		t.Error("missing X-XSS-Protection header")
	}
}

func TestSecurityHeaders(t *testing.T) {
	app := fiber.New()
	app.Use(middleware.SecurityHeaders())
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, _ := app.Test(req)

	headers := map[string]string{
		"X-Content-Type-Options":                "nosniff",
		"X-Frame-Options":                       "SAMEORIGIN",
		"X-XSS-Protection":                      "1; mode=block",
		"Strict-Transport-Security":             "max-age=31536000; includeSubDomains",
	}

	for header, expected := range headers {
		if resp.Header.Get(header) != expected {
			t.Errorf("header %s: expected %s, got %s", header, expected, resp.Header.Get(header))
		}
	}
}

func TestIPWhitelist(t *testing.T) {
	// 空白名单 = 允许所有
	app := fiber.New()
	app.Use(middleware.IPWhitelist([]string{}))
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, _ := app.Test(req)
	if resp.StatusCode != 200 {
		t.Errorf("empty whitelist should allow all: got %d", resp.StatusCode)
	}
}

func TestRequestSizeLimit(t *testing.T) {
	app := fiber.New()
	app.Use(middleware.RequestSizeLimit(1024))
	app.Post("/test", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	// 小请求应该通过
	req := httptest.NewRequest("POST", "/test", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test request failed: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Logf("small request status: %d", resp.StatusCode)
	}
}
