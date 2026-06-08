package admin

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
	"gocms/internal/model"
)

type AdminHandler struct{ db *gorm.DB }

func NewAdminHandler(db *gorm.DB) *AdminHandler { return &AdminHandler{db: db} }

func (h *AdminHandler) List(c *fiber.Ctx) error {
	var admins []model.Admin
	h.db.Find(&admins)
	// 不返回密码
	type safeAdmin struct {
		AdminID       int    `json:"admin_id"`
		AdminName     string `json:"admin_name"`
		AdminRole     int    `json:"admin_role"`
		AdminStatus   int    `json:"admin_status"`
		AdminLastTime int64  `json:"admin_last_time"`
		AdminLoginNum int    `json:"admin_login_num"`
	}
	var result []safeAdmin
	for _, a := range admins {
		result = append(result, safeAdmin{
			AdminID: a.AdminID, AdminName: a.AdminName, AdminRole: a.AdminRole,
			AdminStatus: a.AdminStatus, AdminLastTime: a.AdminLastTime, AdminLoginNum: a.AdminLoginNum,
		})
	}
	return c.JSON(fiber.Map{"code": 1, "data": result})
}

func (h *AdminHandler) Save(c *fiber.Ctx) error {
	var admin model.Admin
	if err := c.BodyParser(&admin); err != nil {
		return c.Status(400).JSON(fiber.Map{"code": 0, "msg": "参数错误"})
	}
	if admin.AdminID > 0 {
		h.db.Model(&admin).Omit("admin_pwd").Updates(admin)
	} else {
		// TODO: 密码加密
		h.db.Create(&admin)
	}
	return c.JSON(fiber.Map{"code": 1, "msg": "保存成功"})
}

func (h *AdminHandler) Delete(c *fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))
	if id == 1 {
		return c.JSON(fiber.Map{"code": 0, "msg": "不能删除超管"})
	}
	h.db.Delete(&model.Admin{}, id)
	return c.JSON(fiber.Map{"code": 1, "msg": "删除成功"})
}
