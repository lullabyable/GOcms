package admin

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"math/rand"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

// VerifyHandler 验证码
type VerifyHandler struct{}

func NewVerifyHandler() *VerifyHandler {
	return &VerifyHandler{}
}

// Image 生成验证码图片
func (h *VerifyHandler) Image(c *fiber.Ctx) error {
	code := strconv.Itoa(rand.Intn(9000) + 1000) // 4位数字

	// 存入 session（简单实现，用 cookie 存明文仅做演示）
	c.Cookie(&fiber.Cookie{
		Name:     "verify_code",
		Value:    code,
		MaxAge:   300, // 5分钟有效
		HTTPOnly: true,
	})

	// 生成简单图片
	width, height := 120, 40
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// 背景
	bg := color.RGBA{R: 245, G: 245, B: 245, A: 255}
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, bg)
		}
	}

	// 干扰线
	for i := 0; i < 5; i++ {
		clr := color.RGBA{
			R: uint8(rand.Intn(200)),
			G: uint8(rand.Intn(200)),
			B: uint8(rand.Intn(200)),
			A: 255,
		}
		x1, y1 := rand.Intn(width), rand.Intn(height)
		x2, y2 := rand.Intn(width), rand.Intn(height)
		drawLine(img, x1, y1, x2, y2, clr)
	}

	// 画数字（简易点阵字体 5x7）
	digitPatterns := getDigitPatterns()
	startX := 15
	for i, ch := range code {
		pattern := digitPatterns[int(ch-'0')]
		clr := color.RGBA{
			R: uint8(rand.Intn(100)),
			G: uint8(rand.Intn(100)),
			B: uint8(rand.Intn(100)),
			A: 255,
		}
		drawDigit(img, startX+i*25, 8, pattern, clr)
	}

	var buf bytes.Buffer
	png.Encode(&buf, img)

	c.Set("Content-Type", "image/png")
	c.Set("Cache-Control", "no-cache, no-store")
	return c.Send(buf.Bytes())
}

// Check 校验验证码（供登录时调用）
func (h *VerifyHandler) Check(c *fiber.Ctx, inputCode string) bool {
	cookie := c.Cookies("verify_code")
	if cookie == "" {
		return false
	}
	// 清除验证码
	c.Cookie(&fiber.Cookie{
		Name:   "verify_code",
		Value:  "",
		MaxAge: -1,
	})
	return cookie == inputCode
}

func drawLine(img *image.RGBA, x1, y1, x2, y2 int, clr color.RGBA) {
	dx := abs(x2 - x1)
	dy := abs(y2 - y1)
	sx, sy := 1, 1
	if x1 > x2 {
		sx = -1
	}
	if y1 > y2 {
		sy = -1
	}
	err := dx - dy
	for {
		if x1 >= 0 && x1 < img.Bounds().Dx() && y1 >= 0 && y1 < img.Bounds().Dy() {
			img.Set(x1, y1, clr)
		}
		if x1 == x2 && y1 == y2 {
			break
		}
		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x1 += sx
		}
		if e2 < dx {
			err += dx
			y1 += sy
		}
	}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func drawDigit(img *image.RGBA, x, y int, pattern []string, clr color.RGBA) {
	scale := 3
	for row, line := range pattern {
		for col, ch := range line {
			if ch == '1' {
				for dy := 0; dy < scale; dy++ {
					for dx := 0; dx < scale; dx++ {
						px := x + col*scale + dx
						py := y + row*scale + dy
						if px >= 0 && px < img.Bounds().Dx() && py >= 0 && py < img.Bounds().Dy() {
							img.Set(px, py, clr)
						}
					}
				}
			}
		}
	}
}

func getDigitPatterns() [][]string {
	return [][]string{
		// 0
		{"01110", "10001", "10011", "10101", "11001", "10001", "01110"},
		// 1
		{"00100", "01100", "00100", "00100", "00100", "00100", "01110"},
		// 2
		{"01110", "10001", "00001", "00110", "01000", "10000", "11111"},
		// 3
		{"01110", "10001", "00001", "00110", "00001", "10001", "01110"},
		// 4
		{"00010", "00110", "01010", "10010", "11111", "00010", "00010"},
		// 5
		{"11111", "10000", "11110", "00001", "00001", "10001", "01110"},
		// 6
		{"00110", "01000", "10000", "11110", "10001", "10001", "01110"},
		// 7
		{"11111", "00001", "00010", "00100", "01000", "01000", "01000"},
		// 8
		{"01110", "10001", "10001", "01110", "10001", "10001", "01110"},
		// 9
		{"01110", "10001", "10001", "01111", "00001", "00010", "01100"},
	}
}
