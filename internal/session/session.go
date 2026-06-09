package session

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/gofiber/fiber/v2"
)

// Store 会话存储接口
type Store interface {
	Get(key string) (string, error)
	Set(key, value string, ttl time.Duration) error
	Delete(key string) error
}

// Manager 会话管理器
type Manager struct {
	store      Store
	cookieName string
	maxAge     time.Duration
	secret     []byte
	secure     bool
}

type ManagerConfig struct {
	Secure     bool   // HTTPS 时设为 true，HTTP 环境必须 false
	Type       string // cookie | redis
	CookieName string
	MaxAge     time.Duration
	Secret     string
	// redis 可选
	RedisAddr     string
	RedisPassword string
	RedisDB       int
}

func NewManager(cfg ManagerConfig) *Manager {
	var store Store
	switch cfg.Type {
	case "redis":
		store = NewRedisStore(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB)
	default:
		store = NewMemStore()
	}

	name := cfg.CookieName
	if name == "" {
		name = "gocms_session"
	}
	secret := make([]byte, 32)
	copy(secret, []byte(cfg.Secret))

	return &Manager{store: store, cookieName: name, maxAge: cfg.MaxAge, secret: secret, secure: cfg.Secure}
}

// Get 获取当前请求的会话
func (m *Manager) Get(c *fiber.Ctx) *Session {
	sid := c.Cookies(m.cookieName)
	if sid == "" {
		sid = m.generateID()
		m.setCookie(c, sid)
	}
	return &Session{id: sid, store: m.store, ttl: m.maxAge}
}

// Regenerate 登录后重新生成 Session ID（防固定攻击）
func (m *Manager) Regenerate(c *fiber.Ctx) *Session {
	old := c.Cookies(m.cookieName)
	sid := m.generateID()
	if old != "" {
		m.store.Delete(old)
	}
	m.setCookie(c, sid)
	return &Session{id: sid, store: m.store, ttl: m.maxAge}
}

// Destroy 销毁会话
func (m *Manager) Destroy(c *fiber.Ctx) {
	sid := c.Cookies(m.cookieName)
	if sid != "" {
		m.store.Delete(sid)
	}
	c.Cookie(&fiber.Cookie{Name: m.cookieName, Value: "", MaxAge: -1, HTTPOnly: true, Secure: m.secure, SameSite: "Lax"})
}

func (m *Manager) setCookie(c *fiber.Ctx, sid string) {
	c.Cookie(&fiber.Cookie{
		Name:     m.cookieName,
		Value:    sid,
		MaxAge:   int(m.maxAge.Seconds()),
		HTTPOnly: true,
		Secure:   m.secure,
		SameSite: "Lax",
	})
}

func (m *Manager) generateID() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// Encrypt 加密敏感数据存 Cookie
func (m *Manager) Encrypt(plaintext string) (string, error) {
	block, err := aes.NewCipher(m.secret)
	if err != nil {
		return "", err
	}
	gcm, _ := cipher.NewGCM(block)
	nonce := make([]byte, gcm.NonceSize())
	rand.Read(nonce)
	return hex.EncodeToString(gcm.Seal(nonce, nonce, []byte(plaintext), nil)), nil
}

// Decrypt 解密
func (m *Manager) Decrypt(ciphertext string) (string, error) {
	data, _ := hex.DecodeString(ciphertext)
	block, err := aes.NewCipher(m.secret)
	if err != nil {
		return "", err
	}
	gcm, _ := cipher.NewGCM(block)
	ns := gcm.NonceSize()
	plaintext, err := gcm.Open(nil, data[:ns], data[ns:], nil)
	return string(plaintext), err
}

// Session 单个会话
type Session struct {
	id    string
	store Store
	ttl   time.Duration
}

func (s *Session) ID() string { return s.id }

func (s *Session) Get(key string) string {
	v, _ := s.store.Get(s.id + ":" + key)
	return v
}

func (s *Session) Set(key, value string) {
	s.store.Set(s.id+":"+key, value, s.ttl)
}

func (s *Session) Delete(key string) {
	s.store.Delete(s.id + ":" + key)
}
