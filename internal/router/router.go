package router

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"gocms/internal/config"
	"gocms/internal/handler/admin"
	"gocms/internal/handler/api"
	"gocms/internal/handler/frontend"
	"gocms/internal/middleware"
	"gocms/internal/model"
	"gocms/internal/service/aicontent"
	"gocms/internal/service/analytics"
	"gocms/internal/service/chat"
	"gocms/internal/service/collect"
	"gocms/internal/service/payment"
	"gocms/internal/service/plugin"
	"gocms/internal/service/scheduler"
	"gocms/internal/service/urlpush"
	"gocms/internal/session"
	"gocms/internal/template"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// Setup 注册所有路由
func Setup(app *fiber.App, sm *session.Manager, db *gorm.DB, tplEngine *template.Engine, installH *admin.InstallHandler, cfg *config.Config) {

	// 安装路由（无需认证，必须最先注册）
	app.Get("/install", installH.Page)
	app.Post("/install/test-db", installH.TestDB)
	app.Post("/install/submit", installH.Submit)

	// 安装检测中间件：未安装时跳转 /install
	app.Use(func(c *fiber.Ctx) error {
		if !installH.IsInstalled() && c.Path() != "/install" &&
			c.Path() != "/install/test-db" && c.Path() != "/install/submit" &&
			!strings.HasPrefix(c.Path(), "/static") &&
			!strings.HasPrefix(c.Path(), "/layui/") &&
			!strings.HasPrefix(c.Path(), "/css/") &&
			!strings.HasPrefix(c.Path(), "/js/") &&
			!strings.HasPrefix(c.Path(), "/images/") &&
			!strings.HasPrefix(c.Path(), "/pages/") {
			return c.Redirect("/install")
		}
		return c.Next()
	})

	// 数据库未就绪时提前返回，仅保留安装路由
	// 首页访问由上方中间件统一处理（已安装时放行，未安装时跳转 /install）
	if db == nil {
		// 注册兜底：访问 / 且已安装时提示重启
		app.Get("/", func(c *fiber.Ctx) error {
			if installH.IsInstalled() {
				c.Set("Content-Type", "text/html; charset=utf-8")
				return c.SendString(`<!DOCTYPE html>
<html lang="zh-CN"><head><meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1.0">
<title>已安装 - GoCMS</title>
<style>
*{margin:0;padding:0;box-sizing:border-box}body{font-family:-apple-system,BlinkMacSystemFont,"Segoe UI",Roboto,sans-serif;background:#f0f2f5;min-height:100vh;display:flex;align-items:center;justify-content:center}
.card{background:#fff;border-radius:12px;box-shadow:0 4px 24px rgba(0,0,0,.08);padding:48px;text-align:center;max-width:420px}
.icon{font-size:48px;margin-bottom:16px}h2{font-size:20px;color:#1a1a1a;margin-bottom:8px}p{color:#888;font-size:14px;margin-bottom:24px}
a{display:inline-block;padding:10px 32px;background:#4f46e5;color:#fff;border-radius:8px;text-decoration:none;font-weight:600;transition:background .2s}
a:hover{background:#4338ca}
</style></head><body>
<div class="card"><div class="icon">🔄</div><h2>安装完成</h2><p>数据库已初始化，请重启服务以加载完整功能。</p>
</div></body></html>`)
			}
			return c.Redirect("/install")
		})
		return
	}

	// 业务路由通用防护
	app.Use(middleware.SecurityHeaders())
	app.Use(middleware.RateLimitMiddleware(100, time.Minute))

	// Phase 4 服务
	analyticsSvc := analytics.NewService(db)
	schedulerSvc := scheduler.NewScheduler(db)
	urlPushMgr := urlpush.NewManager(db, urlpush.Config{})

	// Phase 5 服务
	pluginMgr := plugin.NewManager(db)
	aiSvc := aicontent.NewService(aicontent.Config{})
	paymentSvc := payment.NewService(db)
	chatSvc := chat.NewService(db)

	// 注册内置任务
	schedulerSvc.Register("aggregate_daily", analyticsSvc.AggregateDaily)
	schedulerSvc.Register("cache_clean", func() error { return nil })
	schedulerSvc.Register("db_optimize", func() error {
		return db.Exec("OPTIMIZE TABLE mac_vod, mac_art, mac_visit").Error
	})
	schedulerSvc.Register("url_push", func() error { return nil })
	schedulerSvc.InitBuiltinTasks()
	schedulerSvc.Start()

	// 加载已启用的插件
	pluginMgr.LoadEnabled()

	// 后台公开静态资源（放在 /admin 组外，避免路径解析问题）
	app.Static("/layui", "./web/static/admin/layui")
	app.Static("/css", "./web/static/admin/css")
	app.Static("/images", "./web/static/admin/images")
	app.Static("/js", "./web/static/admin/js")
	app.Static("/pages", "./web/static/admin/pages")

	// 后台路由
	setupAdmin(app, sm, db, analyticsSvc, schedulerSvc, urlPushMgr, pluginMgr, aiSvc, paymentSvc, chatSvc, cfg)

	// API 路由
	setupAPI(app, db, chatSvc)

	// 前台路由
	setupFrontend(app, db, sm, analyticsSvc, chatSvc, tplEngine)
}

func setupAdmin(app *fiber.App, sm *session.Manager, db *gorm.DB,
	analyticsSvc *analytics.Service, schedulerSvc *scheduler.Scheduler,
	urlPushMgr *urlpush.Manager, pluginMgr *plugin.Manager,
	aiSvc *aicontent.Service, paymentSvc *payment.Service,
	chatSvc *chat.Service, cfg *config.Config) {

	// 原有 handlers
	dashboard := admin.NewDashboardHandler(db)
	typeH := admin.NewTypeHandler(db)
	vodH := admin.NewVodHandler(db)
	artH := admin.NewArtHandler(db)
	mangaH := admin.NewMangaHandler(db)
	actorH := admin.NewActorHandler(db)
	roleH := admin.NewRoleHandler(db)
	userH := admin.NewUserHandler(db)
	groupH := admin.NewGroupHandler(db)
	adminH := admin.NewAdminHandler(db)
	systemH := admin.NewSystemHandler(db)
	commentH := admin.NewCommentHandler(db)
	gbookH := admin.NewGbookHandler(db)
	collectEngine := collect.NewConcurrentEngine(db, collect.DefaultConfig())
	collectH := admin.NewCollectHandler(db, collectEngine)
	danmakuH := frontend.NewDanmakuHandler(db)

	// Phase 4 handlers
	urlSendH := admin.NewURLSendHandler(db, urlPushMgr)
	analyticsH := admin.NewAnalyticsHandler(analyticsSvc)
	timmingH := admin.NewTimmingHandler(schedulerSvc)

	// Phase 5 handlers
	pluginH := admin.NewPluginHandler(pluginMgr)
	aiH := admin.NewAIContentHandler(aiSvc)
	orderH := admin.NewOrderHandler(paymentSvc)
	liveH := admin.NewLiveHandler(db)
	chatH := admin.NewChatHandler(chatSvc)

	// Phase 6 handlers
	topicH := admin.NewTopicHandler(db)
	linkH := admin.NewLinkHandler(db)
	dbH := admin.NewDatabaseHandler(db)
	tplH := admin.NewTemplateHandler(cfg.Template.Dir)
	uploadH := admin.NewUploadHandler(cfg.Upload.Dir, int64(cfg.Upload.MaxSize)*1024*1024)
	plogH := admin.NewPlogHandler(db)
	dataReplaceH := admin.NewDataReplaceHandler(db)

	// Phase 7 handlers — 统一 JSON API 规范
	systemSettingsH := admin.NewSystemSettingsHandler(db)
	safetyH := admin.NewSafetyHandler(db)
	annexH := admin.NewAnnexHandler(db, cfg.Upload.Dir)
	visitH := admin.NewVisitHandler(db)
	domainH := admin.NewDomainHandler(db)
	makeH := admin.NewMakeHandler(db)
	cjH := admin.NewCJHandler(db)

	a := app.Group("/admin")

	// API 路径兼容：/admin/api/* → /admin/*
	a.Use("/api", func(c *fiber.Ctx) error {
		// 去掉 /api 前缀
		newPath := "/" + c.Params("*")

		// /del → /delete 兼容
		if strings.HasSuffix(newPath, "/del") {
			newPath = strings.TrimSuffix(newPath, "/del") + "/delete"
		}

		// 重写路径
		c.Path(newPath)

		// 如果有 ?id=X 参数且目标路由用 /:id 格式，转换一下
		if id := c.Query("id"); id != "" && !strings.Contains(newPath, "/"+id) {
			c.Path(newPath + "/" + id)
		}

		return c.Next()
	})

	// SPA 静态资源（公开）

	// 验证码（无需认证）
	verifyH := admin.NewVerifyHandler()
	a.Get("/verify", verifyH.Image)

	// SPA 主入口（需登录，未登录跳转登录页）
	a.Get("/", middleware.AdminAuth(sm), func(c *fiber.Ctx) error {
		return c.SendFile("./web/static/admin/index.html")
	})
	a.Get("/login", func(c *fiber.Ctx) error {
		return c.SendFile("./web/static/admin/pages/index/login.html")
	})
	a.Get("/index/login", func(c *fiber.Ctx) error {
		return c.Redirect("/admin/login")
	})
	a.Get("/index/index", func(c *fiber.Ctx) error {
		return c.Redirect("/admin/")
	})
	a.Get("/index/welcome", func(c *fiber.Ctx) error {
		return c.Redirect("/admin/page/welcome")
	})

	// 登录接口（数据库限流，3次失败锁定60分钟）
	a.Post("/login", func(c *fiber.Ctx) error {
		ip := c.IP()

		// 查询封禁记录
		var ban model.LoginBan
		result := db.Where("ip = ?", ip).First(&ban)
		if result.Error == nil && ban.BanUntil > time.Now().Unix() {
			remaining := (ban.BanUntil - time.Now().Unix()) / 60
			if remaining < 1 {
				remaining = 1
			}
			return c.JSON(fiber.Map{"code": 0, "msg": fmt.Sprintf("登录失败次数过多，请 %d 分钟后重试", remaining)})
		}

		// 兼容 JSON body 和 form-data
		getVal := func(key string) string {
			if v := c.FormValue(key); v != "" {
				return v
			}
			var body map[string]string
			if strings.Contains(c.Get("Content-Type"), "json") {
				if c.BodyParser(&body) == nil {
					return body[key]
				}
			}
			return ""
		}
		name := getVal("admin_name")
		pwd := getVal("admin_pwd")
		// 验证码校验
		if vCode := getVal("verify"); vCode != "" {
			if !verifyH.Check(c, vCode) {
				return c.JSON(fiber.Map{"code": 0, "msg": "验证码错误"})
			}
		}
		var adm model.Admin
		if err := db.Where("admin_name = ?", name).First(&adm).Error; err != nil {
			// 记录失败
			ban.IP = ip
			ban.Failures++
			ban.LastFail = time.Now().Unix()
			if ban.Failures >= 3 {
				ban.BanUntil = time.Now().Add(60 * time.Minute).Unix()
			}
			db.Save(&ban)
			return c.JSON(fiber.Map{"code": 0, "msg": "用户名或密码错误"})
		}
		// bcrypt 密码校验（兼容旧的 hex 格式）
		if err := bcrypt.CompareHashAndPassword([]byte(adm.AdminPwd), []byte(pwd)); err != nil {
			if adm.AdminPwd != fmt.Sprintf("%x", []byte(pwd)) {
				// 记录失败
				ban.IP = ip
				ban.Failures++
				ban.LastFail = time.Now().Unix()
				if ban.Failures >= 3 {
					ban.BanUntil = time.Now().Add(60 * time.Minute).Unix()
				}
				db.Save(&ban)
				return c.JSON(fiber.Map{"code": 0, "msg": "用户名或密码错误"})
			}
			if newHash, err := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost); err == nil {
				db.Model(&adm).Update("admin_pwd", string(newHash))
			}
		}
		if adm.AdminStatus != 1 {
			return c.JSON(fiber.Map{"code": 0, "msg": "账号已被禁用"})
		}
		// 登录成功，清除封禁
		db.Where("ip = ?", ip).Delete(&model.LoginBan{})

		sess := sm.Regenerate(c)
		sess.Set("admin_id", strconv.Itoa(adm.AdminID))
		sess.Set("admin_name", adm.AdminName)
		sess.Set("admin_role", strconv.Itoa(adm.AdminRole))
		db.Model(&adm).Updates(map[string]interface{}{
			"admin_last_time": 0,
			"admin_login_num": gorm.Expr("admin_login_num + 1"),
		})
		return c.JSON(fiber.Map{"code": 1, "msg": "登录成功"})
	})

	// 兼容原版 URL 模式的登录接口
	a.Post("/index/login", func(c *fiber.Ctx) error {
		// 复用 /admin/login 的逻辑
		ip := c.IP()
		var ban model.LoginBan
		result := db.Where("ip = ?", ip).First(&ban)
		if result.Error == nil && ban.BanUntil > time.Now().Unix() {
			remaining := (ban.BanUntil - time.Now().Unix()) / 60
			if remaining < 1 {
				remaining = 1
			}
			return c.JSON(fiber.Map{"code": 0, "msg": fmt.Sprintf("登录失败次数过多，请 %d 分钟后重试", remaining)})
		}
		// 兼容 JSON body 和 form-data
		getVal := func(key string) string {
			if v := c.FormValue(key); v != "" {
				return v
			}
			var body map[string]string
			if strings.Contains(c.Get("Content-Type"), "json") {
				if c.BodyParser(&body) == nil {
					return body[key]
				}
			}
			return ""
		}
		name := getVal("admin_name")
		pwd := getVal("admin_pwd")
		// 验证码校验
		if vCode := getVal("verify"); vCode != "" {
			if !verifyH.Check(c, vCode) {
				return c.JSON(fiber.Map{"code": 0, "msg": "验证码错误"})
			}
		}
		var adm model.Admin
		if err := db.Where("admin_name = ?", name).First(&adm).Error; err != nil {
			ban.IP = ip
			ban.Failures++
			ban.LastFail = time.Now().Unix()
			if ban.Failures >= 3 {
				ban.BanUntil = time.Now().Add(60 * time.Minute).Unix()
			}
			db.Save(&ban)
			return c.JSON(fiber.Map{"code": 0, "msg": "用户名或密码错误"})
		}
		if err := bcrypt.CompareHashAndPassword([]byte(adm.AdminPwd), []byte(pwd)); err != nil {
			if adm.AdminPwd != fmt.Sprintf("%x", []byte(pwd)) {
				ban.IP = ip
				ban.Failures++
				ban.LastFail = time.Now().Unix()
				if ban.Failures >= 3 {
					ban.BanUntil = time.Now().Add(60 * time.Minute).Unix()
				}
				db.Save(&ban)
				return c.JSON(fiber.Map{"code": 0, "msg": "用户名或密码错误"})
			}
			if newHash, err := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost); err == nil {
				db.Model(&adm).Update("admin_pwd", string(newHash))
			}
		}
		if adm.AdminStatus != 1 {
			return c.JSON(fiber.Map{"code": 0, "msg": "账号已被禁用"})
		}
		db.Where("ip = ?", ip).Delete(&model.LoginBan{})
		sess := sm.Regenerate(c)
		sess.Set("admin_id", strconv.Itoa(adm.AdminID))
		sess.Set("admin_name", adm.AdminName)
		sess.Set("admin_role", strconv.Itoa(adm.AdminRole))
		db.Model(&adm).Updates(map[string]interface{}{
			"admin_last_time": 0,
			"admin_login_num": gorm.Expr("admin_login_num + 1"),
		})
		return c.JSON(fiber.Map{"code": 1, "msg": "登录成功"})
	})

	auth := a.Group("", middleware.AdminAuth(sm))

	// SPA 静态资源（需登录）
	// js served from root (public)
	// pages served from root (public)

	// --- 原有路由 ---
	auth.Get("/dashboard", dashboard.Index)
	auth.Get("/api/dashboard", dashboard.API) // SPA 纯 JSON 数据接口
	auth.Get("/system/config", systemH.GetConfig)
	auth.Post("/system/config/save", systemH.SaveConfig)
	auth.Post("/system/cache/clear", systemH.CacheClear)
	auth.Get("/type/list", typeH.List)
	auth.Get("/type/tree", typeH.Tree)
	auth.Get("/type/detail/:id", typeH.Detail)
	auth.Post("/type/save", typeH.Save)
	auth.Post("/type/delete/:id", typeH.Delete)
	auth.Post("/type/sort", typeH.Sort)
	auth.Get("/vod/list", vodH.List)
	auth.Get("/vod/detail/:id", vodH.Detail)
	auth.Post("/vod/save", vodH.Save)
	auth.Post("/vod/delete", vodH.Delete)
	auth.Post("/vod/audit/:id", vodH.Audit)
	auth.Post("/vod/batch", vodH.Batch)
	auth.Get("/art/list", artH.List)
	auth.Get("/art/detail/:id", artH.Detail)
	auth.Post("/art/save", artH.Save)
	auth.Post("/art/delete", artH.Delete)
	auth.Post("/art/batch", artH.Batch)
	auth.Get("/manga/list", mangaH.List)
	auth.Get("/manga/detail/:id", mangaH.Detail)
	auth.Post("/manga/save", mangaH.Save)
	auth.Post("/manga/delete", mangaH.Delete)
	auth.Post("/manga/audit/:id", mangaH.Audit)
	auth.Post("/manga/batch", mangaH.Batch)
	auth.Get("/actor/list", actorH.List)
	auth.Get("/actor/detail/:id", actorH.Detail)
	auth.Post("/actor/save", actorH.Save)
	auth.Post("/actor/delete", actorH.Delete)
	auth.Get("/role/list", roleH.List)
	auth.Get("/role/detail/:id", roleH.Detail)
	auth.Post("/role/save", roleH.Save)
	auth.Post("/role/delete", roleH.Delete)
	auth.Get("/user/list", userH.List)
	auth.Get("/user/detail/:id", userH.Detail)
	auth.Post("/user/save", userH.Save)
	auth.Post("/user/delete", userH.Delete)
	auth.Post("/user/toggle/:id", userH.ToggleStatus)
	auth.Get("/group/list", groupH.List)
	auth.Post("/group/save", groupH.Save)
	auth.Get("/admin/list", adminH.List)
	auth.Get("/admin/detail/:id", adminH.Detail)
	auth.Post("/admin/save", adminH.Save)
	auth.Post("/admin/delete/:id", adminH.Delete)
	auth.Get("/comment/list", commentH.List)
	auth.Post("/comment/audit/:id", commentH.Audit)
	auth.Post("/comment/delete", commentH.Delete)
	auth.Get("/gbook/list", gbookH.List)
	auth.Post("/gbook/reply/:id", gbookH.Reply)
	auth.Post("/gbook/delete", gbookH.Delete)
	// 采集管理（兼容 maccms10 协议）
	auth.Get("/collect/source/list", collectH.SourceList)
	auth.Get("/collect/source/detail/:id", collectH.SourceDetail)
	auth.Post("/collect/source/save", collectH.SourceSave)
	auth.Post("/collect/source/delete", collectH.SourceDelete)
	auth.Post("/collect/test", collectH.TestConnection)
	auth.Post("/collect/start", collectH.CollectStart)
	auth.Get("/collect/api", collectH.CollectAPI)              // 兼容 maccms10 collect/api
	auth.Get("/collect/vod/list", collectH.VodList)
	// 分类绑定
	auth.Get("/collect/bind/list", collectH.BindList)
	auth.Post("/collect/bind/save", collectH.BindSave)
	// 采集配置
	auth.Get("/collect/config", collectH.ConfigGet)
	auth.Post("/collect/config", collectH.ConfigSave)
	// 任务管理
	auth.Get("/collect/job/list", collectH.JobList)
	auth.Get("/collect/job/status", collectH.JobStatus)
	auth.Post("/collect/job/stop", collectH.JobStop)

	// --- Phase 4 路由 ---
	auth.Get("/danmaku/list", danmakuH.AdminList)
	auth.Post("/danmaku/delete", danmakuH.AdminDelete)
	auth.Get("/urlsend/config", urlSendH.Config)
	auth.Post("/urlsend/config", urlSendH.Config)
	auth.Post("/urlsend/push", urlSendH.PushURLs)
	auth.Post("/urlsend/pushall", urlSendH.PushAll)
	auth.Get("/urlsend/logs", urlSendH.Logs)
	auth.Post("/urlsend/sitemap", urlSendH.GenerateSitemap)
	auth.Get("/analytics/dashboard", analyticsH.Dashboard)
	auth.Get("/analytics/trend", analyticsH.Trend)
	auth.Get("/analytics/top", analyticsH.TopContent)
	auth.Get("/analytics/regions", analyticsH.Regions)
	auth.Get("/analytics/visits", analyticsH.VisitList)
	auth.Get("/timming/list", timmingH.List)
	auth.Post("/timming/create", timmingH.Create)
	auth.Post("/timming/update", timmingH.Update)
	auth.Post("/timming/delete/:id", timmingH.Delete)
	auth.Post("/timming/toggle/:id", timmingH.Toggle)
	auth.Post("/timming/trigger/:id", timmingH.Trigger)

	// --- Phase 5 路由 ---
	// 插件管理
	auth.Get("/plugin/list", pluginH.List)
	auth.Post("/plugin/install", pluginH.Install)
	auth.Post("/plugin/uninstall/:name", pluginH.Uninstall)
	auth.Post("/plugin/uninstall", pluginH.Uninstall) // 前端兼容（body ids）
	auth.Post("/plugin/enable/:name", pluginH.Enable)
	auth.Post("/plugin/disable/:name", pluginH.Disable)
	auth.Get("/plugin/config/:name", pluginH.Config)
	auth.Post("/plugin/config/:name", pluginH.SaveConfig)

	// AI 内容生成
	auth.Post("/ai/generate", aiH.Generate)
	auth.Post("/ai/title", aiH.GenerateTitle)
	auth.Post("/ai/summary", aiH.GenerateSummary)
	auth.Post("/ai/tags", aiH.GenerateTags)
	auth.Get("/ai/config", aiH.Config)

	// 订单管理
	auth.Get("/order/list", orderH.List)
	auth.Post("/order/pay", orderH.Pay)
	auth.Post("/order/cancel", orderH.Cancel)
	auth.Get("/order/cards", orderH.CardList)
	auth.Post("/order/cards/generate", orderH.GenerateCards)
	auth.Get("/order/payment/config", orderH.PaymentConfig)
	auth.Post("/order/payment/config", orderH.PaymentConfig)

	// 直播管理
	auth.Get("/live/list", liveH.List)
	auth.Get("/live/detail/:id", liveH.Detail)
	auth.Post("/live/save", liveH.Save)
	auth.Post("/live/delete/:id", liveH.Delete)
	auth.Post("/live/toggle/:id", liveH.ToggleStatus)

	// 聊天管理
	auth.Get("/chat/rooms", chatH.RoomList)
	auth.Post("/chat/room/create", chatH.RoomCreate)
	auth.Post("/chat/room/update", chatH.RoomUpdate)
	auth.Post("/chat/room/delete/:id", chatH.RoomDelete)
	auth.Get("/chat/history", chatH.History)
	auth.Get("/chat/online/:id", chatH.OnlineCount)

	// --- Phase 6 路由 ---
	// 专题管理
	auth.Get("/topic/list", topicH.List)
	auth.Get("/topic/detail/:id", topicH.Detail)
	auth.Post("/topic/save", topicH.Save)
	auth.Post("/topic/delete", topicH.Delete)

	// 友情链接
	auth.Get("/link/list", linkH.List)
	auth.Post("/link/save", linkH.Save)
	auth.Post("/link/delete", linkH.Delete)

	// 数据库管理
	auth.Get("/database/list", dbH.List)
	auth.Post("/database/optimize", dbH.Optimize)
	auth.Post("/database/repair", dbH.Repair)
	auth.Post("/database/backup", dbH.Backup)
	auth.Get("/database/backups", dbH.Backups)
	auth.Post("/database/backup-delete", dbH.BackupDelete)
	auth.Post("/database/restore", dbH.Restore)
	auth.Post("/database/sql", dbH.SQL)

	// 模板管理
	auth.Get("/template/list", tplH.List)
	auth.Get("/template/read", tplH.Read)
	auth.Post("/template/save", tplH.Save)
	auth.Get("/template/themes", tplH.Themes)

	// 文件上传
	auth.Post("/upload/file", uploadH.File)
	auth.Post("/upload/image", uploadH.Image)

	// 操作日志
	auth.Get("/plog/list", plogH.List)

	// 数据替换
	auth.Post("/datareplace/execute", dataReplaceH.Execute)

	// --- Phase 7 路由 — 统一 JSON API ---
	// 系统设置
	auth.Get("/system/settings", systemSettingsH.GetAllConfig)
	auth.Get("/system/settings/group/:group", systemSettingsH.GetGroupConfig)
	auth.Post("/system/settings/save", systemSettingsH.SaveConfig)
	auth.Post("/system/test-email", systemSettingsH.TestEmail)
	auth.Post("/system/test-cache", systemSettingsH.TestCache)
	auth.Get("/system/groups", systemSettingsH.GetConfigGroupList)

	// 安全配置
	auth.Get("/safety/config", safetyH.GetConfig)
	auth.Post("/safety/save", safetyH.SaveConfig)
	auth.Get("/safety/ipblacklist", safetyH.IPBlacklistList)
	auth.Post("/safety/ipblacklist/add", safetyH.IPBlacklistAdd)
	auth.Post("/safety/ipblacklist/delete", safetyH.IPBlacklistDelete)

	// 附件管理
	auth.Get("/annex/list", annexH.List)
	auth.Post("/annex/delete", annexH.Delete)
	auth.Post("/annex/batch-delete", annexH.BatchDelete)

	// 访问日志
	auth.Get("/visit/list", visitH.List)
	auth.Get("/visit/stats", visitH.Stats)

	// 域名管理
	auth.Get("/domain/list", domainH.List)
	auth.Post("/domain/save", domainH.Save)
	auth.Post("/domain/delete", domainH.Delete)

	// 静态生成
	auth.Post("/make/start", makeH.Start)
	auth.Post("/make/run", makeH.Start) // 前端兼容别名
	auth.Get("/make/status", makeH.Status)

	// 自定义采集
	auth.Get("/cj/list", cjH.List)
	auth.Get("/cj/detail/:id", cjH.Detail)
	auth.Post("/cj/save", cjH.Save)
	auth.Post("/cj/run", cjH.Run)

	// --- SPA 补充路由（前端需要但原路由未覆盖的） ---
	// 登录页需要的站点配置
	auth.Get("/config", systemH.GetConfig)
	// 欢迎页系统信息
	auth.Get("/systemInfo", dashboard.API)
	// 播放器/服务器/下载源列表（视频编辑页需要）— 暂返空
	auth.Get("/player/list", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"code": 1, "data": []interface{}{}, "msg": "success"})
	})
	auth.Get("/server/list", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"code": 1, "data": []interface{}{}, "msg": "success"})
	})
	auth.Get("/downer/list", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"code": 1, "data": []interface{}{}, "msg": "success"})
	})
	// 分类扩展字段
	auth.Post("/type/extend", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"code": 1, "data": map[string]interface{}{}, "msg": "success"})
	})
	// 安全列表 → IP黑名单
	auth.Get("/safety/list", safetyH.IPBlacklistList)

	// SPA 缺失路由 stub
	auth.Get("/comment/detail/:id", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"code": 1, "data": map[string]interface{}{}, "msg": "success"})
	})
	auth.Get("/gbook/detail/:id", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"code": 1, "data": map[string]interface{}{}, "msg": "success"})
	})
	auth.Get("/group/detail/:id", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"code": 1, "data": map[string]interface{}{}, "msg": "success"})
	})
	auth.Get("/link/detail/:id", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"code": 1, "data": map[string]interface{}{}, "msg": "success"})
	})
	auth.Get("/addon/detail/:id", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"code": 1, "data": map[string]interface{}{}, "msg": "success"})
	})
	auth.Get("/addon/list", pluginH.List)
	auth.Post("/addon/delete/:id", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"code": 1, "msg": "success"})
	})
	auth.Get("/card/list", orderH.CardList)
	auth.Get("/card/detail/:id", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"code": 1, "data": map[string]interface{}{}, "msg": "success"})
	})
	auth.Post("/card/delete/:id", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"code": 1, "msg": "success"})
	})
	auth.Post("/database/delete", dbH.Backup) // placeholder
	auth.Post("/plog/delete", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"code": 1, "msg": "success"})
	})
	auth.Post("/safety/delete", safetyH.IPBlacklistDelete)
	auth.Post("/template/delete", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"code": 1, "msg": "success"})
	})
	auth.Get("/timming/detail/:id", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"code": 1, "data": map[string]interface{}{}, "msg": "success"})
	})
	auth.Post("/cj/delete", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"code": 1, "msg": "success"})
	})
	auth.Post("/group/delete/:id", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"code": 1, "msg": "success"})
	})

	// --- SPA 页面路由（替换原 Go 模板渲染） ---
	spaPage := func(file string) fiber.Handler {
		return func(c *fiber.Ctx) error {
			return c.SendFile("./web/static/admin/pages/" + file)
		}
	}
	auth.Get("/page/welcome", spaPage("index/welcome.html"))
	auth.Get("/page/vod/data", spaPage("vod/index.html"))
	auth.Get("/page/vod/info", spaPage("vod/info.html"))
	auth.Get("/page/vod/info/:id", spaPage("vod/info.html"))
	auth.Get("/page/art/data", spaPage("art/index.html"))
	auth.Get("/page/art/info", spaPage("art/info.html"))
	auth.Get("/page/art/info/:id", spaPage("art/info.html"))
	auth.Get("/page/topic/data", spaPage("topic/index.html"))
	auth.Get("/page/topic/info", spaPage("topic/info.html"))
	auth.Get("/page/topic/info/:id", spaPage("topic/info.html"))
	auth.Get("/page/link/index", spaPage("link/index.html"))
	auth.Get("/page/link/info", spaPage("link/info.html"))
	auth.Get("/page/link/info/:id", spaPage("link/info.html"))
	auth.Get("/page/type/index", spaPage("type/index.html"))
	auth.Get("/page/type/info", spaPage("type/info.html"))
	auth.Get("/page/type/info/:id", spaPage("type/info.html"))
	auth.Get("/page/actor/data", spaPage("actor/index.html"))
	auth.Get("/page/actor/info", spaPage("actor/info.html"))
	auth.Get("/page/actor/info/:id", spaPage("actor/info.html"))
	auth.Get("/page/role/data", spaPage("role/index.html"))
	auth.Get("/page/role/info", spaPage("role/info.html"))
	auth.Get("/page/role/info/:id", spaPage("role/info.html"))
	auth.Get("/page/user/data", spaPage("user/index.html"))
	auth.Get("/page/user/info", spaPage("user/info.html"))
	auth.Get("/page/user/info/:id", spaPage("user/info.html"))
	auth.Get("/page/admin/index", spaPage("admin/index.html"))
	auth.Get("/page/admin/info", spaPage("admin/info.html"))
	auth.Get("/page/admin/info/:id", spaPage("admin/info.html"))
	auth.Get("/page/comment/data", spaPage("comment/index.html"))
	auth.Get("/page/gbook/data", spaPage("gbook/index.html"))
	auth.Get("/page/database/export", spaPage("database/export.html"))
	auth.Get("/page/database/sql", spaPage("database/sql.html"))
	auth.Get("/page/database/rep", spaPage("database/rep.html"))
	auth.Get("/page/template/index", spaPage("template/index.html"))
	auth.Get("/page/plog/index", spaPage("plog/index.html"))
	auth.Get("/page/collect/index", spaPage("collect/index.html"))
	auth.Get("/page/order/index", spaPage("card/index.html"))
	auth.Get("/page/manga/data", spaPage("manga/index.html"))
	auth.Get("/page/manga/info", spaPage("manga/info.html"))
	auth.Get("/page/manga/info/:id", spaPage("manga/info.html"))
	auth.Get("/page/live/index", spaPage("addon/index.html"))
	auth.Get("/page/system/config", spaPage("system/config.html"))
	auth.Get("/page/safety/index", spaPage("safety/index.html"))
	auth.Get("/page/annex/index", spaPage("annex/index.html"))
	auth.Get("/page/visit/index", spaPage("visit/index.html"))
	auth.Get("/page/timming/index", spaPage("timming/index.html"))
	auth.Get("/page/cj/index", spaPage("cj/index.html"))
	auth.Get("/page/danmaku/index", spaPage("comment/index.html"))
	auth.Get("/page/plugin/index", spaPage("addon/index.html"))
	auth.Get("/page/ai/index", spaPage("addon/index.html"))
	auth.Get("/page/chat/index", spaPage("addon/index.html"))
	auth.Get("/page/urlsend/index", spaPage("urlsend/index.html"))
	auth.Get("/page/urlsend/sitemap", spaPage("urlsend/index.html"))
	auth.Get("/page/database/index", spaPage("database/index.html"))
	auth.Get("/page/gbook/index", spaPage("gbook/index.html"))
	auth.Get("/page/comment/index", spaPage("comment/index.html"))

	// 登出
	auth.Post("/logout", func(c *fiber.Ctx) error {
		sm.Destroy(c)
		return c.JSON(fiber.Map{"code": 1, "msg": "已退出"})
	})
}

func setupAPI(app *fiber.App, db *gorm.DB, chatSvc *chat.Service) {
	provide := api.NewProvideHandler(db)
	danmaku := frontend.NewDanmakuHandler(db)

	apiGroup := app.Group("/api")
	apiGroup.Get("/provide/:ac", provide.ProvideAPI)
	apiGroup.Get("/provide/search", provide.ProvideSearch)

	// 弹幕 API
	apiGroup.Post("/danmaku/:vod_id/send", danmaku.Send)
	apiGroup.Get("/danmaku/:vod_id/history", danmaku.History)
	apiGroup.Get("/danmaku/:vod_id/online", danmaku.OnlineCount)
}

func setupFrontend(app *fiber.App, db *gorm.DB, sm *session.Manager,
	analyticsSvc *analytics.Service, chatSvc *chat.Service, tplEngine *template.Engine) {

	index := frontend.NewIndexHandler(db, tplEngine)
	vod := frontend.NewVodHandler(db, tplEngine)
	art := frontend.NewArtHandler(db, tplEngine)
	manga := frontend.NewMangaHandler(db, tplEngine)
	actor := frontend.NewActorHandler(db, tplEngine)
	role := frontend.NewRoleHandler(db, tplEngine)
	topic := frontend.NewTopicHandler(db, tplEngine)
	user := frontend.NewUserHandler(db, sm)
	gbook := frontend.NewGbookHandler(db, tplEngine)
	chatH := frontend.NewChatHandler(chatSvc)
	danmaku := frontend.NewDanmakuHandler(db)

	// 访问记录中间件
	app.Use(func(c *fiber.Ctx) error {
		err := c.Next()
		analyticsSvc.RecordVisitAsync(c.Path(), c.IP(), c.Get("User-Agent"), c.Get("Referer"))
		return err
	})

	// 首页
	app.Get("/", index.Index)

	// 视频
	app.Get("/vodtype/:id", vod.Type)
	app.Get("/vodtype/:id/:page", vod.Type)
	app.Get("/voddetail/:id", vod.Detail)
	app.Get("/vodplay/:id/:sid/:nid", vod.Play)
	app.Get("/vodsearch", vod.Search)
	app.Get("/vodshow/:id", vod.Show)

	// 文章
	app.Get("/arttype/:id", art.Type)
	app.Get("/arttype/:id/:page", art.Type)
	app.Get("/artdetail/:id", art.Detail)
	app.Get("/artsearch", art.Search)

	// 漫画
	app.Get("/mangatype/:id", manga.Type)
	app.Get("/mangatype/:id/:page", manga.Type)
	app.Get("/mangadetail/:id", manga.Detail)
	app.Get("/mangashow/:id", manga.Show)
	app.Get("/mangasearch", manga.Search)

	// 演员
	app.Get("/actor", actor.Index)
	app.Get("/actor/:page", actor.Index)
	app.Get("/actordetail/:id", actor.Detail)
	app.Get("/actorshow", actor.Show)

	// 角色
	app.Get("/role", role.Index)
	app.Get("/role/:page", role.Index)
	app.Get("/roledetail/:id", role.Detail)
	app.Get("/roleshow", role.Show)

	// 专题
	app.Get("/topic", topic.Index)
	app.Get("/topic/:page", topic.Index)
	app.Get("/topicdetail/:id", topic.Detail)

	// 用户
	app.Post("/user/register", user.Register)
	app.Post("/user/login", user.Login)
	app.Get("/user/info", user.Info)
	app.Post("/user/logout", user.Logout)

	// 留言
	app.Get("/gbook", gbook.Index)
	app.Get("/gbook/:page", gbook.Index)
	app.Post("/gbook/submit", gbook.Submit)

	// 弹幕 WebSocket
	app.Get("/ws/danmaku/:vod_id", danmaku.WebSocketUpgrade, func(c *fiber.Ctx) error {
		return nil
	})

	// 聊天室 WebSocket
	app.Get("/ws/chat", chatH.WebSocketUpgrade, func(c *fiber.Ctx) error {
		return nil
	})

	// Catch-All
	app.Get("/*", func(c *fiber.Ctx) error {
		return c.Status(404).JSON(fiber.Map{"code": 404, "msg": "页面不存在"})
	})
}
