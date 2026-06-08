package frontend

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
	"gocms/internal/model"
	"gocms/internal/session"
)

type UserHandler struct {
	db *gorm.DB
	sm *session.Manager
}

func NewUserHandler(db *gorm.DB, sm *session.Manager) *UserHandler {
	return &UserHandler{db: db, sm: sm}
}

// Register 注册
func (h *UserHandler) Register(c *fiber.Ctx) error {
	name := c.FormValue("user_name")
	pwd := c.FormValue("user_pwd")
	email := c.FormValue("user_email")

	if name == "" || pwd == "" {
		return c.JSON(fiber.Map{"code": 0, "msg": "用户名和密码不能为空"})
	}

	// 检查重复
	var count int64
	h.db.Model(&model.User{}).Where("user_name = ?", name).Count(&count)
	if count > 0 {
		return c.JSON(fiber.Map{"code": 0, "msg": "用户名已存在"})
	}

	user := model.User{
		UserName:    name,
		UserPwd:     pwd, // TODO: bcrypt 加密
		UserEmail:   email,
		UserStatus:  1,
		UserRegTime: time.Now().Unix(),
		UserRegIP:   c.IP(),
	}
	h.db.Create(&user)

	return c.JSON(fiber.Map{"code": 1, "msg": "注册成功"})
}

// Login 登录
func (h *UserHandler) Login(c *fiber.Ctx) error {
	name := c.FormValue("user_name")
	pwd := c.FormValue("user_pwd")

	var user model.User
	if err := h.db.Where("user_name = ? AND user_pwd = ?", name, pwd).First(&user).Error; err != nil {
		return c.JSON(fiber.Map{"code": 0, "msg": "用户名或密码错误"})
	}
	if user.UserStatus != 1 {
		return c.JSON(fiber.Map{"code": 0, "msg": "账号已被禁用"})
	}

	sess := h.sm.Get(c)
	sess.Set("user_id", strconv.Itoa(user.UserID))
	sess.Set("user_name", user.UserName)

	h.db.Model(&user).Updates(map[string]interface{}{
		"user_login_time": time.Now().Unix(),
		"user_login_ip":   c.IP(),
		"user_login_num":  gorm.Expr("user_login_num + 1"),
	})

	return c.JSON(fiber.Map{"code": 1, "msg": "登录成功"})
}

// Info 用户信息
func (h *UserHandler) Info(c *fiber.Ctx) error {
	sess := h.sm.Get(c)
	uid := sess.Get("user_id")
	if uid == "" {
		return c.Status(401).JSON(fiber.Map{"code": 0, "msg": "未登录"})
	}

	var user model.User
	h.db.First(&user, uid)
	return c.JSON(fiber.Map{"code": 1, "data": user})
}

// Logout 登出
func (h *UserHandler) Logout(c *fiber.Ctx) error {
	h.sm.Destroy(c)
	return c.JSON(fiber.Map{"code": 1, "msg": "已退出"})
}
