package router

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
	"gocms/internal/handler/admin"
	"gocms/internal/handler/api"
	"gocms/internal/handler/frontend"
	"gocms/internal/middleware"
	"gocms/internal/model"
	"gocms/internal/session"
	"gocms/internal/service/analytics"
	"gocms/internal/service/scheduler"
	"gocms/internal/service/urlpush"
)

// Setup 注册所有路由
func Setup(app *fiber.App, sm *session.Manager, db *gorm.DB) {
	// 初始化 Phase 4 服务
	analyticsSvc := analytics.NewService(db)
	schedulerSvc := scheduler.NewScheduler(db)
	urlPushMgr := urlpush.NewManager(db, urlpush.Config{})

	// 注册内置任务函数
	schedulerSvc.Register("aggregate_daily", analyticsSvc.AggregateDaily)
	schedulerSvc.Register("cache_clean", func() error { return nil })
	schedulerSvc.Register("db_optimize", func() error {
		return db.Exec("OPTIMIZE TABLE mac_vod, mac_art, mac_visit").Error
	})
	schedulerSvc.Register("url_push", func() error {
		// TODO: 自动推送最近更新的内容URL
		return nil
	})

	// 初始化内置任务（数据库中不存在则创建）
	schedulerSvc.InitBuiltinTasks()

	// 启动调度器
	schedulerSvc.Start()

	// 后台路由
	setupAdmin(app, sm, db, analyticsSvc, schedulerSvc, urlPushMgr)

	// API 路由
	setupAPI(app, db)

	// 前台路由（含访问记录中间件）
	setupFrontend(app, db, sm, analyticsSvc)
}

func setupAdmin(app *fiber.App, sm *session.Manager, db *gorm.DB,
	analyticsSvc *analytics.Service, schedulerSvc *scheduler.Scheduler,
	urlPushMgr *urlpush.Manager) {

	// 初始化 handlers
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
	urlSendH := admin.NewURLSendHandler(db, urlPushMgr)
	analyticsH := admin.NewAnalyticsHandler(analyticsSvc)
	timmingH := admin.NewTimmingHandler(schedulerSvc)

	a := app.Group("/admin")

	// 登录（无需鉴权）
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

	// 需要鉴权的路由
	auth := a.Group("", middleware.AdminAuth(sm))

	// 仪表盘
	auth.Get("/dashboard", dashboard.Index)

	// 系统设置
	auth.Get("/system/config", systemH.GetConfig)
	auth.Post("/system/config/save", systemH.SaveConfig)
	auth.Post("/system/cache/clear", systemH.CacheClear)

	// 分类管理
	auth.Get("/type/list", typeH.List)
	auth.Get("/type/tree", typeH.Tree)
	auth.Get("/type/detail/:id", typeH.Detail)
	auth.Post("/type/save", typeH.Save)
	auth.Post("/type/delete/:id", typeH.Delete)
	auth.Post("/type/sort", typeH.Sort)

	// 视频管理
	auth.Get("/vod/list", vodH.List)
	auth.Get("/vod/detail/:id", vodH.Detail)
	auth.Post("/vod/save", vodH.Save)
	auth.Post("/vod/delete", vodH.Delete)
	auth.Post("/vod/audit/:id", vodH.Audit)
	auth.Post("/vod/batch", vodH.Batch)

	// 文章管理
	auth.Get("/art/list", artH.List)
	auth.Get("/art/detail/:id", artH.Detail)
	auth.Post("/art/save", artH.Save)
	auth.Post("/art/delete", artH.Delete)

	// 漫画管理
	auth.Get("/manga/list", mangaH.List)
	auth.Get("/manga/detail/:id", mangaH.Detail)
	auth.Post("/manga/save", mangaH.Save)
	auth.Post("/manga/delete", mangaH.Delete)
	auth.Post("/manga/audit/:id", mangaH.Audit)

	// 演员管理
	auth.Get("/actor/list", actorH.List)
	auth.Get("/actor/detail/:id", actorH.Detail)
	auth.Post("/actor/save", actorH.Save)
	auth.Post("/actor/delete", actorH.Delete)

	// 角色管理
	auth.Get("/role/list", roleH.List)
	auth.Get("/role/detail/:id", roleH.Detail)
	auth.Post("/role/save", roleH.Save)
	auth.Post("/role/delete", roleH.Delete)

	// 用户管理
	auth.Get("/user/list", userH.List)
	auth.Post("/user/save", userH.Save)
	auth.Post("/user/delete", userH.Delete)
	auth.Post("/user/toggle/:id", userH.ToggleStatus)

	// 用户组
	auth.Get("/group/list", groupH.List)
	auth.Post("/group/save", groupH.Save)

	// 管理员管理
	auth.Get("/admin/list", adminH.List)
	auth.Post("/admin/save", adminH.Save)
	auth.Post("/admin/delete/:id", adminH.Delete)

	// 评论管理
	auth.Get("/comment/list", commentH.List)
	auth.Post("/comment/audit/:id", commentH.Audit)
	auth.Post("/comment/delete", commentH.Delete)

	// 留言管理
	auth.Get("/gbook/list", gbookH.List)
	auth.Post("/gbook/reply/:id", gbookH.Reply)
	auth.Post("/gbook/delete", gbookH.Delete)

	// 采集管理
	auth.Post("/collect/test", collectH.TestConnection)
	auth.Post("/collect/start", collectH.StartCollect)

	// 弹幕管理
	auth.Get("/danmaku/list", danmakuH.AdminList)
	auth.Post("/danmaku/delete", danmakuH.AdminDelete)

	// URL 推送
	auth.Get("/urlsend/config", urlSendH.Config)
	auth.Post("/urlsend/config", urlSendH.Config)
	auth.Post("/urlsend/push", urlSendH.PushURLs)
	auth.Post("/urlsend/pushall", urlSendH.PushAll)
	auth.Get("/urlsend/logs", urlSendH.Logs)
	auth.Post("/urlsend/sitemap", urlSendH.GenerateSitemap)

	// 数据分析
	auth.Get("/analytics/dashboard", analyticsH.Dashboard)
	auth.Get("/analytics/trend", analyticsH.Trend)
	auth.Get("/analytics/top", analyticsH.TopContent)
	auth.Get("/analytics/regions", analyticsH.Regions)
	auth.Get("/analytics/visits", analyticsH.VisitList)

	// 定时任务
	auth.Get("/timming/list", timmingH.List)
	auth.Post("/timming/create", timmingH.Create)
	auth.Post("/timming/update", timmingH.Update)
	auth.Post("/timming/delete/:id", timmingH.Delete)
	auth.Post("/timming/toggle/:id", timmingH.Toggle)
	auth.Post("/timming/trigger/:id", timmingH.Trigger)

	// 登出
	auth.Post("/logout", func(c *fiber.Ctx) error {
		sm.Destroy(c)
		return c.JSON(fiber.Map{"code": 1, "msg": "已退出"})
	})
}

func setupAPI(app *fiber.App, db *gorm.DB) {
	provide := api.NewProvideHandler(db)
	danmaku := frontend.NewDanmakuHandler(db)

	apiGroup := app.Group("/api")
	apiGroup.Get("/provide/:ac", provide.ProvideAPI)
	apiGroup.Get("/provide/search", provide.ProvideSearch)

	// 弹幕 HTTP API
	apiGroup.Post("/danmaku/:vod_id/send", danmaku.Send)
	apiGroup.Get("/danmaku/:vod_id/history", danmaku.History)
	apiGroup.Get("/danmaku/:vod_id/online", danmaku.OnlineCount)
}

func setupFrontend(app *fiber.App, db *gorm.DB, sm *session.Manager, analyticsSvc *analytics.Service) {
	index := frontend.NewIndexHandler(db)
	vod := frontend.NewVodHandler(db)
	art := frontend.NewArtHandler(db)
	manga := frontend.NewMangaHandler(db)
	actor := frontend.NewActorHandler(db)
	role := frontend.NewRoleHandler(db)
	topic := frontend.NewTopicHandler(db)
	user := frontend.NewUserHandler(db, sm)
	gbook := frontend.NewGbookHandler(db)

	// 访问记录中间件（异步记录，不影响响应速度）
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
	danmaku := frontend.NewDanmakuHandler(db)
	app.Get("/ws/danmaku/:vod_id", danmaku.WebSocketUpgrade, func(c *fiber.Ctx) error {
		return nil
	})

	// Catch-All（兜底）
	app.Get("/*", func(c *fiber.Ctx) error {
		return c.Status(404).JSON(fiber.Map{"code": 404, "msg": "页面不存在"})
	})
}
