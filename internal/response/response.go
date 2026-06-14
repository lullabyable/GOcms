package response

import "github.com/gofiber/fiber/v2"

// Response 统一 JSON 响应结构
type Response struct {
	Code  int         `json:"code"`  // 0=失败, 1=成功
	Msg   string      `json:"msg"`   // 提示信息
	Data  interface{} `json:"data"`  // 数据
	Count int64       `json:"count"` // 列表总数(分页用)
}

// PageData 分页数据结构
type PageData struct {
	List  interface{} `json:"list"`
	Total int64       `json:"total"`
	Page  int         `json:"page"`
	Limit int         `json:"limit"`
}

// OK 返回成功响应（带数据）
func OK(c *fiber.Ctx, data interface{}) error {
	return c.JSON(Response{Code: 1, Msg: "success", Data: data})
}

// OKMsg 返回成功响应（仅消息）
func OKMsg(c *fiber.Ctx, msg string) error {
	return c.JSON(Response{Code: 1, Msg: msg})
}

// OKData 返回成功响应（带消息和数据）
func OKData(c *fiber.Ctx, msg string, data interface{}) error {
	return c.JSON(Response{Code: 1, Msg: msg, Data: data})
}

// OKList 返回列表成功响应（带总数）
func OKList(c *fiber.Ctx, data interface{}, count int64) error {
	return c.JSON(Response{Code: 1, Msg: "success", Data: data, Count: count})
}

// Fail 返回失败响应
func Fail(c *fiber.Ctx, msg string) error {
	return c.JSON(Response{Code: 0, Msg: msg})
}

// FailWithCode 返回失败响应（自定义状态码）
func FailWithCode(c *fiber.Ctx, code int, msg string) error {
	return c.JSON(Response{Code: code, Msg: msg})
}

// Page 返回分页响应
func Page(c *fiber.Ctx, data interface{}, count int64, page, limit int) error {
	return c.JSON(Response{
		Code:  1,
		Msg:   "success",
		Data:  PageData{List: data, Total: count, Page: page, Limit: limit},
		Count: count,
	})
}

// Error 返回错误响应（带 HTTP 状态码）
func Error(c *fiber.Ctx, httpCode int, msg string) error {
	return c.Status(httpCode).JSON(Response{Code: 0, Msg: msg})
}
