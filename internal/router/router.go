package router

import (
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
	"gocms/internal/handler/admin"
	"gocms/internal/handler/api"
	"gocms/internal/handler/frontend"
	"gocms/internal/middleware"
	"gocms/internal/model"
	"gocms/internal/session"
	"gocms/internal/template"
	"gocms/internal/service/aicontent"
	"gocms/internal/service/analytics"
	"gocms/internal/service/chat"
	"gocms/internal/service/payment"
	"gocms/internal/service/plugin"
	"gocms/internal/service/scheduler"
	"gocms/internal/service/urlpush"
)

// Setup 注册所有路由
func Setup(app *fiber.App, sm *session.Manager, db *gorm.DB, tplEngine *template.Engine, installH *admin.InstallHandler) {

	// 安装路由（无需认证，必须最先注册）
	app.Get("/install", installH.Page)
	app.Post("/install/test-db", installH.TestDB)
	app.Post("/install/submit", installH.Submit)

	// 安装检测中间件：未安装时跳转 /install
	app.Use(func(c *fiber.Ctx) error {
		if !installH.IsInstalled() && c.Path() != "/install" &&
			c.Path() != "/install/test-db" && c.Path() != "/install/submit" &&
			!strings.HasPrefix(c.Path(), "/static") {
			return c.Redirect("/install")
		}
		return c.Next()
	})

	// 数据库未就绪时只注册安装路由
	if db == nil {
		app.Get("/", func(c *fiber.Ctx) error {
			return c.Redirect("/install")
		})
		return
	}

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

	// 后台路由
	setupAdmin(app, sm, db, analyticsSvc, schedulerSvc, urlPushMgr, pluginMgr, aiSvc, paymentSvc, chatSvc)

	// API 路由
	setupAPI(app, db, chatSvc)

	// 前台路由
	setupFrontend(app, db, sm, analyticsSvc, chatSvc, tplEngine)
}

func setupAdmin(app *fiber.App, sm *session.Manager, db *gorm.DB,
	analyticsSvc *analytics.Service, schedulerSvc *scheduler.Scheduler,
	urlPushMgr *urlpush.Manager, pluginMgr *plugin.Manager,
	aiSvc *aicontent.Service, paymentSvc *payment.Service,
	chatSvc *chat.Service) {

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
	collectH := admin.NewCollectHandler(db)
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

	a := app.Group("/admin")

	// 登录
	a.Post("/login", func(c *fiber.Ctx) error {
		name := c.FormValue("admin_name")
		pwd := c.FormValue("admin_pwd")
		var adm model.Admin
		if err := db.Where("admin_name = ? AND admin_pwd = ?", name, pwd).First(&adm).Error; err != nil {
			return c.JSON(fiber.Map{"code": 0, "msg": "用户名或密码错误"})
		}
		if adm.AdminStatus != 1 {
			return c.JSON(fiber.Map{"code": 0, "msg": "账号已被禁用"})
		}
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

	// --- 原有路由 ---
	auth.Get("/dashboard", dashboard.Index)
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
	auth.Get("/manga/list", mangaH.List)
	auth.Get("/manga/detail/:id", mangaH.Detail)
	auth.Post("/manga/save", mangaH.Save)
	auth.Post("/manga/delete", mangaH.Delete)
	auth.Post("/manga/audit/:id", mangaH.Audit)
	auth.Get("/actor/list", actorH.List)
	auth.Get("/actor/detail/:id", actorH.Detail)
	auth.Post("/actor/save", actorH.Save)
	auth.Post("/actor/delete", actorH.Delete)
	auth.Get("/role/list", roleH.List)
	auth.Get("/role/detail/:id", roleH.Detail)
	auth.Post("/role/save", roleH.Save)
	auth.Post("/role/delete", roleH.Delete)
	auth.Get("/user/list", userH.List)
	auth.Post("/user/save", userH.Save)
	auth.Post("/user/delete", userH.Delete)
	auth.Post("/user/toggle/:id", userH.ToggleStatus)
	auth.Get("/group/list", groupH.List)
	auth.Post("/group/save", groupH.Save)
	auth.Get("/admin/list", adminH.List)
	auth.Post("/admin/save", adminH.Save)
	auth.Post("/admin/delete/:id", adminH.Delete)
	auth.Get("/comment/list", commentH.List)
	auth.Post("/comment/audit/:id", commentH.Audit)
	auth.Post("/comment/delete", commentH.Delete)
	auth.Get("/gbook/list", gbookH.List)
	auth.Post("/gbook/reply/:id", gbookH.Reply)
	auth.Post("/gbook/delete", gbookH.Delete)
	auth.Post("/collect/test", collectH.TestConnection)
	auth.Post("/collect/start", collectH.StartCollect)

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

	index := frontend.NewIndexHandler(db)
	vod := frontend.NewVodHandler(db)
	art := frontend.NewArtHandler(db)
	manga := frontend.NewMangaHandler(db)
	actor := frontend.NewActorHandler(db)
	role := frontend.NewRoleHandler(db)
	topic := frontend.NewTopicHandler(db)
	user := frontend.NewUserHandler(db, sm)
	gbook := frontend.NewGbookHandler(db)
	chatH := frontend.NewChatHandler(chatSvc)
	danmaku := frontend.NewDanmakuHandler(db)

	// 安全中间件
	app.Use(middleware.SecurityHeaders())
	app.Use(middleware.RateLimitMiddleware(100, 60)) // 每分钟100次

	// 访问记录中间件
	app.Use(func(c *fiber.Ctx) error {
		err := c.Next()
		go analyticsSvc.RecordVisit(c.Path(), c.IP(), c.Get("User-Agent"), c.Get("Referer"))
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
