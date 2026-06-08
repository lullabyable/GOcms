package admin

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"gocms/internal/service/payment"
)

// OrderHandler 订单管理处理器
type OrderHandler struct {
	payment *payment.Service
}

func NewOrderHandler(svc *payment.Service) *OrderHandler {
	return &OrderHandler{payment: svc}
}

// List 订单列表
func (h *OrderHandler) List(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	pageSize, _ := strconv.Atoi(c.Query("page_size", "20"))
	status, _ := strconv.Atoi(c.Query("status", "-1"))

	orders, total, err := h.payment.GetAllOrders(page, pageSize, status)
	if err != nil {
		return c.JSON(fiber.Map{"code": 0, "msg": "查询失败"})
	}

	return c.JSON(fiber.Map{
		"code": 1,
		"data": fiber.Map{
			"list":      orders,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}

// Pay 模拟支付完成
func (h *OrderHandler) Pay(c *fiber.Ctx) error {
	orderNo := c.FormValue("order_no")
	if err := h.payment.PayOrder(orderNo); err != nil {
		return c.JSON(fiber.Map{"code": 0, "msg": err.Error()})
	}
	return c.JSON(fiber.Map{"code": 1, "msg": "支付成功"})
}

// Cancel 取消订单
func (h *OrderHandler) Cancel(c *fiber.Ctx) error {
	orderNo := c.FormValue("order_no")
	if err := h.payment.CancelOrder(orderNo); err != nil {
		return c.JSON(fiber.Map{"code": 0, "msg": err.Error()})
	}
	return c.JSON(fiber.Map{"code": 1, "msg": "已取消"})
}

// CardList 卡密列表
func (h *OrderHandler) CardList(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	pageSize, _ := strconv.Atoi(c.Query("page_size", "20"))
	status, _ := strconv.Atoi(c.Query("status", "-1"))
	batchNo := c.Query("batch_no", "")

	cards, total, err := h.payment.GetCardKeys(page, pageSize, status, batchNo)
	if err != nil {
		return c.JSON(fiber.Map{"code": 0, "msg": "查询失败"})
	}

	return c.JSON(fiber.Map{
		"code": 1,
		"data": fiber.Map{
			"list":      cards,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}

// GenerateCards 生成卡密
func (h *OrderHandler) GenerateCards(c *fiber.Ctx) error {
	batchNo := c.FormValue("batch_no")
	count, _ := strconv.Atoi(c.FormValue("count", "10"))
	groupID, _ := strconv.Atoi(c.FormValue("group_id", "0"))
	points, _ := strconv.Atoi(c.FormValue("points", "0"))
	days, _ := strconv.Atoi(c.FormValue("days", "30"))

	if count <= 0 || count > 10000 {
		return c.JSON(fiber.Map{"code": 0, "msg": "数量必须在 1-10000 之间"})
	}

	cards, err := h.payment.GenerateCardKeys(batchNo, count, groupID, points, days)
	if err != nil {
		return c.JSON(fiber.Map{"code": 0, "msg": "生成失败"})
	}

	return c.JSON(fiber.Map{"code": 1, "msg": "生成成功", "data": fiber.Map{"count": len(cards)}})
}

// PaymentConfig 获取/保存支付配置
func (h *OrderHandler) PaymentConfig(c *fiber.Ctx) error {
	if c.Method() == "GET" {
		configs, err := h.payment.GetPaymentConfigs()
		if err != nil {
			return c.JSON(fiber.Map{"code": 0, "msg": "查询失败"})
		}
		return c.JSON(fiber.Map{"code": 1, "data": configs})
	}
	return c.JSON(fiber.Map{"code": 1, "msg": "配置已保存"})
}
