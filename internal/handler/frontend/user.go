package frontend

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"gocms/internal/model"
	"gocms/internal/session"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserHandler struct {
	db *gorm.DB
	sm *session.Manager
}

func NewUserHandler(db *gorm.DB, sm *session.Manager) *UserHandler {
	return &UserHandler{db: db, sm: sm}
}

func (h *UserHandler) Register(c *fiber.Ctx) error {
	name := c.FormValue("user_name")
	pwd := c.FormValue("user_pwd")
	email := c.FormValue("user_email")

	if name == "" || pwd == "" {
		return c.JSON(fiber.Map{"code": 0, "msg": "username and password are required"})
	}

	var count int64
	h.db.Model(&model.User{}).Where("user_name = ?", name).Count(&count)
	if count > 0 {
		return c.JSON(fiber.Map{"code": 0, "msg": "username already exists"})
	}

	hashedPwd, err := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"code": 0, "msg": "registration failed"})
	}

	user := model.User{
		UserName:    name,
		UserPwd:     string(hashedPwd),
		UserEmail:   email,
		UserStatus:  1,
		UserRegTime: time.Now().Unix(),
		UserRegIP:   c.IP(),
	}
	h.db.Create(&user)

	return c.JSON(fiber.Map{"code": 1, "msg": "registration successful"})
}

func (h *UserHandler) Login(c *fiber.Ctx) error {
	name := c.FormValue("user_name")
	pwd := c.FormValue("user_pwd")

	var user model.User
	if err := h.db.Where("user_name = ?", name).First(&user).Error; err != nil {
		return c.JSON(fiber.Map{"code": 0, "msg": "invalid username or password"})
	}
	if !validPassword(user.UserPwd, pwd) {
		return c.JSON(fiber.Map{"code": 0, "msg": "invalid username or password"})
	}
	if user.UserStatus != 1 {
		return c.JSON(fiber.Map{"code": 0, "msg": "account disabled"})
	}

	if needsPasswordUpgrade(user.UserPwd) {
		if hashedPwd, err := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost); err == nil {
			h.db.Model(&user).Update("user_pwd", string(hashedPwd))
		}
	}

	sess := h.sm.Regenerate(c)
	sess.Set("user_id", strconv.Itoa(user.UserID))
	sess.Set("user_name", user.UserName)

	h.db.Model(&user).Updates(map[string]interface{}{
		"user_login_time": time.Now().Unix(),
		"user_login_ip":   c.IP(),
		"user_login_num":  gorm.Expr("user_login_num + 1"),
	})

	return c.JSON(fiber.Map{"code": 1, "msg": "login successful"})
}

func validPassword(stored, supplied string) bool {
	if err := bcrypt.CompareHashAndPassword([]byte(stored), []byte(supplied)); err == nil {
		return true
	}
	return stored == supplied
}

func needsPasswordUpgrade(stored string) bool {
	_, err := bcrypt.Cost([]byte(stored))
	return err != nil
}

func (h *UserHandler) Info(c *fiber.Ctx) error {
	sess := h.sm.Get(c)
	uid := sess.Get("user_id")
	if uid == "" {
		return c.Status(401).JSON(fiber.Map{"code": 0, "msg": "not logged in"})
	}

	var user model.User
	h.db.First(&user, uid)
	return c.JSON(fiber.Map{"code": 1, "data": user})
}

func (h *UserHandler) Logout(c *fiber.Ctx) error {
	h.sm.Destroy(c)
	return c.JSON(fiber.Map{"code": 1, "msg": "logout successful"})
}
