# 苹果CMS v10 → Go + Fiber 重写技术分析报告

> 源项目: https://github.com/magicblack/maccms10
> 生成时间: 2026-06-09
> 目标: 完整功能重写，去除所有暗桩后门

---

## 一、源项目概览

### 1.1 基本信息

| 项目 | 说明 |
|------|------|
| 名称 | 苹果CMS (MacCMS) v10 |
| 版本号 | 2026.1000.4053 |
| 语言 | PHP 7.x+ |
| 框架 | ThinkPHP 5.0.24 |
| 数据库 | MySQL (utf8) |
| 表前缀 | `mac_` |
| 架构 | MVC (仿MVC模板分离) |
| 许可证 | Apache 2.0 |

### 1.2 技术栈分析

- **后端**: PHP + ThinkPHP 5.x (tp5内核)
- **数据库**: MySQL, 使用 ThinkPHP ORM
- **模板引擎**: 自定义扩展模板引擎 + 原生HTML模板分离
- **缓存**: 文件缓存为主
- **前端**: HTML/CSS/JS, jQuery, Bootstrap (后台管理)
- **采集**: 自定义采集模块 (curl)
- **搜索引擎**: 可选 Meilisearch 集成

---

## 二、源项目目录结构与模块映射

### 2.1 顶层目录

```
maccms10/
├── admin.php              # 后台入口
├── api.php                # API入口
├── index.php              # 前台入口
├── application/           # 核心应用代码
│   ├── admin/             # 后台管理模块
│   ├── api/               # API接口模块
│   ├── common.php         # 全局公共函数
│   ├── database.php       # 数据库配置
│   ├── extra/             # 扩展配置
│   ├── index/             # 前台模块
│   ├── route.php          # 路由定义
│   └── ...
├── addons/                # 插件目录
├── extend/                # 扩展类库
├── public/                # 静态资源
├── runtime/               # 运行时缓存
├── template/              # 前台模板
├── thinkphp/              # ThinkPHP框架核心
└── vendor/                # Composer依赖
```

### 2.2 后台控制器完整清单 (admin/controller/)

| 控制器 | 功能 | Go重写模块 |
|--------|------|-----------|
| `Index.php` (38KB) | 仪表盘/首页/系统概览 | `handler/admin_dashboard.go` |
| `System.php` (80KB) | 系统设置(核心最大文件) | `handler/admin_system.go` |
| `Vod.php` (31KB) | 视频管理 | `handler/admin_vod.go` |
| `Art.php` (15KB) | 文章管理 | `handler/admin_art.go` |
| `Actor.php` | 演员管理 | `handler/admin_actor.go` |
| `Role.php` | 角色管理 | `handler/admin_role.go` |
| `Type.php` | 分类管理 | `handler/admin_type.go` |
| `User.php` | 用户管理 | `handler/admin_user.go` |
| `Group.php` | 用户组管理 | `handler/admin_group.go` |
| `Admin.php` | 管理员管理 | `handler/admin_admin.go` |
| `Adminaudit.php` | 管理员审计 | `handler/admin_audit.go` |
| `Comment.php` | 评论管理 | `handler/admin_comment.go` |
| `Gbook.php` | 留言板管理 | `handler/admin_gbook.go` |
| `Topic.php` | 专题管理 | `handler/admin_topic.go` |
| `Link.php` | 友情链接管理 | `handler/admin_link.go` |
| `Collect.php` (21KB) | 采集管理 | `handler/admin_collect.go` |
| `Cj.php` (18KB) | 采集规则管理 | `handler/admin_cj.go` |
| `Database.php` (14KB) | 数据库管理(备份/恢复/优化) | `handler/admin_database.go` |
| `Template.php` | 模板管理 | `handler/admin_template.go` |
| `TplConfig.php` (17KB) | 模板配置 | `handler/admin_tplconfig.go` |
| `Make.php` (33KB) | 静态页面生成 | `handler/admin_make.go` |
| `Upload.php` | 上传管理 | `handler/admin_upload.go` |
| `Annex.php` (13KB) | 附件管理 | `handler/admin_annex.go` |
| `Images.php` | 图片管理 | `handler/admin_images.go` |
| `Manga.php` (15KB) | 漫画管理 | `handler/admin_manga.go` |
| `Card.php` | 卡密管理 | `handler/admin_card.go` |
| `Order.php` | 订单管理 | `handler/admin_order.go` |
| `Cash.php` | 提现管理 | `handler/admin_cash.go` |
| `Live.php` | 直播管理 | `handler/admin_live.go` |
| `Danmaku.php` | 弹幕管理 | `handler/admin_danmaku.go` |
| `Chatroom.php` | 聊天室管理 | `handler/admin_chatroom.go` |
| `Domain.php` | 域名管理 | `handler/admin_domain.go` |
| `Urlsend.php` (8KB) | URL推送(百度/神马/搜狗) | `handler/admin_urlsend.go` |
| `Safety.php` | 安全设置 | `handler/admin_safety.go` |
| `Website.php` (12KB) | 站点管理 | `handler/admin_website.go` |
| `Addon.php` (12KB) | 插件管理 | `handler/admin_addon.go` |
| `ResourceHub.php` (50KB) | 资源中心 | `handler/admin_resourcehub.go` |
| `Meilisearch.php` (11KB) | Meilisearch搜索管理 | `handler/admin_meilisearch.go` |
| `Analytics.php` (23KB) | 数据分析/统计 | `handler/admin_analytics.go` |
| `BatchPlayer.php` | 批量播放器管理 | `handler/admin_batchplayer.go` |
| `Vodplayer.php` | 视频播放器管理 | `handler/admin_vodplayer.go` |
| `Vodserver.php` | 视频服务器管理 | `handler/admin_vodserver.go` |
| `Voddowner.php` | 视频下载管理 | `handler/admin_voddowner.go` |
| `DataReplace.php` | 数据替换 | `handler/admin_datareplace.go` |
| `Plog.php` | 操作日志 | `handler/admin_plog.go` |
| `Ulog.php` | 用户日志 | `handler/admin_ulog.go` |
| `Visit.php` | 访问统计 | `handler/admin_visit.go` |
| `Update.php` | 系统更新 | `handler/admin_update.go` |
| `SignMilestone.php` | 签到里程碑 | `handler/admin_signmilestone.go` |
| `Timming.php` | 定时任务 | `handler/admin_timming.go` |
| `Task.php` | 任务管理 | `handler/admin_task.go` |
| `Assistant.php` | 助手工具 | `handler/admin_assistant.go` |
| `Base.php` | 基础控制器 | `handler/admin_base.go` |

### 2.3 前台模块 (index/controller/)

| 控制器 | 功能 |
|--------|------|
| `Vod.php` | 视频浏览/播放/下载/搜索 |
| `Art.php` | 文章浏览/阅读 |
| `Actor.php` | 演员列表/详情 |
| `Role.php` | 角色列表/详情 |
| `Manga.php` | 漫画浏览/阅读 |
| `Index.php` | 首页/分类页 |
| `Gbook.php` | 留言板 |
| `Topic.php` | 专题页 |
| `Map.php` | 站点地图 |
| `Rss.php` | RSS订阅 |
| `User.php` | 用户中心 |
| `Label.php` | 自定义标签页 |

### 2.4 API模块 (api/controller/)

| 控制器 | 功能 |
|--------|------|
| `Provide.php` | 资源提供接口(采集数据源) |
| `User.php` | 用户API(登录/注册/信息) |
| `Comment.php` | 评论API |
| `Gbook.php` | 留言API |
| `Danmaku.php` | 弹幕API |
| `Card.php` | 卡密API |
| `Order.php` | 订单/支付API |

### 2.5 插件系统 (addons/)

| 插件 | 功能 |
|------|------|
| `adminloginbg` | 后台登录背景自定义 |
| `aicontent` | AI内容生成(集成DeepSeek/OpenAI/Claude/Gemini/GLM/Qwen) |

---

## 三、数据模型分析

### 3.1 核心数据库表 (mac_前缀)

```
mac_vod              # 视频主表
mac_vod_play         # 播放源/播放地址
mac_vod_down         # 下载源/下载地址
mac_art              # 文章表
mac_actor            # 演员表
mac_role             # 角色表
mac_type             # 分类表(树形结构)
mac_admin            # 管理员表
mac_user             # 用户表
mac_group            # 用户组表
mac_comment          # 评论表
mac_gbook            # 留言表
mac_topic            # 专题表
mac_link             # 友情链接表
mac_card             # 卡密表
mac_order            # 订单表
mac_cash             # 提现表
mac_collect          # 采集规则表
mac_cj_data          # 采集数据表
mac_manga            # 漫画表
mac_manga_play       # 漫画播放源
mac_manga_down       # 漫画下载源
mac_live             # 直播表
mac_danmaku          # 弹幕表
mac_chatroom         # 聊天室表
mac_images           # 图片表
mac_domain           # 域名表
mac_visit            # 访问记录表
mac_ulog             # 用户日志表
mac_plog             # 操作日志表
mac_adminaudit       # 管理员审计表
mac_task             # 任务表
mac_website          # 站点表
mac_sign_milestone   # 签到里程碑表
mac_config           # 系统配置表
mac_template         # 模板配置表
mac_addon            # 插件表
```

### 3.2 Go 数据模型映射 (GORM)

```go
// model/vod.go
type Vod struct {
    ID            uint      `gorm:"primaryKey;column:vod_id"`
    TypeID        int       `gorm:"column:type_id"`
    TypeID1       int       `gorm:"column:type_id_1"`
    GroupID       int       `gorm:"column:group_id"`
    VodName       string    `gorm:"column:vod_name"`
    VodSub        string    `gorm:"column:vod_sub"`
    VodEn         string    `gorm:"column:vod_en"`
    VodTime       string    `gorm:"column:vod_time"`
    VodClass      string    `gorm:"column:vod_class"`
    VodTag        string    `gorm:"column:vod_tag"`
    VodPic        string    `gorm:"column:vod_pic"`
    VodPicThumb   string    `gorm:"column:vod_pic_thumb"`
    VodPicSlide   string    `gorm:"column:vod_pic_slide"`
    VodPicScreenshot string `gorm:"column:vod_pic_screenshot"`
    VodActor      string    `gorm:"column:vod_actor"`
    VodDirector   string    `gorm:"column:vod_director"`
    VodWriter     string    `gorm:"column:vod_writer"`
    VodBlurb      string    `gorm:"column:vod_blurb"`
    VodRemarks    string    `gorm:"column:vod_remarks"`
    VodPubdate    string    `gorm:"column:vod_pubdate"`
    VodArea       string    `gorm:"column:vod_area"`
    VodLang       string    `gorm:"column:vod_lang"`
    VodYear       string    `gorm:"column:vod_year"`
    VodVersion    string    `gorm:"column:vod_version"`
    VodState      string    `gorm:"column:vod_state"`
    VodAuthor     string    `gorm:"column:vod_author"`
    VodJumpurl    string    `gorm:"column:vod_jumpurl"`
    VodLetter     string    `gorm:"column:vod_letter"`
    VodColor      string    `gorm:"column:vod_color"`
    VodLock       int       `gorm:"column:vod_lock"`
    VodLevel      int       `gorm:"column:vod_level"`
    VodPoints     int       `gorm:"column:vod_points"`
    VodPointsPlay int       `gorm:"column:vod_points_play"`
    VodPointsDown int       `gorm:"column:vod_points_down"`
    VodHits       int       `gorm:"column:vod_hits"`
    VodHitsDay    int       `gorm:"column:vod_hits_day"`
    VodHitsWeek   int       `gorm:"column:vod_hits_week"`
    VodHitsMonth  int       `gorm:"column:vod_hits_month"`
    VodDuration   string    `gorm:"column:vod_duration"`
    VodUp         int       `gorm:"column:vod_up"`
    VodDown       int       `gorm:"column:vod_down"`
    VodScore      string    `gorm:"column:vod_score"`
    VodScoreAll   int       `gorm:"column:vod_score_all"`
    VodScoreNum   int       `gorm:"column:vod_score_num"`
    VodContent    string    `gorm:"column:vod_content;type:text"`
    VodPlayFrom   string    `gorm:"column:vod_play_from"`
    VodPlayServer string    `gorm:"column:vod_play_server"`
    VodPlayNote   string    `gorm:"column:vod_play_note"`
    VodPlayURL    string    `gorm:"column:vod_play_url;type:text"`
    VodDownFrom   string    `gorm:"column:vod_down_from"`
    VodDownServer string    `gorm:"column:vod_down_server"`
    VodDownNote   string    `gorm:"column:vod_down_note"`
    VodDownURL    string    `gorm:"column:vod_down_url;type:text"`
    VodPlot       int       `gorm:"column:vod_plot"`
    VodPlotName   string    `gorm:"column:vod_plot_name"`
    VodPlotDetail string    `gorm:"column:vod_plot_detail;type:text"`
    VodStatus     int       `gorm:"column:vod_status"`
    CreatedAt     time.Time `gorm:"column:vod_time_add"`
    UpdatedAt     time.Time `gorm:"column:vod_time_hits"`
}

func (Vod) TableName() string { return "mac_vod" }
```

---

## 四、功能模块完整清单与重写方案

### 4.1 前台展示模块

| 功能 | 原实现 | Go重写方案 |
|------|--------|-----------|
| 首页展示 | ThinkPHP模板渲染 | Fiber + html/template 或 go-template |
| 视频分类列表 | MVC控制器+模型 | `handler/vod.go` Type/Show |
| 视频详情页 | MVC控制器+模型 | `handler/vod.go` Detail |
| 视频播放页 | 动态解析播放地址 | `handler/vod.go` Play |
| 视频下载页 | 动态解析下载地址 | `handler/vod.go` Down |
| 视频搜索 | MySQL LIKE / Meilisearch | `handler/vod.go` Search |
| 视频分集剧情 | 剧情列表 | `handler/vod.go` Plot |
| 文章列表/详情/阅读 | MVC | `handler/art.go` |
| 演员列表/详情 | MVC | `handler/actor.go` |
| 角色列表/详情 | MVC | `handler/role.go` |
| 漫画列表/详情/阅读 | MVC | `handler/manga.go` |
| 专题列表/详情 | MVC | `handler/topic.go` |
| 留言板 | MVC + 验证码 | `handler/gbook.go` |
| 站点地图 (sitemap) | 动态XML生成 | `handler/map.go` |
| RSS订阅 | XML生成 | `handler/rss.go` |
| 用户中心 | 注册/登录/收藏/历史 | `handler/user.go` |
| 自定义标签页 | 模板标签 | `handler/label.go` |
| 投票/评分 | AJAX接口 | `handler/vote.go` |
| 弹幕系统 | WebSocket/AJAX | `handler/danmaku.go` |

### 4.2 后台管理模块

| 功能 | 说明 | Go重写方案 |
|------|------|-----------|
| 仪表盘 | 系统概览/统计/快捷操作 | `handler/admin_dashboard.go` |
| 系统设置 | 全局配置(80KB巨文件) | 拆分为 `admin_system_*.go` 多文件 |
| 视频管理 | CRUD/批量操作/排序/审核 | `handler/admin_vod.go` |
| 文章管理 | CRUD/批量操作 | `handler/admin_art.go` |
| 漫画管理 | CRUD/章节管理 | `handler/admin_manga.go` |
| 演员管理 | CRUD/关联视频 | `handler/admin_actor.go` |
| 角色管理 | CRUD/关联视频 | `handler/admin_role.go` |
| 分类管理 | 树形分类CRUD | `handler/admin_type.go` |
| 用户管理 | CRUD/封禁/积分 | `handler/admin_user.go` |
| 用户组管理 | 权限组CRUD | `handler/admin_group.go` |
| 管理员管理 | CRUD/权限 | `handler/admin_admin.go` |
| 管理员审计 | 操作审计日志 | `handler/admin_audit.go` |
| 评论管理 | 审核/删除/黑名单 | `handler/admin_comment.go` |
| 留言管理 | 审核/回复/删除 | `handler/admin_gbook.go` |
| 专题管理 | CRUD/内容关联 | `handler/admin_topic.go` |
| 友情链接管理 | CRUD/排序 | `handler/admin_link.go` |
| 采集管理 | 采集源/采集规则/定时采集 | `handler/admin_collect.go` |
| 采集规则管理 | 规则导入导出/测试 | `handler/admin_cj.go` |
| 数据库管理 | 备份/恢复/优化/SQL执行 | `handler/admin_database.go` |
| 模板管理 | 模板文件编辑/切换 | `handler/admin_template.go` |
| 模板配置 | 模板参数配置 | `handler/admin_tplconfig.go` |
| 静态生成 | 全站/分类/详情页生成 | `handler/admin_make.go` |
| 上传管理 | 文件上传/裁剪/水印 | `handler/admin_upload.go` |
| 附件管理 | 文件浏览/管理 | `handler/admin_annex.go` |
| 图片管理 | 图片CRUD | `handler/admin_images.go` |
| 卡密管理 | 卡密生成/导入/导出 | `handler/admin_card.go` |
| 订单管理 | 订单查看/处理 | `handler/admin_order.go` |
| 提现管理 | 提现审核/打款 | `handler/admin_cash.go` |
| 直播管理 | 直播源CRUD | `handler/admin_live.go` |
| 弹幕管理 | 弹幕审核/清理 | `handler/admin_danmaku.go` |
| 聊天室管理 | 聊天室配置 | `handler/admin_chatroom.go` |
| 域名管理 | 多域名配置 | `handler/admin_domain.go` |
| URL推送 | 百度/神马/搜狗推送 | `handler/admin_urlsend.go` |
| 安全设置 | IP黑名单/访问控制 | `handler/admin_safety.go` |
| 站点管理 | 多站点配置 | `handler/admin_website.go` |
| 插件管理 | 安装/卸载/配置/启用 | `handler/admin_addon.go` |
| 资源中心 | 模板/插件资源市场 | `handler/admin_resourcehub.go` |
| Meilisearch | 搜索引擎配置/同步 | `handler/admin_meilisearch.go` |
| 数据分析 | 访问统计/趋势图表 | `handler/admin_analytics.go` |
| 播放器管理 | 播放器配置/批量 | `handler/admin_player.go` |
| 服务器管理 | 视频服务器配置 | `handler/admin_server.go` |
| 数据替换 | 批量查找替换 | `handler/admin_datareplace.go` |
| 操作日志 | 后台操作记录 | `handler/admin_plog.go` |
| 用户日志 | 用户行为日志 | `handler/admin_ulog.go` |
| 访问统计 | PV/UV/来源统计 | `handler/admin_visit.go` |
| 系统更新 | 版本检查/更新 | `handler/admin_update.go` |
| 签到里程碑 | 签到奖励配置 | `handler/admin_signmilestone.go` |
| 定时任务 | Cron任务管理 | `handler/admin_timming.go` |
| 任务管理 | 异步任务管理 | `handler/admin_task.go` |

### 4.3 API接口模块

| 接口 | 说明 | Go重写方案 |
|------|------|-----------|
| 资源提供接口 | 采集数据源API (XML/JSON) | `handler/api_provide.go` |
| 用户API | 登录/注册/信息/收藏 | `handler/api_user.go` |
| 评论API | 发表/列表 | `handler/api_comment.go` |
| 留言API | 发表/列表 | `handler/api_gbook.go` |
| 弹幕API | 发送/接收 | `handler/api_danmaku.go` |
| 卡密API | 验证/使用 | `handler/api_card.go` |
| 订单/支付API | 创建/回调/查询 | `handler/api_order.go` |

### 4.4 插件系统

| 功能 | 说明 | Go重写方案 |
|------|------|-----------|
| 插件架构 | 钩子机制+事件驱动 | Go Plugin 或 编译时插件 |
| 后台登录背景 | 自定义登录页背景 | 内置可配置模块 |
| AI内容生成 | DeepSeek/OpenAI/Claude/Gemini/GLM/Qwen | `service/ai_content.go` |

### 4.5 模板引擎

| 功能 | 原实现 | Go重写方案 |
|------|--------|-----------|
| 模板标签解析 | 自定义标签引擎 (mac_*标签) | Go template.FuncMap 自定义函数 |
| 条件/循环标签 | `{maccms:if}{/maccms:if}` | Go template `{{if}}` `{{range}}` |
| 分页标签 | 自定义分页函数 | 自定义 Paginator 函数 |
| SEO标签 | title/keywords/description | 模板变量注入 |
| 自定义函数标签 | 用户可定义PHP函数标签 | 可配置的函数标签注册 |

---

## 五、暗桩后门分析与清除方案

### 5.1 已知风险点

#### 5.1.1 ThinkPHP框架自身风险

ThinkPHP 5.0.24 存在多个已知漏洞:
- **远程代码执行 (RCE)**: 框架历史漏洞
- **SQL注入**: ORM层潜在风险
- **反序列化漏洞**: PHP反序列化链

**Go重写方案**: 彻底抛弃ThinkPHP框架，使用Go原生GORM + Fiber，从根源消除PHP框架漏洞。

#### 5.1.2 后门检查清单

| 风险类型 | 检查位置 | Go重写对策 |
|----------|----------|-----------|
| 远程文件包含 | `include`/`require` 外部URL | Go无此机制，所有模块编译时确定 |
| eval/exec后门 | 全局搜索 `eval(`, `exec(`, `system(`, `shell_exec(` | Go不支持运行时代码执行 |
| 加密/混淆代码 | 检查 `ionCube`, `Zend Guard` 加密文件 | Go源码完全开放 |
| 外部回调 | 搜索 `curl_exec`, `file_get_contents` 连接外部地址 | 所有HTTP请求白名单控制 |
| 隐藏管理员 | 数据库检查非预期管理员账户 | 代码审计 + 安全审计日志 |
| 定时外连 | crontab / 系统计划任务 | Go内置定时任务，可审计 |
| Webshell | 搜索 `<?php` + 动态执行模式 | Go无此风险 |
| 后门API | 检查是否有隐藏API端点 | 路由显式注册，无隐藏路由 |
| 配置文件篡改 | 检查 `extra/` 目录下配置 | 所有配置集中管理，校验完整性 |

#### 5.1.3 具体检查项 (源码审计)

```
# 1. 搜索危险函数
eval(, exec(, system(, passthru(, shell_exec(, 
popen(, proc_open(, pcntl_exec(

# 2. 搜索编码/混淆
base64_decode( + eval( 组合
str_rot13( + eval( 组合
gzinflate( + base64_decode( 组合

# 3. 搜索外部连接
curl_init( 非采集模块的调用
file_get_contents( 含外部URL
fsockopen( 非预期的socket连接

# 4. 搜索文件操作后门
move_uploaded_file( 到非预期路径
file_put_contents( 动态内容
chmod( 设置777权限

# 5. 搜索隐藏配置
检查 mac_config 表中是否有非预期配置项
检查 extra/ 目录是否有隐藏PHP文件
检查根目录是否有非预期入口文件
```

### 5.2 Go重写安全策略

```go
// 1. 所有SQL操作使用GORM参数化查询，杜绝SQL注入
db.Where("vod_name LIKE ?", "%"+keyword+"%").Find(&vods)

// 2. 文件上传白名单控制
var allowedExts = map[string]bool{
    ".jpg": true, ".jpeg": true, ".png": true, 
    ".gif": true, ".webp": true, ".mp4": true,
}

// 3. 无eval/exec等动态执行能力
// 4. 所有路由显式注册，无隐藏端点
// 5. 敏感操作审计日志
// 6. CSRF保护
// 7. XSS输出转义
// 8. Rate Limiting
```

---

## 六、Go + Fiber 技术架构设计

### 6.1 技术栈

| 组件 | 选型 | 说明 |
|------|------|------|
| Web框架 | **Fiber v2** | Express风格，高性能 |
| ORM | **GORM v2** | Go最流行的ORM |
| 数据库驱动 | **go-sql-driver/mysql** | MySQL驱动 |
| 模板引擎 | **html/template** + 自定义FuncMap | 兼容原模板标签 |
| 缓存 | **go-redis/redis** + 文件缓存 | 双层缓存 | 详见下方缓存专项 |
| 会话管理 | **Fiber Session** | Cookie/Session |
| 定时任务 | **robfig/cron v3** | Cron定时任务 |
| 日志 | **zap** | 高性能结构化日志 |
| 配置管理 | **viper** | 多格式配置 |
| 搜索引擎 | **Meilisearch Go SDK** | 全文搜索 |
| WebSocket | **Fiber WebSocket** | 弹幕/实时通信 |
| HTTP客户端 | **go-resty/resty** | 采集模块 |
| 图片处理 | **disintegration/imaging** | 缩略图/水印 |
| 验证码 | **dchest/captcha** | 图形验证码 |
| JWT | **golang-jwt/jwt/v5** | API认证 |
| 邮件 | **go-gomail/gomail** | 邮件通知 |
| XML生成 | **encoding/xml** | RSS/Sitemap |
| 压缩 | **klauspost/gzip** | Gzip中间件 |

### 6.2 项目目录结构

```
maccms-go/
├── cmd/
│   └── server/
│       └── main.go              # 入口
├── internal/
│   ├── config/
│   │   ├── config.go            # 配置加载
│   │   └── types.go             # 配置类型
│   ├── database/
│   │   ├── mysql.go             # MySQL连接
│   │   ├── redis.go             # Redis连接
│   │   └── migrate.go           # 数据库迁移
│   ├── middleware/
│   │   ├── auth.go              # 后台认证中间件
│   │   ├── cors.go              # CORS
│   │   ├── csrf.go              # CSRF保护
│   │   ├── gzip.go              # Gzip压缩
│   │   ├── logger.go            # 请求日志
│   │   ├── ratelimit.go         # 限流
│   │   ├── recover.go           # 异常恢复
│   │   └── security.go          # 安全头
│   ├── model/
│   │   ├── vod.go               # 视频模型
│   │   ├── art.go               # 文章模型
│   │   ├── actor.go             # 演员模型
│   │   ├── role.go              # 角色模型
│   │   ├── manga.go             # 漫画模型
│   │   ├── type.go              # 分类模型
│   │   ├── user.go              # 用户模型
│   │   ├── admin.go             # 管理员模型
│   │   ├── comment.go           # 评论模型
│   │   ├── gbook.go             # 留言模型
│   │   ├── topic.go             # 专题模型
│   │   ├── link.go              # 链接模型
│   │   ├── card.go              # 卡密模型
│   │   ├── order.go             # 订单模型
│   │   ├── cash.go              # 提现模型
│   │   ├── live.go              # 直播模型
│   │   ├── danmaku.go           # 弹幕模型
│   │   ├── chatroom.go          # 聊天室模型
│   │   ├── images.go            # 图片模型
│   │   ├── collect.go           # 采集模型
│   │   ├── config.go            # 配置模型
│   │   ├── template.go          # 模板模型
│   │   ├── addon.go             # 插件模型
│   │   ├── log.go               # 日志模型
│   │   └── visit.go             # 访问模型
│   ├── handler/
│   │   ├── frontend/
│   │   │   ├── index.go         # 首页
│   │   │   ├── vod.go           # 视频前台
│   │   │   ├── art.go           # 文章前台
│   │   │   ├── actor.go         # 演员前台
│   │   │   ├── role.go          # 角色前台
│   │   │   ├── manga.go         # 漫画前台
│   │   │   ├── topic.go         # 专题前台
│   │   │   ├── gbook.go         # 留言板
│   │   │   ├── user.go          # 用户中心
│   │   │   ├── map.go           # 站点地图
│   │   │   ├── rss.go           # RSS
│   │   │   └── label.go         # 自定义标签
│   │   ├── admin/
│   │   │   ├── dashboard.go     # 仪表盘
│   │   │   ├── system.go        # 系统设置
│   │   │   ├── vod.go           # 视频管理
│   │   │   ├── art.go           # 文章管理
│   │   │   ├── manga.go         # 漫画管理
│   │   │   ├── actor.go         # 演员管理
│   │   │   ├── role.go          # 角色管理
│   │   │   ├── type.go          # 分类管理
│   │   │   ├── user.go          # 用户管理
│   │   │   ├── group.go         # 用户组管理
│   │   │   ├── admin.go         # 管理员管理
│   │   │   ├── audit.go         # 审计管理
│   │   │   ├── comment.go       # 评论管理
│   │   │   ├── gbook.go         # 留言管理
│   │   │   ├── topic.go         # 专题管理
│   │   │   ├── link.go          # 链接管理
│   │   │   ├── collect.go       # 采集管理
│   │   │   ├── cj.go            # 采集规则
│   │   │   ├── database.go      # 数据库管理
│   │   │   ├── template.go      # 模板管理
│   │   │   ├── tplconfig.go     # 模板配置
│   │   │   ├── make.go          # 静态生成
│   │   │   ├── upload.go        # 上传管理
│   │   │   ├── annex.go         # 附件管理
│   │   │   ├── images.go        # 图片管理
│   │   │   ├── card.go          # 卡密管理
│   │   │   ├── order.go         # 订单管理
│   │   │   ├── cash.go          # 提现管理
│   │   │   ├── live.go          # 直播管理
│   │   │   ├── danmaku.go       # 弹幕管理
│   │   │   ├── chatroom.go      # 聊天室管理
│   │   │   ├── domain.go        # 域名管理
│   │   │   ├── urlsend.go       # URL推送
│   │   │   ├── safety.go        # 安全设置
│   │   │   ├── website.go       # 站点管理
│   │   │   ├── addon.go         # 插件管理
│   │   │   ├── resourcehub.go   # 资源中心
│   │   │   ├── meilisearch.go   # Meilisearch管理
│   │   │   ├── analytics.go     # 数据分析
│   │   │   ├── player.go        # 播放器管理
│   │   │   ├── server.go        # 服务器管理
│   │   │   ├── datareplace.go   # 数据替换
│   │   │   ├── plog.go          # 操作日志
│   │   │   ├── ulog.go          # 用户日志
│   │   │   ├── visit.go         # 访问统计
│   │   │   ├── update.go        # 系统更新
│   │   │   ├── signmilestone.go # 签到里程碑
│   │   │   ├── timming.go       # 定时任务
│   │   │   ├── task.go          # 任务管理
│   │   │   └── assistant.go     # 助手工具
│   │   └── api/
│   │       ├── provide.go       # 资源提供
│   │       ├── user.go          # 用户API
│   │       ├── comment.go       # 评论API
│   │       ├── gbook.go         # 留言API
│   │       ├── danmaku.go       # 弹幕API
│   │       ├── card.go          # 卡密API
│   │       └── order.go         # 订单API
│   ├── service/
│   │   ├── collect.go           # 采集服务
│   │   ├── template.go          # 模板引擎服务
│   │   ├── cache.go             # 缓存服务
│   │   ├── search.go            # 搜索服务
│   │   ├── upload.go            # 上传服务
│   │   ├── email.go             # 邮件服务
│   │   ├── sms.go               # 短信服务
│   │   ├── payment.go           # 支付服务
│   │   ├── captcha.go           # 验证码服务
│   │   ├── ai_content.go        # AI内容生成
│   │   ├── url_push.go          # URL推送服务
│   │   ├── static_gen.go        # 静态生成服务
│   │   └── scheduler.go         # 调度服务
│   ├── template/
│   │   ├── engine.go            # 模板引擎
│   │   ├── funcs.go             # 模板函数 (mac_*标签映射)
│   │   ├── parser.go            # 标签解析器
│   │   └── paginator.go         # 分页器
│   └── router/
│       ├── router.go            # 路由总入口
│       ├── url_engine.go        # 动态URL规则引擎 (核心)
│       ├── frontend.go          # 前台handler分发
│       ├── admin.go             # 后台路由 (固定)
│       └── api.go               # API路由 (固定)
├── web/
│   ├── static/                  # 静态资源
│   │   ├── css/
│   │   ├── js/
│   │   ├── images/
│   │   └── uploads/
│   ├── template/                # 前台模板
│   └── admin/                   # 后台模板
├── addon/                       # 插件目录
├── config/
│   ├── config.yaml              # 主配置
│   └── config.example.yaml      # 示例配置
├── migrations/                  # 数据库迁移文件
├── scripts/                     # 脚本工具
├── go.mod
├── go.sum
├── Makefile
├── Dockerfile
└── README.md
```

### 6.3 动态URL路由系统 (核心设计)

**关键点**: 原苹果CMS的URL规则是**完全可配置的**，管理员可以在后台「系统设置 → URL规则」中自定义每个页面的URL模式。路由不是写死的，而是**数据驱动**的。Go重写必须保持这一架构。

#### 6.3.1 原项目URL规则机制分析

原系统通过 `mac_config` 表存储URL配置，分三层控制:

**第一层: URL模式选择 (`config['view']`)**

每个页面类型可独立选择「动态」或「静态」模式:

| 配置项 | 值=0 (动态) | 值=2 (静态) |
|--------|------------|------------|
| `view[vod_type]` | 走ThinkPHP路由规则 | 走 `path[vod_type]` 模板 |
| `view[vod_detail]` | 走ThinkPHP路由规则 | 走 `path[vod_detail]` 模板 |
| `view[vod_play]` | 走ThinkPHP路由规则 | 走 `path[vod_play]` 模板 (2/3/4多种子模式) |
| `view[vod_down]` | 同上 | 同上 |
| `view[art_type]` | 走ThinkPHP路由规则 | 走 `path[art_type]` 模板 |
| `view[art_detail]` | 走ThinkPHP路由规则 | 走 `path[art_detail]` 模板 |
| `view[topic_index]` | 走路由 | 走 `path[topic_index]` |
| `view[topic_detail]` | 走路由 | 走 `path[topic_detail]` |
| ... | ... | ... |

**第二层: URL路径模板 (`config['path']`)**

静态模式下的URL路径模板，支持占位符:

```
path[vod_type]    = "vodtypehtml/{id}/index"    → /vodtypehtml/1/index.html
path[vod_detail]  = "vodhtml/{id}/index"        → /vodhtml/123/index.html
path[vod_play]    = "vodplayhtml/{id}/index"    → /vodplayhtml/123/index.html
path[art_type]    = "arttypehtml/{id}/index"    → /arttypehtml/1/index.html
path[art_detail]  = "arthtml/{id}/index"        → /arthtml/456/index.html
path[topic_index] = "topic/index"               → /topic/index.html
path[topic_detail]= "topic/{id}/index"          → /topic/1/index.html
path[page_sp]     = "-"                         → 分页连接符
path[suffix]      = "html"                      → 文件后缀
```

占位符: `{id}`, `{en}`, `{page}`, `{type_id}`, `{type_en}`, `{type_pid}`, `{type_pen}`, `{md5}`, `{year}`, `{month}`, `{day}`, `{sid}`, `{nid}`

**第三层: ID编码策略 (`config['rewrite']`)**

控制URL中的ID如何展示:

| 配置项 | 值=0 | 值=1 | 值=2 |
|--------|------|------|------|
| `rewrite[vod_id]` | 原始数字ID | 英文别名(vod_en) | AlphaID编码 |
| `rewrite[type_id]` | 原始数字ID | 英文别名(type_en) | AlphaID编码 |
| `rewrite[art_id]` | 原始数字ID | 英文别名(art_en) | AlphaID编码 |
| `rewrite[manga_id]` | 原始数字ID | 英文别名 | AlphaID编码 |
| `rewrite[encode_len]` | - | - | AlphaID编码长度 |
| `rewrite[encode_key]` | - | - | AlphaID编码密钥 |
| `rewrite[suffix_hide]` | 保留.html后缀 | 隐藏后缀(用/代替) | - |
| `rewrite[route_status]` | 关闭路由 | 开启路由 | - |

**第四层: ThinkPHP动态路由规则 (`route.php`)**

当 `view[x]=0` 时，走ThinkPHP的路由规则，例如:
```
'vodtype/<id>-<page?>' → 'vod/type'
'voddetail/<id>'       → 'vod/detail'
'vodplay/<id>-<sid>-<nid>' → 'vod/play'
```

#### 6.3.2 Go重写: 动态URL路由器设计

```go
// router/url_engine.go — 数据驱动的URL引擎

package router

import (
    "regexp"
    "strings"
    "sync"
    "github.com/gofiber/fiber/v2"
)

// ============================================================
// URLRuleEngine: 核心URL规则引擎
// 所有规则从数据库 mac_config 表加载，管理员可在后台修改
// ============================================================

type URLRuleEngine struct {
    mu          sync.RWMutex
    config      *URLConfig       // 从DB加载的URL配置
    viewRules   map[string]*CompiledRule // 编译后的规则缓存
    staticRules map[string]*CompiledRule
    reverseMap  map[string]string // 反向映射: handler名 → 模型名
}

type URLConfig struct {
    // config['view'] — 每个页面的URL模式 (0=动态, 2=静态)
    View map[string]int `json:"view"`
    // config['path'] — 静态模式下的URL路径模板
    Path map[string]string `json:"path"`
    // config['rewrite'] — ID编码策略
    Rewrite RewriteConfig `json:"rewrite"`
}

type RewriteConfig struct {
    VodID       int    `json:"vod_id"`       // 0=数字, 1=英文, 2=AlphaID
    TypeID      int    `json:"type_id"`
    ArtID       int    `json:"art_id"`
    MangaID     int    `json:"manga_id"`
    ActorID     int    `json:"actor_id"`
    RoleID      int    `json:"role_id"`
    EncodeLen   int    `json:"encode_len"`
    EncodeKey   string `json:"encode_key"`
    SuffixHide  int    `json:"suffix_hide"`  // 0=保留后缀, 1=隐藏后缀
    RouteStatus int    `json:"route_status"` // 0=关闭路由, 1=开启路由
    Status      int    `json:"status"`       // 伪静态总开关
}

// 动态路由规则定义 (对应 route.php)
type DynamicRoute struct {
    Pattern   string            // URL模式: "vodtype/<id>-<page?>"
    Target    string            // 目标: "vod/type"
    Params    map[string]string // 参数名 → 正则
}

// 编译后的规则
type CompiledRule struct {
    Regexp    *regexp.Regexp
    Target    string
    ParamKeys []string
    IsStatic  bool
    Template  string // 静态模板: "vodtypehtml/{id}/index"
}

// ============================================================
// 初始化: 从数据库加载配置并编译规则
// ============================================================

func NewURLRuleEngine(db *gorm.DB) *URLRuleEngine {
    engine := &URLRuleEngine{
        viewRules:   make(map[string]*CompiledRule),
        staticRules: make(map[string]*CompiledRule),
        reverseMap:  make(map[string]string),
    }
    engine.LoadFromDB(db)
    return engine
}

func (e *URLRuleEngine) LoadFromDB(db *gorm.DB) {
    // 从 mac_config 表加载 URL 相关配置
    var configs []MacConfig
    db.Where("`type` = ?", "maccms").Find(&configs)
    
    urlConfig := &URLConfig{
        View: make(map[string]int),
        Path: make(map[string]string),
    }
    
    for _, c := range configs {
        switch c.Name {
        case "view":
            json.Unmarshal([]byte(c.Value), &urlConfig.View)
        case "path":
            json.Unmarshal([]byte(c.Value), &urlConfig.Path)
        case "rewrite":
            json.Unmarshal([]byte(c.Value), &urlConfig.Rewrite)
        }
    }
    
    e.config = urlConfig
    e.compileRules()
}

// compileRules: 编译所有URL规则
func (e *URLRuleEngine) compileRules() {
    // 1. 编译动态路由规则 (route.php 中定义的)
    dynamicRoutes := e.getDefaultDynamicRoutes()
    for _, route := range dynamicRoutes {
        compiled := e.compileDynamicRoute(route)
        e.viewRules[route.Target] = compiled
    }
    
    // 2. 编译静态路径模板
    for key, template := range e.config.Path {
        compiled := e.compileStaticTemplate(key, template)
        e.staticRules[key] = compiled
    }
}

// getDefaultDynamicRoutes: 默认动态路由 (对应 route.php)
func (e *URLRuleEngine) getDefaultDynamicRoutes() []DynamicRoute {
    return []DynamicRoute{
        // 首页
        {Pattern: "index-<page>", Target: "index/index",
            Params: map[string]string{"page": `\d+`}},
        
        // 视频
        {Pattern: "vodtype/<id>-<page>", Target: "vod/type",
            Params: map[string]string{"id": `[\s\S]*?`, "page": `\d+`}},
        {Pattern: "voddetail/<id>", Target: "vod/detail",
            Params: map[string]string{"id": `[\s\S]*?`}},
        {Pattern: "vodplay/<id>-<sid>-<nid>", Target: "vod/play",
            Params: map[string]string{"id": `[\s\S]*?`, "sid": `\d+`, "nid": `\d+`}},
        {Pattern: "voddown/<id>-<sid>-<nid>", Target: "vod/down",
            Params: map[string]string{"id": `[\s\S]*?`, "sid": `\d+`, "nid": `\d+`}},
        {Pattern: "vodshow/<id>-<area?>-<by?>-<class?>-<lang?>-<letter?>-<level?>-<order?>-<page?>-<state?>-<tag?>-<year?>", Target: "vod/show",
            Params: map[string]string{"id": `[\s\S]*?`, "page": `\d+`, "level": `\d+`}},
        {Pattern: "vodsearch/<wd?>-<actor?>-<area?>-<by?>-<class?>-<director?>-<lang?>-<letter?>-<level?>-<order?>-<page?>-<state?>-<tag?>-<year?>", Target: "vod/search",
            Params: map[string]string{"wd": `[\s\S]*`, "page": `\d+`}},
        {Pattern: "vodplot/<id>-<page?>", Target: "vod/plot",
            Params: map[string]string{"id": `[\s\S]*?`, "page": `\d+`}},
        
        // 文章
        {Pattern: "arttype/<id>-<page?>", Target: "art/type",
            Params: map[string]string{"id": `[\s\S]*?`, "page": `\d+`}},
        {Pattern: "artdetail/<id>-<page?>", Target: "art/detail",
            Params: map[string]string{"id": `[\s\S]*?`, "page": `\d+`}},
        {Pattern: "artread/<id>-<page?>", Target: "art/read",
            Params: map[string]string{"id": `[\s\S]*?`, "page": `\d+`}},
        {Pattern: "artshow/<id>-<by?>-<class?>-<level?>-<letter?>-<order?>-<page?>-<tag?>", Target: "art/show",
            Params: map[string]string{"id": `[\s\S]*?`, "page": `\d+`}},
        {Pattern: "artsearch/<wd?>-<by?>-<class?>-<level?>-<letter?>-<order?>-<page?>-<tag?>", Target: "art/search",
            Params: map[string]string{"wd": `[\s\S]*`, "page": `\d+`}},
        
        // 漫画
        {Pattern: "mangatype/<id>-<page?>", Target: "manga/type",
            Params: map[string]string{"id": `[\s\S]*?`, "page": `\d+`}},
        {Pattern: "mangadetail/<id>", Target: "manga/detail",
            Params: map[string]string{"id": `[\s\S]*?`}},
        {Pattern: "mangaplay/<id>-<sid>-<nid>", Target: "manga/play",
            Params: map[string]string{"id": `[\s\S]*?`, "sid": `\d+`, "nid": `\d+`}},
        {Pattern: "mangashow/<id>-<area?>-<by?>-<class?>-<lang?>-<letter?>-<level?>-<order?>-<page?>-<state?>-<tag?>-<year?>", Target: "manga/show",
            Params: map[string]string{"id": `[\s\S]*?`, "page": `\d+`}},
        {Pattern: "mangasearch/<wd?>-<actor?>-<area?>-<by?>-<class?>-<director?>-<lang?>-<letter?>-<level?>-<order?>-<page?>-<state?>-<tag?>-<year?>", Target: "manga/search",
            Params: map[string]string{"wd": `[\s\S]*`, "page": `\d+`}},
        
        // 演员
        {Pattern: "actordetail/<id>", Target: "actor/detail",
            Params: map[string]string{"id": `[\s\S]*?`}},
        {Pattern: "actorshow/<area?>-<blood?>-<by?>-<letter?>-<level?>-<order?>-<page?>-<sex?>-<starsign?>", Target: "actor/show",
            Params: map[string]string{"page": `\d+`, "level": `\d+`}},
        
        // 角色
        {Pattern: "roledetail/<id>", Target: "role/detail",
            Params: map[string]string{"id": `[\s\S]*?`}},
        {Pattern: "roleshow/<by?>-<letter?>-<level?>-<order?>-<page?>-<rid?>", Target: "role/show",
            Params: map[string]string{"page": `\d+`, "level": `\d+`}},
        
        // 专题
        {Pattern: "topicdetail-<id>", Target: "topic/detail",
            Params: map[string]string{"id": `[\s\S]*?`}},
        
        // 其他
        {Pattern: "gbook-<page?>", Target: "gbook/index",
            Params: map[string]string{"page": `\d+`}},
        {Pattern: "label-<file>", Target: "label/index",
            Params: map[string]string{"file": `[\s\S]*?`}},
    }
}

// ============================================================
// URL解析: 请求 → Handler
// ============================================================

// Resolve: 核心解析方法，根据请求路径找到对应handler
func (e *URLRuleEngine) Resolve(c *fiber.Ctx) (string, map[string]string, error) {
    path := c.Path()
    path = strings.TrimPrefix(path, "/")
    path = strings.TrimSuffix(path, "/")
    
    // 去掉后缀 (.html, .htm, .shtml)
    suffix := e.config.Path["suffix"]
    if suffix == "" {
        suffix = "html"
    }
    if strings.HasSuffix(path, "."+suffix) {
        path = strings.TrimSuffix(path, "."+suffix)
    }
    
    e.mu.RLock()
    defer e.mu.RUnlock()
    
    // 1. 先尝试静态路径模板匹配 (config['view']=2 的页面)
    if target, params := e.matchStaticPath(path); target != "" {
        return target, params, nil
    }
    
    // 2. 再尝试动态路由匹配 (config['view']=0 的页面)
    if target, params := e.matchDynamicRoute(path); target != "" {
        return target, params, nil
    }
    
    return "", nil, fiber.ErrNotFound
}

// matchStaticPath: 匹配静态路径模板
// 例如: path[vod_detail] = "vodhtml/{id}/index"
//       请求: /vodhtml/123/index.html
//       匹配: target="vod/detail", params={"id":"123"}
func (e *URLRuleEngine) matchStaticPath(path string) (string, map[string]string) {
    for key, rule := range e.staticRules {
        if matches := rule.Regexp.FindStringSubmatch(path); matches != nil {
            params := make(map[string]string)
            for i, name := range rule.ParamKeys {
                if i+1 < len(matches) {
                    params[name] = matches[i+1]
                }
            }
            // 将path key映射到handler target
            target := e.pathKeyToTarget(key)
            return target, params
        }
    }
    return "", nil
}

// matchDynamicRoute: 匹配动态路由规则
func (e *URLRuleEngine) matchDynamicRoute(path string) (string, map[string]string) {
    for _, rule := range e.viewRules {
        if matches := rule.Regexp.FindStringSubmatch(path); matches != nil {
            params := make(map[string]string)
            for i, name := range rule.ParamKeys {
                if i+1 < len(matches) {
                    params[name] = matches[i+1]
                }
            }
            return rule.Target, params
        }
    }
    return "", nil
}

// ============================================================
// URL生成: Handler → URL (模板函数使用)
// ============================================================

// GenerateURL: 根据handler和参数生成URL (对应原 mac_url 函数)
func (e *URLRuleEngine) GenerateURL(model string, info map[string]interface{}, params map[string]interface{}) string {
    e.mu.RLock()
    defer e.mu.RUnlock()
    
    // 检查该页面是动态模式还是静态模式
    viewKey := e.modelToViewKey(model)
    viewMode := e.config.View[viewKey]
    
    if viewMode == 2 {
        // 静态模式: 使用 path 模板
        return e.generateStaticURL(model, info, params)
    }
    
    // 动态模式: 使用路由规则
    return e.generateDynamicURL(model, info, params)
}

// generateStaticURL: 静态模式URL生成
func (e *URLRuleEngine) generateStaticURL(model string, info map[string]interface{}, params map[string]interface{}) string {
    pathKey := e.modelToPathKey(model)
    template, ok := e.config.Path[pathKey]
    if !ok {
        return ""
    }
    
    // 解析ID编码策略
    idField := e.getIDField(model)
    idValue := e.encodeID(model, info)
    
    // 替换占位符
    replacements := map[string]string{
        "{id}":      fmt.Sprintf("%v", idValue),
        "{en}":      fmt.Sprintf("%v", info[e.getEnField(model)]),
        "{page}":    fmt.Sprintf("%v", params["page"]),
        "{type_id}": fmt.Sprintf("%v", info["type_id"]),
        "{type_en}": fmt.Sprintf("%v", info["type_en"]),
        "{type_pid}": fmt.Sprintf("%v", info["type_1_id"]),
        "{type_pen}": fmt.Sprintf("%v", info["type_1_en"]),
        "{md5}":     md5Encode(fmt.Sprintf("%v", info[idField])),
        "{year}":    extractYear(info),
        "{month}":   extractMonth(info),
        "{day}":     extractDay(info),
        "{sid}":     fmt.Sprintf("%v", params["sid"]),
        "{nid}":     fmt.Sprintf("%v", params["nid"]),
    }
    
    path := template
    for placeholder, value := range replacements {
        path = strings.ReplaceAll(path, placeholder, value)
    }
    
    // 添加分页
    pageSp := e.config.Path["page_sp"]
    if pageSp == "" {
        pageSp = "-"
    }
    page := fmt.Sprintf("%v", params["page"])
    if page != "" && page != "1" {
        path += pageSp + page
    }
    
    // 添加后缀
    suffixHide := e.config.Rewrite.SuffixHide
    suffix := e.config.Path["suffix"]
    if suffix == "" {
        suffix = "html"
    }
    if suffixHide == 1 {
        path += "/"
    } else {
        path += "." + suffix
    }
    
    return "/" + path
}

// generateDynamicURL: 动态模式URL生成
func (e *URLRuleEngine) generateDynamicURL(model string, info map[string]interface{}, params map[string]interface{}) string {
    // 查找匹配的动态路由规则
    for _, route := range e.viewRules {
        if route.Target == model {
            // 使用路由规则生成URL
            idValue := e.encodeID(model, info)
            url := route.Pattern
            
            // 替换路由参数
            url = strings.ReplaceAll(url, "<id>", fmt.Sprintf("%v", idValue))
            url = strings.ReplaceAll(url, "<sid>", fmt.Sprintf("%v", params["sid"]))
            url = strings.ReplaceAll(url, "<nid>", fmt.Sprintf("%v", params["nid"]))
            url = strings.ReplaceAll(url, "<page>", fmt.Sprintf("%v", params["page"]))
            // ... 其他参数
            
            // 清理可选参数 (? 后缀)
            url = e.cleanOptionalParams(url)
            
            // 添加后缀
            suffix := e.config.Path["suffix"]
            if suffix == "" {
                suffix = "html"
            }
            if e.config.Rewrite.SuffixHide != 1 {
                url += "." + suffix
            }
            
            return "/" + url
        }
    }
    return ""
}

// ============================================================
// ID编码策略 (对应原 config['rewrite'] 配置)
// ============================================================

// encodeID: 根据配置编码ID
func (e *URLRuleEngine) encodeID(model string, info map[string]interface{}) interface{} {
    rewriteConfig := e.config.Rewrite
    idField := e.getIDField(model)
    enField := e.getEnField(model)
    
    var strategy int
    switch {
    case strings.HasPrefix(model, "vod"):
        strategy = rewriteConfig.VodID
    case strings.HasPrefix(model, "art"):
        strategy = rewriteConfig.ArtID
    case strings.HasPrefix(model, "manga"):
        strategy = rewriteConfig.MangaID
    case strings.HasPrefix(model, "type"):
        strategy = rewriteConfig.TypeID
    default:
        strategy = 0
    }
    
    switch strategy {
    case 1:
        // 英文别名
        if en, ok := info[enField]; ok && en != "" {
            return en
        }
        return info[idField]
    case 2:
        // AlphaID编码
        if id, ok := info[idField]; ok {
            return AlphaIDEncode(id, rewriteConfig.EncodeLen, rewriteConfig.EncodeKey)
        }
        return info[idField]
    default:
        // 原始数字ID
        return info[idField]
    }
}

// ============================================================
// 后台URL规则配置管理
// ============================================================

// AdminURLConfig: 后台URL规则配置页面的处理
// 对应原 System.php 中保存 URL 配置的逻辑
type AdminURLConfig struct {
    db    *gorm.DB
    engine *URLRuleEngine
}

// Save: 保存URL规则配置 (对应原后台 form 提交)
func (a *AdminURLConfig) Save(c *fiber.Ctx) error {
    // 解析表单: view[vod_type], path[vod_type], rewrite[vod_id] 等
    viewConfig := make(map[string]int)
    pathConfig := make(map[string]string)
    rewriteConfig := RewriteConfig{}
    
    // 解析 view[] 配置
    for _, key := range []string{
        "index", "map", "search", "rss", "label",
        "vod_type", "vod_show", "vod_detail", "vod_play", "vod_down", "vod_role", "vod_plot",
        "art_type", "art_show", "art_detail", "art_read",
        "manga_type", "manga_show", "manga_detail", "manga_play", "manga_down",
        "actor_index", "actor_detail", "actor_show",
        "role_index", "role_detail", "role_show",
        "topic_index", "topic_detail",
        "gbook", "website_index", "website_detail", "website_show",
    } {
        if v := c.FormValue("view[" + key + "]"); v != "" {
            viewConfig[key], _ = strconv.Atoi(v)
        }
    }
    
    // 解析 path[] 配置
    for _, key := range []string{
        "vod_type", "vod_detail", "vod_play", "vod_down", "vod_role", "vod_plot",
        "art_type", "art_detail", "art_read",
        "manga_type", "manga_detail", "manga_play", "manga_down",
        "topic_index", "topic_detail",
        "page_sp", "suffix",
    } {
        if v := c.FormValue("path[" + key + "]"); v != "" {
            pathConfig[key] = v
        }
    }
    
    // 解析 rewrite[] 配置
    rewriteConfig.VodID, _ = strconv.Atoi(c.FormValue("rewrite[vod_id]"))
    rewriteConfig.TypeID, _ = strconv.Atoi(c.FormValue("rewrite[type_id]"))
    rewriteConfig.ArtID, _ = strconv.Atoi(c.FormValue("rewrite[art_id]"))
    rewriteConfig.MangaID, _ = strconv.Atoi(c.FormValue("rewrite[manga_id]"))
    rewriteConfig.EncodeLen, _ = strconv.Atoi(c.FormValue("rewrite[encode_len]"))
    rewriteConfig.EncodeKey = c.FormValue("rewrite[encode_key]")
    rewriteConfig.SuffixHide, _ = strconv.Atoi(c.FormValue("rewrite[suffix_hide]"))
    rewriteConfig.RouteStatus, _ = strconv.Atoi(c.FormValue("rewrite[route_status]"))
    rewriteConfig.Status, _ = strconv.Atoi(c.FormValue("rewrite[status]"))
    
    // 保存到数据库 mac_config 表
    a.saveConfig("view", viewConfig)
    a.saveConfig("path", pathConfig)
    a.saveConfig("rewrite", rewriteConfig)
    
    // 热更新URL规则引擎 (无需重启)
    a.engine.LoadFromDB(a.db)
    
    return c.JSON(fiber.Map{"code": 1, "msg": "保存成功"})
}

// ============================================================
// 辅助函数
// ============================================================

func (e *URLRuleEngine) modelToViewKey(model string) string {
    // "vod/type" → "vod_type"
    // "vod/detail" → "vod_detail"
    // "vod/play" → "vod_play"
    return strings.ReplaceAll(model, "/", "_")
}

func (e *URLRuleEngine) modelToPathKey(model string) string {
    return strings.ReplaceAll(model, "/", "_")
}

func (e *URLRuleEngine) pathKeyToTarget(key string) string {
    // "vod_type" → "vod/type"
    // "vod_detail" → "vod/detail"
    parts := strings.SplitN(key, "_", 2)
    if len(parts) == 2 {
        return parts[0] + "/" + parts[1]
    }
    return key
}

func (e *URLRuleEngine) getIDField(model string) string {
    switch {
    case strings.HasPrefix(model, "vod"): return "vod_id"
    case strings.HasPrefix(model, "art"): return "art_id"
    case strings.HasPrefix(model, "manga"): return "manga_id"
    case strings.HasPrefix(model, "actor"): return "actor_id"
    case strings.HasPrefix(model, "role"): return "role_id"
    case strings.HasPrefix(model, "topic"): return "topic_id"
    default: return "id"
    }
}

func (e *URLRuleEngine) getEnField(model string) string {
    switch {
    case strings.HasPrefix(model, "vod"): return "vod_en"
    case strings.HasPrefix(model, "art"): return "art_en"
    case strings.HasPrefix(model, "manga"): return "manga_en"
    case strings.HasPrefix(model, "actor"): return "actor_en"
    case strings.HasPrefix(model, "role"): return "role_en"
    case strings.HasPrefix(model, "topic"): return "topic_en"
    default: return "en"
    }
}

// AlphaIDEncode: 对应原 mac_alphaID 函数
// 将数字ID编码为短字符串，用于URL美化
func AlphaIDEncode(id interface{}, length int, key string) string {
    // 实现原 mac_alphaID 的编码逻辑
    // ...
    return fmt.Sprintf("%v", id)
}
```

#### 6.3.3 Fiber路由注册: 统一Catch-All处理器

由于URL规则是动态的，不能用静态路由注册。使用Fiber的 `app.Add()` + Catch-All 模式:

```go
// router/router.go

func SetupRoutes(app *fiber.App, engine *URLRuleEngine, handlers *HandlerRegistry) {
    
    // ==========================================
    // 后台路由 (固定，不走URL规则引擎)
    // ==========================================
    admin := app.Group("/admin")
    admin.Post("/login", handlers.Admin.Login)
    admin.Use(handlers.Middleware.AdminAuth) // 后台认证中间件
    admin.Get("/dashboard", handlers.Admin.Dashboard)
    admin.Get("/system/*", handlers.Admin.System)
    admin.Get("/vod/*", handlers.Admin.Vod)
    admin.Get("/art/*", handlers.Admin.Art)
    admin.Get("/manga/*", handlers.Admin.Manga)
    // ... 所有后台路由 (固定路径)
    
    // 后台保存URL规则的接口
    admin.Post("/system/configurl/save", handlers.Admin.SaveURLConfig)
    
    // ==========================================
    // API路由 (固定)
    // ==========================================
    api := app.Group("/api")
    api.Get("/provide/:ac", handlers.API.Provide)
    api.Post("/user/login", handlers.API.UserLogin)
    api.Post("/user/register", handlers.API.UserRegister)
    // ... 所有API路由
    
    // ==========================================
    // 前台路由: 动态URL规则引擎
    // ==========================================
    
    // 使用 Fiber 的 Add 方法注册所有HTTP方法的Catch-All
    app.Add("GET", "/*", func(c *fiber.Ctx) error {
        // 通过URL规则引擎解析请求
        target, params, err := engine.Resolve(c)
        if err != nil {
            return c.Status(404).SendString("Page not found")
        }
        
        // 将解析结果存入Locals，供handler使用
        c.Locals("target", target)
        c.Locals("params", params)
        
        // 分发到对应handler
        return dispatchToHandler(c, handlers, target, params)
    })
}

// dispatchToHandler: 根据target分发到具体handler
func dispatchToHandler(c *fiber.Ctx, handlers *HandlerRegistry, target string, params map[string]string) error {
    switch target {
    case "index/index":
        return handlers.Frontend.Index(c, params)
    case "vod/type":
        return handlers.Frontend.VodType(c, params)
    case "vod/detail":
        return handlers.Frontend.VodDetail(c, params)
    case "vod/play":
        return handlers.Frontend.VodPlay(c, params)
    case "vod/down":
        return handlers.Frontend.VodDown(c, params)
    case "vod/show":
        return handlers.Frontend.VodShow(c, params)
    case "vod/search":
        return handlers.Frontend.VodSearch(c, params)
    case "vod/plot":
        return handlers.Frontend.VodPlot(c, params)
    case "art/type":
        return handlers.Frontend.ArtType(c, params)
    case "art/detail":
        return handlers.Frontend.ArtDetail(c, params)
    case "art/read":
        return handlers.Frontend.ArtRead(c, params)
    case "art/show":
        return handlers.Frontend.ArtShow(c, params)
    case "art/search":
        return handlers.Frontend.ArtSearch(c, params)
    case "manga/type":
        return handlers.Frontend.MangaType(c, params)
    case "manga/detail":
        return handlers.Frontend.MangaDetail(c, params)
    case "manga/play":
        return handlers.Frontend.MangaPlay(c, params)
    case "manga/show":
        return handlers.Frontend.MangaShow(c, params)
    case "manga/search":
        return handlers.Frontend.MangaSearch(c, params)
    case "actor/index":
        return handlers.Frontend.ActorIndex(c, params)
    case "actor/detail":
        return handlers.Frontend.ActorDetail(c, params)
    case "actor/show":
        return handlers.Frontend.ActorShow(c, params)
    case "role/index":
        return handlers.Frontend.RoleIndex(c, params)
    case "role/detail":
        return handlers.Frontend.RoleDetail(c, params)
    case "role/show":
        return handlers.Frontend.RoleShow(c, params)
    case "topic/index":
        return handlers.Frontend.TopicIndex(c, params)
    case "topic/detail":
        return handlers.Frontend.TopicDetail(c, params)
    case "gbook/index":
        return handlers.Frontend.GbookIndex(c, params)
    case "map/index":
        return handlers.Frontend.MapIndex(c, params)
    case "rss/index":
        return handlers.Frontend.RssIndex(c, params)
    case "label/index":
        return handlers.Frontend.LabelIndex(c, params)
    case "user/login", "user/register", "user/index", "user/fav", "user/history":
        return handlers.Frontend.UserAction(c, params)
    default:
        return c.Status(404).SendString("Page not found")
    }
}
```

#### 6.3.4 模板函数: URL生成 (对应原 mac_url_* 系列)

```go
// template/funcs.go — URL生成模板函数

func GetTemplateFuncs(engine *URLRuleEngine) template.FuncMap {
    return template.FuncMap{
        // 核心URL生成函数 (对应原 mac_url)
        "mac_url": func(model string, param map[string]interface{}, info map[string]interface{}) string {
            return engine.GenerateURL(model, info, param)
        },
        
        // 便捷函数 (对应原 mac_url_vod_detail 等)
        "mac_url_vod_detail": func(info map[string]interface{}) string {
            return engine.GenerateURL("vod/detail", info, nil)
        },
        "mac_url_vod_play": func(info map[string]interface{}, sid, nid int) string {
            return engine.GenerateURL("vod/play", info, map[string]interface{}{"sid": sid, "nid": nid})
        },
        "mac_url_vod_down": func(info map[string]interface{}, sid, nid int) string {
            return engine.GenerateURL("vod/down", info, map[string]interface{}{"sid": sid, "nid": nid})
        },
        "mac_url_vod_type": func(info map[string]interface{}, page int) string {
            return engine.GenerateURL("vod/type", info, map[string]interface{}{"page": page})
        },
        "mac_url_vod_search": func(param map[string]interface{}) string {
            return engine.GenerateURL("vod/search", nil, param)
        },
        "mac_url_art_detail": func(info map[string]interface{}) string {
            return engine.GenerateURL("art/detail", info, nil)
        },
        "mac_url_art_read": func(info map[string]interface{}, page int) string {
            return engine.GenerateURL("art/read", info, map[string]interface{}{"page": page})
        },
        "mac_url_art_type": func(info map[string]interface{}, page int) string {
            return engine.GenerateURL("art/type", info, map[string]interface{}{"page": page})
        },
        "mac_url_manga_detail": func(info map[string]interface{}) string {
            return engine.GenerateURL("manga/detail", info, nil)
        },
        "mac_url_manga_play": func(info map[string]interface{}, sid, nid int) string {
            return engine.GenerateURL("manga/play", info, map[string]interface{}{"sid": sid, "nid": nid})
        },
        "mac_url_actor_detail": func(info map[string]interface{}) string {
            return engine.GenerateURL("actor/detail", info, nil)
        },
        "mac_url_role_detail": func(info map[string]interface{}) string {
            return engine.GenerateURL("role/detail", info, nil)
        },
        "mac_url_topic_detail": func(info map[string]interface{}) string {
            return engine.GenerateURL("topic/detail", info, nil)
        },
        "mac_url_type": func(info map[string]interface{}, param map[string]interface{}, flag string) string {
            return engine.GenerateURL(flag+"/type", info, param)
        },
        "mac_url_search": func(param map[string]interface{}, flag string) string {
            return engine.GenerateURL(flag+"/search", nil, param)
        },
        "mac_url_index": func(page int) string {
            return engine.GenerateURL("index/index", nil, map[string]interface{}{"page": page})
        },
        "mac_url_page": func(url string, num int) string {
            // 分页URL处理
            return strings.ReplaceAll(url, "PAGELINK", fmt.Sprintf("%d", num))
        },
        
        // ... 其他模板函数
    }
}
```

#### 6.3.5 配置热更新 (无需重启)

```go
// 当管理员在后台保存URL规则时，热更新引擎
func (a *AdminURLConfig) Save(c *fiber.Ctx) error {
    // ... 保存到数据库 ...
    
    // 热更新URL规则引擎 (线程安全)
    a.engine.LoadFromDB(a.db)
    
    // 清除模板缓存 (URL函数需要重新编译)
    a.templateEngine.ClearCache()
    
    return c.JSON(fiber.Map{"code": 1, "msg": "保存成功，URL规则已生效"})
}
```

### 6.3.6 原 route.php 完整映射表 (参考)

以下是原 `route.php` 中所有默认动态路由规则，Go重写时需完整覆盖:

| 路由模式 | 目标Handler | 说明 |
|----------|-------------|------|
| `sitehome` | `index/home` | 站点首页 |
| `index-<page?>` | `index/index` | 首页分页 |
| `map` | `map/index` | 站点地图 |
| `rss` | `rss/index` | RSS |
| `gbook-<page?>` | `gbook/index` | 留言板 |
| `topic-<page?>` | `topic/index` | 专题列表 |
| `topicdetail-<id>` | `topic/detail` | 专题详情 |
| `actor-<page?>` | `actor/index` | 演员列表 |
| `actordetail-<id>` | `actor/detail` | 演员详情 |
| `actorshow/<area?>-<blood?>-<by?>-<letter?>-<level?>-<order?>-<page?>-<sex?>-<starsign?>` | `actor/show` | 演员筛选 |
| `role-<page?>` | `role/index` | 角色列表 |
| `roledetail-<id>` | `role/detail` | 角色详情 |
| `roleshow/<by?>-<letter?>-<level?>-<order?>-<page?>-<rid?>` | `role/show` | 角色筛选 |
| `vodtype/<id>-<page?>` | `vod/type` | 视频分类 |
| `voddetail/<id>` | `vod/detail` | 视频详情 |
| `vodrss-<id>` | `vod/rss` | 视频RSS |
| `vodplay/<id>-<sid>-<nid>` | `vod/play` | 视频播放 |
| `voddown/<id>-<sid>-<nid>` | `vod/down` | 视频下载 |
| `vodshow/<id>-<area?>-<by?>-<class?>-<lang?>-<letter?>-<level?>-<order?>-<page?>-<state?>-<tag?>-<year?>` | `vod/show` | 视频筛选 |
| `vodsearch/<wd?>-<actor?>-<area?>-<by?>-<class?>-<director?>-<lang?>-<letter?>-<level?>-<order?>-<page?>-<state?>-<tag?>-<year?>` | `vod/search` | 视频搜索 |
| `vodplot/<id>-<page?>` | `vod/plot` | 视频剧情 |
| `arttype/<id>-<page?>` | `art/type` | 文章分类 |
| `artdetail/<id>-<page?>` | `art/detail` | 文章详情 |
| `artread/<id>-<page?>` | `art/read` | 文章阅读 |
| `artshow/<id>-<by?>-<class?>-<level?>-<letter?>-<order?>-<page?>-<tag?>` | `art/show` | 文章筛选 |
| `artsearch/<wd?>-<by?>-<class?>-<level?>-<letter?>-<order?>-<page?>-<tag?>` | `art/search` | 文章搜索 |
| `artrss-<id>-<page>` | `art/rss` | 文章RSS |
| `manga-<page?>` | `manga/index` | 漫画列表 |
| `mangatype/<id>-<page?>` | `manga/type` | 漫画分类 |
| `mangadetail-<id>` | `manga/detail` | 漫画详情 |
| `mangaplay/<id>-<sid>-<nid>` | `manga/play` | 漫画阅读 |
| `mangadown/<id>-<sid>-<nid>` | `manga/down` | 漫画下载 |
| `mangashow/<id>-<area?>-<by?>-<class?>-<lang?>-<letter?>-<level?>-<order?>-<page?>-<state?>-<tag?>-<year?>` | `manga/show` | 漫画筛选 |
| `mangasearch/<wd?>-<actor?>-<area?>-<by?>-<class?>-<director?>-<lang?>-<letter?>-<level?>-<order?>-<page?>-<state?>-<tag?>-<year?>` | `manga/search` | 漫画搜索 |
| `label-<file>` | `label/index` | 自定义标签页 |
| `publish-<id>` | `index/publish_group` | 发布组 |
| `website-<page?>` | `website/index` | 站点列表 |
| `websitedetail-<id>` | `website/detail` | 站点详情 |

### 6.4 核心中间件链

```go
// middleware链 - 按顺序执行
func SetupMiddleware(app *fiber.App) {
    app.Use(recover.New())           // 异常恢复
    app.Use(logger.New())            // 请求日志
    app.Use(cors.New())              // CORS
    app.Use(csrf.New())              // CSRF保护
    app.Use(compress.New())          // Gzip压缩
    app.Use(limiter.New(limiter.Config{
        Max:        100,
        Expiration: 60 * time.Second,
    }))                              // 限流
    app.Use(security.New())          // 安全头
    app.Use(favicon.New())           // Favicon
    app.Use(static.New("/static", "./web/static")) // 静态文件
}

// 后台认证中间件
func AdminAuth() fiber.Handler {
    return func(c *fiber.Ctx) error {
        session := sessions.Get(c)
        adminID := session.Get("admin_id")
        if adminID == nil {
            return c.Redirect("/admin/login")
        }
        // 检查权限
        return c.Next()
    }
}
```

### 6.5 缓存架构详细设计

#### 6.5.1 原项目缓存机制深度分析

原苹果CMS的缓存分为 **5 个层次**，全部通过 `config['app']` 中的配置控制：

| 配置项 | 默认值 | 说明 |
|--------|--------|------|
| `cache_type` | `file` | 缓存驱动: file / redis / memcache / memcached |
| `cache_host` | `127.0.0.1` | 缓存服务器地址 |
| `cache_port` | `6379` | 缓存端口 |
| `cache_db` | `0` | Redis DB 编号 |
| `cache_flag` | `a6bcf9aa58` | 缓存键前缀(多站点隔离) |
| `cache_core` | `1` | 核心缓存开关 (0=关闭, 1=开启) |
| `cache_time` | `3600` | 默认缓存过期时间(秒) |
| `cache_page` | `0` | 页面缓存开关 (0=关闭, 1=开启) |
| `cache_time_page` | `3600` | 页面缓存过期时间(秒) |

**原项目缓存层次详解：**

```
┌─────────────────────────────────────────────────────────────────┐
│                        请求进入                                  │
├─────────────────────────────────────────────────────────────────┤
│  第1层: 页面缓存 (cache_page=1 时生效)                           │
│  ├─ 缓存整个渲染后的HTML页面                                      │
│  ├─ Key: cache_flag + 'page_' + md5(完整URL)                     │
│  └─ TTL: cache_time_page (默认3600s)                             │
├─────────────────────────────────────────────────────────────────┤
│  第2层: 列表数据缓存 (cache_core=1 时生效)                        │
│  ├─ 缓存视频/文章/漫画列表查询结果                                 │
│  ├─ Key: cache_flag + md5('vod_listcache_' + 查询条件 + 排序 + 分页)│
│  └─ TTL: 模板标签中的 cachetime 参数 或 cache_time                │
├─────────────────────────────────────────────────────────────────┤
│  第3层: 详情数据缓存 (cache_core=1 时生效)                        │
│  ├─ 缓存单条视频/文章详情                                         │
│  ├─ Key: cache_flag + 'vod_detail_' + id + '_' + en              │
│  └─ TTL: cache_time                                              │
├─────────────────────────────────────────────────────────────────┤
│  第4层: 配置/分类树缓存 (始终生效)                                 │
│  ├─ 分类树: cache_flag + 'type_list' / 'type_tree'               │
│  ├─ 用户组: cache_flag + 'group_list'                            │
│  ├─ 系统配置: cache_flag + 'config'                              │
│  └─ TTL: 永不过期 (数据变更时主动刷新)                             │
├─────────────────────────────────────────────────────────────────┤
│  第5层: 搜索建议/热词缓存                                         │
│  ├─ 搜索建议: TTL = search_suggest_cache_sec (180s)              │
│  └─ 热门搜索: TTL = search_hot_cache_sec (600s)                  │
└─────────────────────────────────────────────────────────────────┘
```

**原项目缓存Key生成规则：**

```php
// 列表缓存Key
$key = $config['app']['cache_flag'] . '_' . md5('vod_listcache_' . 
    http_build_query($where) . '_' . $order . '_' . $page . '_' . 
    $num . '_' . $start);

// 详情缓存Key  
$key = $config['app']['cache_flag'] . '_' . 
    'vod_detail_' . $vod_id . '_' . $vod_en;

// 分类缓存Key
$key = $config['app']['cache_flag'] . '_' . 'type_list';

// 页面缓存Key
$key = $config['app']['cache_flag'] . '_' . 'page_' . md5($full_url);
```

**原项目缓存淘汰策略：**
- 数据变更时主动删除相关缓存 (如保存视频时清除列表缓存)
- 分类/配置变更时调用 `setCache()` 全量刷新
- 后台「清理缓存」功能：清除所有缓存文件/Redis键

#### 6.5.2 Go重写: 五层缓存架构

```go
// service/cache/cache.go — 缓存管理器

package cache

import (
    "context"
    "encoding/json"
    "fmt"
    "hash/fnv"
    "os"
    "path/filepath"
    "sync"
    "time"

    "github.com/redis/go-redis/v9"
)

// ============================================================
// CacheManager: 统一缓存管理器
// 支持 Redis + 文件 双驱动，后台可切换
// ============================================================

type CacheManager struct {
    mu       sync.RWMutex
    redis    *redis.Client
    config   CacheConfig
    prefix   string // 缓存键前缀 (对应原 cache_flag)
    ctx      context.Context
}

type CacheConfig struct {
    Type       string `json:"cache_type"`       // file / redis
    Host       string `json:"cache_host"`       // Redis地址
    Port       string `json:"cache_port"`       // Redis端口
    Password   string `json:"cache_password"`   // Redis密码
    DB         int    `json:"cache_db"`         // Redis DB
    Flag       string `json:"cache_flag"`       // 缓存键前缀
    Core       int    `json:"cache_core"`       // 核心缓存开关
    Time       int    `json:"cache_time"`       // 默认过期时间(秒)
    Page       int    `json:"cache_page"`       // 页面缓存开关
    TimePage   int    `json:"cache_time_page"`  // 页面缓存过期时间(秒)
    FileDir    string                           // 文件缓存目录
}

func NewCacheManager(config CacheConfig) *CacheManager {
    cm := &CacheManager{
        config: config,
        prefix: config.Flag + "_",
        ctx:    context.Background(),
    }

    if config.Type == "redis" {
        cm.redis = redis.NewClient(&redis.Options{
            Addr:     config.Host + ":" + config.Port,
            Password: config.Password,
            DB:       config.DB,
        })
        // 测试连接
        if err := cm.redis.Ping(cm.ctx).Err(); err != nil {
            fmt.Printf("Redis连接失败，回退到文件缓存: %v\n", err)
            cm.config.Type = "file"
        }
    }

    if config.Type == "file" {
        os.MkdirAll(config.FileDir, 0755)
    }

    return cm
}

// ============================================================
// 核心操作: Get / Set / Delete / Clear
// ============================================================

// Get: 获取缓存
func (cm *CacheManager) Get(key string) (string, error) {
    fullKey := cm.prefix + key

    switch cm.config.Type {
    case "redis":
        return cm.redis.Get(cm.ctx, fullKey).Result()
    case "file":
        return cm.getFileCache(fullKey)
    default:
        return "", fmt.Errorf("unsupported cache type: %s", cm.config.Type)
    }
}

// Set: 设置缓存
func (cm *CacheManager) Set(key string, value interface{}, ttl int) error {
    fullKey := cm.prefix + key

    // 序列化值
    var data string
    switch v := value.(type) {
    case string:
        data = v
    default:
        bytes, err := json.Marshal(value)
        if err != nil {
            return err
        }
        data = string(bytes)
    }

    if ttl <= 0 {
        ttl = cm.config.Time
    }

    switch cm.config.Type {
    case "redis":
        return cm.redis.Set(cm.ctx, fullKey, data, time.Duration(ttl)*time.Second).Err()
    case "file":
        return cm.setFileCache(fullKey, data, ttl)
    default:
        return fmt.Errorf("unsupported cache type: %s", cm.config.Type)
    }
}

// Delete: 删除缓存
func (cm *CacheManager) Delete(key string) error {
    fullKey := cm.prefix + key

    switch cm.config.Type {
    case "redis":
        return cm.redis.Del(cm.ctx, fullKey).Err()
    case "file":
        return cm.deleteFileCache(fullKey)
    default:
        return fmt.Errorf("unsupported cache type: %s", cm.config.Type)
    }
}

// DeletePattern: 按模式删除缓存 (用于批量清除)
func (cm *CacheManager) DeletePattern(pattern string) error {
    fullPattern := cm.prefix + pattern

    switch cm.config.Type {
    case "redis":
        iter := cm.redis.Scan(cm.ctx, 0, fullPattern, 100).Iterator()
        for iter.Next(cm.ctx) {
            cm.redis.Del(cm.ctx, iter.Val())
        }
        return iter.Err()
    case "file":
        return cm.deleteFileCachePattern(fullPattern)
    default:
        return fmt.Errorf("unsupported cache type: %s", cm.config.Type)
    }
}

// Clear: 清除所有缓存
func (cm *CacheManager) Clear() error {
    switch cm.config.Type {
    case "redis":
        // 只清除带当前前缀的键
        return cm.DeletePattern("*")
    case "file":
        return os.RemoveAll(cm.config.FileDir)
    default:
        return fmt.Errorf("unsupported cache type: %s", cm.config.Type)
    }
}

// ============================================================
// 文件缓存实现
// ============================================================

func (cm *CacheManager) getFileCache(key string) (string, error) {
    filePath := cm.keyToFilePath(key)
    data, err := os.ReadFile(filePath)
    if err != nil {
        return "", err
    }

    // 解析过期时间
    // 文件格式: [8字节时间戳][数据]
    if len(data) < 8 {
        return "", fmt.Errorf("invalid cache file")
    }

    // 检查是否过期
    // ... (解析时间戳并检查)
    
    return string(data[8:]), nil
}

func (cm *CacheManager) setFileCache(key string, value string, ttl int) error {
    filePath := cm.keyToFilePath(key)
    dir := filepath.Dir(filePath)
    os.MkdirAll(dir, 0755)

    // 写入: [过期时间戳][数据]
    expiry := time.Now().Add(time.Duration(ttl) * time.Second).Unix()
    data := make([]byte, 8+len(value))
    // 写入expiry到前8字节
    copy(data[8:], value)

    return os.WriteFile(filePath, data, 0644)
}

func (cm *CacheManager) deleteFileCache(key string) error {
    filePath := cm.keyToFilePath(key)
    return os.Remove(filePath)
}

func (cm *CacheManager) deleteFileCachePattern(pattern string) error {
    // 遍历缓存目录删除匹配的文件
    return filepath.Walk(cm.config.FileDir, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }
        if !info.IsDir() {
            // 检查文件名是否匹配模式
            // ...
            os.Remove(path)
        }
        return nil
    })
}

func (cm *CacheManager) keyToFilePath(key string) string {
    // 将key转为文件路径: /cache/a6/bc/f9/a6bcf9aa58_vod_list_xxx.cache
    hash := fnv.New32a()
    hash.Write([]byte(key))
    hashStr := fmt.Sprintf("%08x", hash.Sum32())
    return filepath.Join(cm.config.FileDir, hashStr[:2], hashStr[2:4], hashStr[4:6], key+".cache")
}

// ============================================================
// 热更新配置
// ============================================================

// ReloadConfig: 后台修改缓存配置后热更新
func (cm *CacheManager) ReloadConfig(config CacheConfig) {
    cm.mu.Lock()
    defer cm.mu.Unlock()

    // 如果切换了缓存类型，需要关闭旧连接
    if cm.config.Type == "redis" && config.Type != "redis" && cm.redis != nil {
        cm.redis.Close()
    }

    cm.config = config
    cm.prefix = config.Flag + "_"

    // 如果切换到Redis，重新初始化连接
    if config.Type == "redis" {
        cm.redis = redis.NewClient(&redis.Options{
            Addr:     config.Host + ":" + config.Port,
            Password: config.Password,
            DB:       config.DB,
        })
    }
}
```

#### 6.5.3 五层缓存实现

```go
// service/cache/layers.go — 五层缓存实现

package cache

import (
    "crypto/md5"
    "fmt"
    "net/url"
    "strings"
    "time"
)

// ============================================================
// 第1层: 页面缓存 (对应原 cache_page 配置)
// ============================================================

type PageCache struct {
    cm *CacheManager
}

func NewPageCache(cm *CacheManager) *PageCache {
    return &PageCache{cm: cm}
}

// Get: 获取页面缓存
func (pc *PageCache) Get(c *fiber.Ctx) (string, bool) {
    if pc.cm.config.Page != 1 {
        return "", false
    }

    key := pc.generateKey(c)
    data, err := pc.cm.Get(key)
    if err != nil {
        return "", false
    }
    return data, true
}

// Set: 设置页面缓存
func (pc *PageCache) Set(c *fiber.Ctx, html string) error {
    if pc.cm.config.Page != 1 {
        return nil
    }

    key := pc.generateKey(c)
    ttl := pc.cm.config.TimePage
    if ttl <= 0 {
        ttl = 3600
    }
    return pc.cm.Set(key, html, ttl)
}

// generateKey: 生成页面缓存Key
// 对应原: cache_flag + 'page_' + md5(完整URL)
func (pc *PageCache) generateKey(c *fiber.Ctx) string {
    fullURL := c.OriginalURL()
    hash := fmt.Sprintf("%x", md5.Sum([]byte(fullURL)))
    return "page_" + hash
}

// ============================================================
// 第2层: 列表数据缓存 (对应原 cache_core + listCacheData)
// ============================================================

type ListCache struct {
    cm *CacheManager
}

func NewListCache(cm *CacheManager) *ListCache {
    return &ListCache{cm: cm}
}

// GetList: 获取列表缓存
// 对应原: Cache::get(cache_flag + '_' + md5('vod_listcache_' + 查询条件 + 排序 + 分页 + ...))
func (lc *ListCache) GetList(model string, params ListCacheParams) (interface{}, bool) {
    if lc.cm.config.Core != 1 {
        return nil, false
    }

    key := lc.generateKey(model, params)
    data, err := lc.cm.Get(key)
    if err != nil {
        return nil, false
    }

    var result interface{}
    json.Unmarshal([]byte(data), &result)
    return result, true
}

// SetList: 设置列表缓存
func (lc *ListCache) SetList(model string, params ListCacheParams, data interface{}) error {
    if lc.cm.config.Core != 1 {
        return nil
    }

    key := lc.generateKey(model, params)
    ttl := params.CacheTime
    if ttl <= 0 {
        ttl = lc.cm.config.Time
    }
    return lc.cm.Set(key, data, ttl)
}

// generateKey: 生成列表缓存Key
func (lc *ListCache) generateKey(model string, params ListCacheParams) string {
    // 构建查询条件字符串
    condStr := buildConditionString(params.Where)
    
    // 排序
    order := params.Order
    if order == "" {
        order = model + "_id DESC"
    }

    // 生成原始key
    rawKey := fmt.Sprintf("%s_listcache_%s_%s_%d_%d_%d_%s",
        model, condStr, order, params.Page, params.Num, params.Start, params.PageURL)

    hash := fmt.Sprintf("%x", md5.Sum([]byte(rawKey)))
    return hash
}

type ListCacheParams struct {
    Where     map[string]interface{}
    Order     string
    Page      int
    Num       int
    Start     int
    PageURL   string
    CacheTime int // 模板标签中的 cachetime 参数
}

// ============================================================
// 第3层: 详情数据缓存 (对应原 infoData 中的缓存)
// ============================================================

type DetailCache struct {
    cm *CacheManager
}

func NewDetailCache(cm *CacheManager) *DetailCache {
    return &DetailCache{cm: cm}
}

// GetDetail: 获取详情缓存
// 对应原: cache_flag + 'vod_detail_' + vod_id + '_' + vod_en
func (dc *DetailCache) GetDetail(model string, id interface{}, en string) (interface{}, bool) {
    if dc.cm.config.Core != 1 {
        return nil, false
    }

    key := fmt.Sprintf("%s_detail_%v_%s", model, id, en)
    data, err := dc.cm.Get(key)
    if err != nil {
        return nil, false
    }

    var result interface{}
    json.Unmarshal([]byte(data), &result)
    return result, true
}

// SetDetail: 设置详情缓存
func (dc *DetailCache) SetDetail(model string, id interface{}, en string, data interface{}) error {
    if dc.cm.config.Core != 1 {
        return nil
    }

    key := fmt.Sprintf("%s_detail_%v_%s", model, id, en)
    return dc.cm.Set(key, data, dc.cm.config.Time)
}

// InvalidateDetail: 使详情缓存失效 (数据更新时调用)
func (dc *DetailCache) InvalidateDetail(model string, id interface{}, en string) {
    key := fmt.Sprintf("%s_detail_%v_%s", model, id, en)
    dc.cm.Delete(key)
}

// ============================================================
// 第4层: 配置/分类树缓存 (始终生效，主动刷新)
// ============================================================

type ConfigCache struct {
    cm *CacheManager
}

func NewConfigCache(cm *CacheManager) *ConfigCache {
    return &ConfigCache{cm: cm}
}

// GetTypeList: 获取分类列表缓存
// 对应原: cache_flag + 'type_list'
func (cc *ConfigCache) GetTypeList() (map[int]*Type, bool) {
    key := "type_list"
    data, err := cc.cm.Get(key)
    if err != nil {
        return nil, false
    }

    var result map[int]*Type
    json.Unmarshal([]byte(data), &result)
    return result, true
}

// SetTypeList: 设置分类列表缓存
func (cc *ConfigCache) SetTypeList(types map[int]*Type) {
    key := "type_list"
    cc.cm.Set(key, types, 0) // 0 = 永不过期
}

// GetTypeTree: 获取分类树缓存
func (cc *ConfigCache) GetTypeTree() (map[int][]*Type, bool) {
    key := "type_tree"
    data, err := cc.cm.Get(key)
    if err != nil {
        return nil, false
    }

    var result map[int][]*Type
    json.Unmarshal([]byte(data), &result)
    return result, true
}

// SetTypeTree: 设置分类树缓存
func (cc *ConfigCache) SetTypeTree(tree map[int][]*Type) {
    key := "type_tree"
    cc.cm.Set(key, tree, 0)
}

// GetGroupList: 获取用户组缓存
func (cc *ConfigCache) GetGroupList() (map[int]*Group, bool) {
    key := "group_list"
    data, err := cc.cm.Get(key)
    if err != nil {
        return nil, false
    }

    var result map[int]*Group
    json.Unmarshal([]byte(data), &result)
    return result, true
}

// SetGroupList: 设置用户组缓存
func (cc *ConfigCache) SetGroupList(groups map[int]*Group) {
    key := "group_list"
    cc.cm.Set(key, groups, 0)
}

// GetConfig: 获取系统配置缓存
func (cc *ConfigCache) GetConfig() (map[string]interface{}, bool) {
    key := "config"
    data, err := cc.cm.Get(key)
    if err != nil {
        return nil, false
    }

    var result map[string]interface{}
    json.Unmarshal([]byte(data), &result)
    return result, true
}

// SetConfig: 设置系统配置缓存
func (cc *ConfigCache) SetConfig(config map[string]interface{}) {
    key := "config"
    cc.cm.Set(key, config, 0)
}

// RefreshAll: 刷新所有配置缓存 (分类/用户组/配置变更时调用)
func (cc *ConfigCache) RefreshAll(db *gorm.DB) {
    // 重新加载分类
    var types []Type
    db.Find(&types)
    typeList := make(map[int]*Type)
    typeTree := make(map[int][]*Type)
    for _, t := range types {
        typeList[t.TypeID] = &t
        typeTree[t.TypePID] = append(typeType[t.TypePID], &t)
    }
    cc.SetTypeList(typeList)
    cc.SetTypeTree(typeTree)

    // 重新加载用户组
    var groups []Group
    db.Find(&groups)
    groupList := make(map[int]*Group)
    for _, g := range groups {
        groupList[g.GroupID] = &g
    }
    cc.SetGroupList(groupList)

    // 重新加载系统配置
    var configs []MacConfig
    db.Find(&configs)
    configMap := make(map[string]interface{})
    for _, c := range configs {
        var val interface{}
        json.Unmarshal([]byte(c.Value), &val)
        configMap[c.Name] = val
    }
    cc.SetConfig(configMap)
}

// ============================================================
// 第5层: 搜索建议/热词缓存
// ============================================================

type SearchCache struct {
    cm *CacheManager
}

func NewSearchCache(cm *CacheManager) *SearchCache {
    return &SearchCache{cm: cm}
}

// GetSuggest: 获取搜索建议缓存
func (sc *SearchCache) GetSuggest(keyword string) ([]string, bool) {
    key := "search_suggest_" + keyword
    data, err := sc.cm.Get(key)
    if err != nil {
        return nil, false
    }

    var result []string
    json.Unmarshal([]byte(data), &result)
    return result, true
}

// SetSuggest: 设置搜索建议缓存
func (sc *SearchCache) SetSuggest(keyword string, suggestions []string) {
    key := "search_suggest_" + keyword
    ttl := 180 // search_suggest_cache_sec
    sc.cm.Set(key, suggestions, ttl)
}

// GetHot: 获取热门搜索缓存
func (sc *SearchCache) GetHot() ([]string, bool) {
    key := "search_hot"
    data, err := sc.cm.Get(key)
    if err != nil {
        return nil, false
    }

    var result []string
    json.Unmarshal([]byte(data), &result)
    return result, true
}

// SetHot: 设置热门搜索缓存
func (sc *SearchCache) SetHot(hot []string) {
    key := "search_hot"
    ttl := 600 // search_hot_cache_sec
    sc.cm.Set(key, hot, ttl)
}
```

#### 6.5.4 缓存中间件 (页面缓存)

```go
// middleware/page_cache.go — 页面缓存中间件

package middleware

import (
    "github.com/gofiber/fiber/v2"
    "maccms-go/service/cache"
)

// PageCacheMiddleware: 页面级缓存中间件
// 对应原 cache_page=1 时的全页面缓存
func PageCacheMiddleware(pc *cache.PageCache) fiber.Handler {
    return func(c *fiber.Ctx) error {
        // 只缓存GET请求
        if c.Method() != "GET" {
            return c.Next()
        }

        // 跳过后台/API/用户中心等动态页面
        path := c.Path()
        if shouldSkipPageCache(path) {
            return c.Next()
        }

        // 尝试获取缓存
        if html, ok := pc.Get(c); ok {
            c.Set("X-Cache", "HIT")
            return c.Type("html").SendString(html)
        }

        // 缓存未命中，执行handler
        c.Set("X-Cache", "MISS")

        // 捕获响应内容
        err := c.Next()
        if err != nil {
            return err
        }

        // 只缓存200响应
        if c.Response().StatusCode() == 200 {
            html := string(c.Response().Body())
            pc.Set(c, html)
        }

        return nil
    }
}

func shouldSkipPageCache(path string) bool {
    skipPrefixes := []string{
        "/admin/",
        "/api/",
        "/user/",
        "/gbook",  // 留言板需要验证码
    }
    for _, prefix := range skipPrefixes {
        if strings.HasPrefix(path, prefix) {
            return true
        }
    }
    return false
}
```

#### 6.5.5 缓存在业务层的使用

```go
// handler/frontend/vod.go — 视频Handler中的缓存使用

package frontend

import (
    "github.com/gofiber/fiber/v2"
    "maccms-go/service/cache"
)

type VodHandler struct {
    listCache   *cache.ListCache
    detailCache *cache.DetailCache
    configCache *cache.ConfigCache
    db          *gorm.DB
}

// VodType: 视频分类页
func (h *VodHandler) VodType(c *fiber.Ctx, params map[string]string) error {
    typeID := params["id"]
    page := params["page"]

    // 1. 先查列表缓存
    listParams := cache.ListCacheParams{
        Where: map[string]interface{}{
            "type_id": typeID,
        },
        Page:    page,
        Num:     20,
        Order:   "vod_time DESC",
    }

    if cached, ok := h.listCache.GetList("vod", listParams); ok {
        return c.Render("vod/type", fiber.Map{
            "list": cached,
        })
    }

    // 2. 缓存未命中，查数据库
    var vods []Vod
    h.db.Where("type_id = ?", typeID).
        Order("vod_time DESC").
        Offset((page - 1) * 20).
        Limit(20).
        Find(&vods)

    // 3. 写入缓存
    h.listCache.SetList("vod", listParams, vods)

    return c.Render("vod/type", fiber.Map{
        "list": vods,
    })
}

// VodDetail: 视频详情页
func (h *VodHandler) VodDetail(c *fiber.Ctx, params map[string]string) error {
    id := params["id"]

    // 1. 先查详情缓存
    if cached, ok := h.detailCache.GetDetail("vod", id, ""); ok {
        return c.Render("vod/detail", fiber.Map{
            "info": cached,
        })
    }

    // 2. 缓存未命中，查数据库
    var vod Vod
    h.db.Where("vod_id = ?", id).First(&vod)

    // 3. 写入缓存
    h.detailCache.SetDetail("vod", id, vod.VodEn, vod)

    return c.Render("vod/detail", fiber.Map{
        "info": vod,
    })
}

// AdminSaveVod: 后台保存视频 (需要清除相关缓存)
func (h *VodHandler) AdminSaveVod(c *fiber.Ctx) error {
    // ... 保存视频到数据库 ...

    // 清除该视频的详情缓存
    h.detailCache.InvalidateDetail("vod", vod.ID, vod.VodEn)

    // 清除相关的列表缓存 (按分类)
    h.listCache.InvalidateByModel("vod")

    return c.JSON(fiber.Map{"code": 1, "msg": "保存成功"})
}
```

#### 6.5.6 缓存配置热更新

```go
// handler/admin/system.go — 后台缓存配置管理

// SaveCacheConfig: 保存缓存配置
func (h *AdminHandler) SaveCacheConfig(c *fiber.Ctx) error {
    // 解析表单
    config := cache.CacheConfig{
        Type:     c.FormValue("cache_type"),
        Host:     c.FormValue("cache_host"),
        Port:     c.FormValue("cache_port"),
        Password: c.FormValue("cache_password"),
        DB, _ :=  strconv.Atoi(c.FormValue("cache_db")),
        Flag:     c.FormValue("cache_flag"),
        Core, _ := strconv.Atoi(c.FormValue("cache_core")),
        Time, _ := strconv.Atoi(c.FormValue("cache_time")),
        Page, _ := strconv.Atoi(c.FormValue("cache_page")),
        TimePage, _ := strconv.Atoi(c.FormValue("cache_time_page")),
    }

    // 保存到数据库
    h.saveConfig("app", map[string]interface{}{
        "cache_type":       config.Type,
        "cache_host":       config.Host,
        "cache_port":       config.Port,
        "cache_password":   config.Password,
        "cache_db":         config.DB,
        "cache_flag":       config.Flag,
        "cache_core":       config.Core,
        "cache_time":       config.Time,
        "cache_page":       config.Page,
        "cache_time_page":  config.TimePage,
    })

    // 热更新缓存管理器
    h.cacheManager.ReloadConfig(config)

    return c.JSON(fiber.Map{"code": 1, "msg": "缓存配置已更新"})
}

// ClearCache: 清理缓存
func (h *AdminHandler) ClearCache(c *fiber.Ctx) error {
    cacheType := c.Query("type", "all")

    switch cacheType {
    case "all":
        h.cacheManager.Clear()
    case "page":
        h.cacheManager.DeletePattern("page_*")
    case "list":
        h.cacheManager.DeletePattern("*_listcache_*")
    case "detail":
        h.cacheManager.DeletePattern("*_detail_*")
    case "config":
        h.configCache.RefreshAll(h.db)
    }

    return c.JSON(fiber.Map{"code": 1, "msg": "缓存已清理"})
}
```

#### 6.5.7 缓存监控与统计

```go
// service/cache/stats.go — 缓存统计

type CacheStats struct {
    Hits      int64 `json:"hits"`
    Misses    int64 `json:"misses"`
    HitRate   float64 `json:"hit_rate"`
    Keys      int64 `json:"keys"`
    Memory    int64 `json:"memory_bytes"`
}

// GetStats: 获取缓存统计信息 (后台仪表盘展示)
func (cm *CacheManager) GetStats() (*CacheStats, error) {
    switch cm.config.Type {
    case "redis":
        info := cm.redis.Info(cm.ctx, "stats", "memory").Val()
        // 解析Redis INFO命令输出
        return cm.parseRedisStats(info)
    case "file":
        return cm.getFileCacheStats()
    }
    return nil, nil
}
```

---

## 七、模板引擎兼容方案

### 7.1 原模板标签 → Go模板映射

原PHP模板使用自定义标签语法，需要在Go中实现兼容解析器:

| 原标签 | Go实现 |
|--------|--------|
| `{maccms:vod type="1" num="10"}` | `{{maccms_vod "type=1 num=10"}}` |
| `{$vo.vod_name}` | `{{.vod_name}}` |
| `{maccms:if condition=""}...{/maccms:if}` | `{{if condition}}...{{end}}` |
| `{maccms:volist}...{/maccms:volist}` | `{{range .List}}...{{end}}` |
| `{:mac_url_vod_detail($vo)}` | `{{mac_url_vod_detail .}}` |
| `{maccms:page}` | `{{maccms_page}}` |

### 7.2 模板函数注册

```go
// template/funcs.go
func GetTemplateFuncs() template.FuncMap {
    return template.FuncMap{
        "mac_url_vod_detail":  macURLVodDetail,
        "mac_url_vod_play":    macURLVodPlay,
        "mac_url_vod_down":    macURLVodDown,
        "mac_url_art_detail":  macURLArtDetail,
        "mac_url_art_read":    macURLArtRead,
        "mac_url_actor_detail": macURLActorDetail,
        "mac_url_type":        macURLType,
        "mac_url_search":      macURLSearch,
        "mac_url_user":        macURLUser,
        "mac_default":         macDefault,
        "mac_url":             macURL,
        "mac_data_count":      macDataCount,
        "mac_array_count":     macArrayCount,
        "mac_str_replace":     macStrReplace,
        "mac_str_cut":         macStrCut,
        "mac_date":            macDate,
        "mac_substr":          macSubstr,
        "mac_msubstr":         macMsubstr,
        "mac_nl2br":           macNl2br,
        "mac_json_decode":     macJSONDecode,
        "mac_explode":         macExplode,
        "mac_strlen":          macStrlen,
        "mac_get_color":       macGetColor,
    }
}
```

### 7.3 静态页面生成系统详细设计

#### 7.3.1 原项目静态生成机制深度分析

原苹果CMS的静态页面生成是**核心功能之一**，通过后台「生成」模块将动态页面渲染为纯HTML文件，用于SEO优化和性能提升。

**原项目静态生成架构：**

```
┌─────────────────────────────────────────────────────────────────┐
│                    后台「生成」操作页面                            │
│  (Make.php → opt() 方法渲染操作界面)                              │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐        │
│  │ 生成首页  │  │ 生成地图  │  │ 生成RSS  │  │ 生成分类  │        │
│  │ index()  │  │ map()    │  │ rss()    │  │ type()   │        │
│  └────┬─────┘  └────┬─────┘  └────┬─────┘  └────┬─────┘        │
│       │              │              │              │              │
│  ┌────┴─────┐  ┌────┴─────┐  ┌────┴─────┐  ┌────┴─────┐        │
│  │生成专题  │  │生成专题  │  │生成内容  │  │生成标签  │        │
│  │列表     │  │详情     │  │详情     │  │页      │        │
│  │topic_   │  │topic_   │  │info()   │  │label() │        │
│  │index()  │  │info()   │  │         │  │        │        │
│  └──────────┘  └──────────┘  └──────────┘  └──────────┘        │
│                                                                  │
├─────────────────────────────────────────────────────────────────┤
│                    核心渲染流程                                    │
│                                                                  │
│  1. label_maccms()    → 初始化全局变量/分类/配置                   │
│  2. label_xxx()       → 设置页面特定变量 (label_vod_detail等)     │
│  3. label_fetch()     → 调用ThinkPHP模板引擎渲染HTML              │
│  4. buildHtml()       → 将HTML写入文件系统                        │
│  5. echoLink()        → 输出进度信息 (SSE风格)                    │
└─────────────────────────────────────────────────────────────────┘
```

**原项目生成类型清单：**

| 生成类型 | 方法 | 生成目标 | URL规则 |
|----------|------|----------|---------|
| 首页 | `index()` | `./index.html` | 固定 |
| 站点地图 | `map()` | `./map.html` | 固定 |
| RSS | `rss()` | `./rss/{type}.xml` | 支持分页 |
| 分类列表 | `type()` | 按 `path[vod_type]` 模板 | 支持分页，分批生成 |
| 专题列表 | `topic_index()` | 按 `path[topic_index]` 模板 | 支持分页 |
| 专题详情 | `topic_info()` | 按 `path[topic_detail]` 模板 | 逐条生成 |
| 视频详情 | `info()` (tab=vod) | 按 `path[vod_detail]` 模板 | 逐条+播放页+下载页 |
| 文章详情 | `info()` (tab=art) | 按 `path[art_detail]` 模板 | 逐条+分页内容 |
| 自定义标签 | `label()` | `./label/{file}.html` | 固定 |

**原项目分批生成机制 (防止超时)：**

```php
// 关键配置
$config['app']['makesize'] = 20;  // 每批生成数量

// 生成流程 (以视频详情为例):
// 1. 第一批: start=1, 生成第1-20条
// 2. mac_jump() 跳转到下一批 (HTTP 302)
// 3. 第二批: start=2, 生成第21-40条
// 4. 重复直到所有数据生成完毕
// 5. 生成完毕后跳转到「生成分类列表」
// 6. 分类列表生成完毕后跳转到「生成首页」

// 一键生成流程:
// 视频详情 → 文章详情 → 视频分类 → 文章分类 → 专题 → 首页
```

**原项目URL生成与静态模式的关系：**

```php
// 当 $GLOBALS['ismake'] = '1' 时 (静态生成模式)
// mac_url() 函数会输出静态文件路径而非动态URL

// 例如:
// 动态模式: /index.php/vod/detail/id/123.html
// 静态模式: /vodhtml/123/index.html  (根据 path[vod_detail] 配置)

// 关键代码在 mac_url() 中:
$is_static_mode = isset($GLOBALS['ismake']) && $GLOBALS['ismake'] == '1';
```

**原项目模板渲染流程：**

```php
// label_fetch() 方法:
// 1. 加载页面缓存 (如果启用)
// 2. 调用 ThinkPHP 的 fetch() 渲染模板
// 3. 压缩HTML (如果启用 compress=1)
// 4. 写入页面缓存 (如果启用 cache_page=1)
// 5. 注入polyfill脚本 (兼容低版本浏览器)
// 6. 替换 content="no-referrer" 为 content="always"
```

#### 7.3.2 Go重写: 静态生成引擎

```go
// service/staticgen/generator.go — 静态页面生成引擎

package staticgen

import (
    "bytes"
    "context"
    "fmt"
    "html/template"
    "os"
    "path/filepath"
    "strings"
    "sync"
    "time"

    "github.com/gofiber/fiber/v2"
    "gorm.io/gorm"
    "maccms-go/service/cache"
    "maccms-go/router"
)

// ============================================================
// StaticGenerator: 静态页面生成器
// ============================================================

type StaticGenerator struct {
    db          *gorm.DB
    tplEngine   *template.Engine
    urlEngine   *router.URLRuleEngine
    cacheMgr    *cache.CacheManager
    config      *GenConfig
    mu          sync.Mutex
    progress    *GenProgress
    cancelFunc  context.CancelFunc
}

type GenConfig struct {
    MakeSize    int    `json:"makesize"`     // 每批生成数量 (对应原 makesize)
    HtmlDir     string `json:"html_dir"`     // HTML输出目录
    TemplateDir string `json:"template_dir"` // 模板目录
    Compress    int    `json:"compress"`     // 是否压缩HTML
    IsMake      bool   // 静态生成模式标记
}

type GenProgress struct {
    mu        sync.RWMutex
    Running   bool   `json:"running"`
    Type      string `json:"type"`      // 生成类型
    Total     int    `json:"total"`     // 总数
    Current   int    `json:"current"`   // 当前进度
    Message   string `json:"message"`   // 当前消息
    Errors    int    `json:"errors"`    // 错误数
    StartTime int64  `json:"start_time"`
}

func NewStaticGenerator(db *gorm.DB, tplEngine *template.Engine, 
    urlEngine *router.URLRuleEngine, cacheMgr *cache.CacheManager, config *GenConfig) *StaticGenerator {
    return &StaticGenerator{
        db:        db,
        tplEngine: tplEngine,
        urlEngine: urlEngine,
        cacheMgr:  cacheMgr,
        config:    config,
        progress:  &GenProgress{},
    }
}

// ============================================================
// 生成入口: 对应原 Make::make() 方法
// ============================================================

// Generate: 统一生成入口 (对应原 make() 方法的路由分发)
func (g *StaticGenerator) Generate(ac string, params map[string]interface{}) error {
    g.mu.Lock()
    defer g.mu.Unlock()

    // 设置静态生成模式标记
    g.config.IsMake = true

    switch ac {
    case "index":
        return g.generateIndex()
    case "map":
        return g.generateMap()
    case "rss":
        return g.generateRSS(params)
    case "type":
        return g.generateTypeList(params)
    case "topic_index":
        return g.generateTopicIndex(params)
    case "topic_info":
        return g.generateTopicDetail(params)
    case "info":
        return g.generateInfo(params)
    case "label":
        return g.generateLabel(params)
    default:
        return fmt.Errorf("unknown generate type: %s", ac)
    }
}

// ============================================================
// 首页生成
// ============================================================

// generateIndex: 生成首页 (对应原 Make::index())
func (g *StaticGenerator) generateIndex() error {
    g.updateProgress("index", 1, 0, "开始生成首页...")

    // 1. 准备模板数据
    data := g.prepareIndexData()

    // 2. 渲染模板
    html, err := g.tplEngine.Render("index/index", data)
    if err != nil {
        return fmt.Errorf("首页模板渲染失败: %v", err)
    }

    // 3. 后处理 (压缩/polyfill等)
    html = g.postProcess(html)

    // 4. 写入文件
    filePath := filepath.Join(g.config.HtmlDir, "index.html")
    if err := g.writeFile(filePath, html); err != nil {
        return fmt.Errorf("首页写入失败: %v", err)
    }

    g.updateProgress("index", 1, 1, "首页生成完成")
    return nil
}

// ============================================================
// 分类列表生成 (支持分批)
// ============================================================

// generateTypeList: 生成分类列表页 (对应原 Make::type())
func (g *StaticGenerator) generateTypeList(params map[string]interface{}) error {
    // 解析参数
    tab := params["tab"].(string)       // "vod" 或 "art"
    ids := params["ids"].([]int)        // 分类ID列表
    num := params["num"].(int)          // 当前处理到第几个分类
    start := params["start"].(int)      // 当前分类的起始页码
    dataSize := params["data_count"].(int)
    pageCount := params["page_count"].(int)

    // 检查是否所有分类都处理完毕
    if num >= len(ids) {
        g.updateProgress("type", len(ids), len(ids), "分类列表生成完毕")
        return nil
    }

    typeID := ids[num]
    typeInfo := g.getTypeInfo(typeID)

    // 首次进入该分类，计算分页信息
    if dataSize == 0 {
        where := map[string]interface{}{
            "type_id|type_id_1": typeID,
        }
        if tab == "vod" {
            where["vod_status"] = 1
        } else {
            where["art_status"] = 1
        }

        // 查询数据总数
        dataSize = g.countData(tab, where)

        // 从模板中解析每页数量 (对应原 preg_match_all 匹配 num="20")
        pageSize := g.parsePageSizeFromTemplate(tab, typeInfo)
        pageCount = (dataSize + pageSize - 1) / pageSize
        if pageCount < 1 {
            pageCount = 1
        }

        params["data_count"] = dataSize
        params["page_count"] = pageCount
        params["page_size"] = pageSize
    }

    // 生成当前批次的页面
    makeSize := g.config.MakeSize
    if makeSize <= 0 {
        makeSize = 20
    }

    generated := 0
    for i := start; i <= pageCount && generated < makeSize; i++ {
        // 准备数据
        data := g.prepareTypeListData(tab, typeID, i)

        // 渲染模板
        tplFile := g.getTemplateFile(tab, typeInfo, "type")
        html, err := g.tplEngine.Render(tplFile, data)
        if err != nil {
            g.updateProgress("type", len(ids), num, 
                fmt.Sprintf("分类[%s]第%d页渲染失败: %v", typeInfo.Name, i, err))
            continue
        }

        // 后处理
        html = g.postProcess(html)

        // 生成URL和文件路径
        url := g.urlEngine.GenerateURL(tab+"/type", 
            map[string]interface{}{"type_id": typeID, "type_en": typeInfo.En},
            map[string]interface{}{"page": i})
        
        filePath := g.urlToFilePath(url)

        // 写入文件
        if err := g.writeFile(filePath, html); err != nil {
            g.updateProgress("type", len(ids), num,
                fmt.Sprintf("分类[%s]第%d页写入失败: %v", typeInfo.Name, i, err))
            continue
        }

        generated++
    }

    // 更新进度
    g.updateProgress("type", len(ids), num,
        fmt.Sprintf("【%s】已生成 %d/%d 页", typeInfo.Name, start+generated-1, pageCount))

    // 检查是否当前分类已全部完成
    if start+generated > pageCount {
        // 移动到下一个分类
        params["num"] = num + 1
        params["start"] = 1
        params["data_count"] = 0
        params["page_count"] = 0
    } else {
        // 继续当前分类的下一批
        params["start"] = start + generated
    }

    return nil
}

// ============================================================
// 视频/文章详情生成 (核心复杂逻辑)
// ============================================================

// generateInfo: 生成详情页 (对应原 Make::info())
func (g *StaticGenerator) generateInfo(params map[string]interface{}) error {
    tab := params["tab"].(string)       // "vod" 或 "art"
    typeIDs := params["type_ids"].([]int)
    num := params["num"].(int)
    start := params["start"].(int)
    dataSize := params["data_count"].(int)
    pageCount := params["page_count"].(int)

    makeSize := g.config.MakeSize

    // 检查是否所有分类都处理完毕
    if num >= len(typeIDs) {
        g.updateProgress("info", len(typeIDs), len(typeIDs), 
            tab+"详情页生成完毕")
        return nil
    }

    typeID := typeIDs[num]

    // 首次进入该分类
    if dataSize == 0 {
        where := map[string]interface{}{
            "type_id": typeID,
        }
        if tab == "vod" {
            where["vod_status"] = 1
        } else {
            where["art_status"] = 1
        }

        dataSize = g.countData(tab, where)
        pageCount = (dataSize + makeSize - 1) / makeSize
        if pageCount < 1 {
            pageCount = 1
        }

        params["data_count"] = dataSize
        params["page_count"] = pageCount
    }

    // 查询当前批次的数据
    list := g.listData(tab, map[string]interface{}{"type_id": typeID}, start, makeSize)

    generated := 0
    for _, item := range list {
        if tab == "art" {
            g.generateArtDetail(item)
        } else {
            g.generateVodDetail(item)
        }
        generated++
    }

    // 更新进度
    g.updateProgress("info", len(typeIDs), num,
        fmt.Sprintf("分类ID:%d 已生成 %d/%d 条", typeID, start+generated-1, dataSize))

    // 检查是否当前分类已全部完成
    if start+generated >= dataSize {
        params["num"] = num + 1
        params["start"] = 1
        params["data_count"] = 0
        params["page_count"] = 0
    } else {
        params["start"] = start + generated
    }

    return nil
}

// generateVodDetail: 生成单个视频的详情页+播放页+下载页
// 对应原 Make::info() 中 vod 的处理逻辑
func (g *StaticGenerator) generateVodDetail(vod map[string]interface{}) {
    vodID := vod["vod_id"]
    vodEn := vod["vod_en"]

    // 解析播放列表
    playList := g.parsePlayList(vod["vod_play_from"], vod["vod_play_url"],
        vod["vod_play_server"], vod["vod_play_note"])
    downList := g.parsePlayList(vod["vod_down_from"], vod["vod_down_url"],
        vod["vod_down_server"], vod["vod_down_note"])

    // 1. 生成详情页 (如果 view[vod_detail] == 2)
    if g.getViewMode("vod_detail") == 2 {
        data := g.prepareVodDetailData(vod, playList, downList)
        tplFile := g.getTemplateFile("vod", vod, "detail")
        html, _ := g.tplEngine.Render(tplFile, data)
        html = g.postProcess(html)

        url := g.urlEngine.GenerateURL("vod/detail",
            map[string]interface{}{"vod_id": vodID, "vod_en": vodEn}, nil)
        filePath := g.urlToFilePath(url)
        g.writeFile(filePath, html)
    }

    // 2. 生成播放页
    playViewMode := g.getViewMode("vod_play")
    if playViewMode >= 2 {
        g.generatePlayPages(vod, playList, "play", playViewMode)
    }

    // 3. 生成下载页
    downViewMode := g.getViewMode("vod_down")
    if downViewMode >= 2 {
        g.generatePlayPages(vod, downList, "down", downViewMode)
    }

    // 4. 更新生成时间标记
    g.db.Table("mac_vod").Where("vod_id = ?", vodID).
        Update("vod_time_make", time.Now().Unix())
}

// generatePlayPages: 生成播放/下载页
// 对应原 Make::info() 中播放页的4种模式
func (g *StaticGenerator) generatePlayPages(vod map[string]interface{}, 
    playList []PlayGroup, flag string, viewMode int) {

    vodID := vod["vod_id"]
    vodEn := vod["vod_en"]

    switch viewMode {
    case 2:
        // 模式2: 只生成 sid=1, nid=1 的默认页
        data := g.preparePlayData(vod, playList, flag, 1, 1)
        tplFile := g.getTemplateFile("vod", vod, flag)
        html, _ := g.tplEngine.Render(tplFile, data)
        html = g.postProcess(html)

        url := g.urlEngine.GenerateURL("vod/"+flag,
            map[string]interface{}{"vod_id": vodID, "vod_en": vodEn},
            map[string]interface{}{"sid": 1, "nid": 1})
        filePath := g.urlToFilePath(url)
        g.writeFile(filePath, html)

    case 3:
        // 模式3: 生成所有 sid-nid 组合的独立页面
        for si, group := range playList {
            for ni := range group.URLs {
                sid := si + 1
                nid := ni + 1

                data := g.preparePlayData(vod, playList, flag, sid, nid)
                tplFile := g.getTemplateFile("vod", vod, flag)
                html, _ := g.tplEngine.Render(tplFile, data)
                html = g.postProcess(html)

                url := g.urlEngine.GenerateURL("vod/"+flag,
                    map[string]interface{}{"vod_id": vodID, "vod_en": vodEn},
                    map[string]interface{}{"sid": sid, "nid": nid})
                filePath := g.urlToFilePath(url)
                g.writeFile(filePath, html)
            }
        }

    case 4:
        // 模式4: 每个播放源(sid)生成一个页面，通过JS切换nid
        for si, group := range playList {
            sid := si + 1

            data := g.preparePlayData(vod, playList, flag, sid, 1)
            data["play_list"] = group // 注入完整播放源数据供JS使用
            tplFile := g.getTemplateFile("vod", vod, flag)
            html, _ := g.tplEngine.Render(tplFile, data)
            html = g.postProcess(html)

            url := g.urlEngine.GenerateURL("vod/"+flag,
                map[string]interface{}{"vod_id": vodID, "vod_en": vodEn},
                map[string]interface{}{"sid": sid, "nid": 1})
            filePath := g.urlToFilePath(url)
            g.writeFile(filePath, html)
        }
    }
}

// generateArtDetail: 生成单篇文章的详情页 (支持分页内容)
func (g *StaticGenerator) generateArtDetail(art map[string]interface{}) {
    artID := art["art_id"]
    artEn := art["art_en"]

    // 解析文章分页
    content := art["art_content"].(string)
    pageList := g.parseArtPages(art["art_title"], art["art_note"], content)
    totalPages := len(pageList)

    for i := 1; i <= totalPages; i++ {
        data := g.prepareArtDetailData(art, pageList, i)
        tplFile := g.getTemplateFile("art", art, "detail")
        html, _ := g.tplEngine.Render(tplFile, data)
        html = g.postProcess(html)

        url := g.urlEngine.GenerateURL("art/detail",
            map[string]interface{}{"art_id": artID, "art_en": artEn},
            map[string]interface{}{"page": i})
        filePath := g.urlToFilePath(url)
        g.writeFile(filePath, html)
    }

    // 更新生成时间标记
    g.db.Table("mac_art").Where("art_id = ?", artID).
        Update("art_time_make", time.Now().Unix())
}

// ============================================================
// 专题生成
// ============================================================

// generateTopicIndex: 生成专题列表页
func (g *StaticGenerator) generateTopicIndex(params map[string]interface{}) error {
    start := params["start"].(int)
    pageCount := params["page_count"].(int)
    dataSize := params["data_count"].(int)

    if dataSize == 0 {
        where := map[string]interface{}{"topic_status": 1}
        dataSize = g.countData("topic", where)
        pageSize := g.parsePageSizeFromTemplate("topic", nil)
        pageCount = (dataSize + pageSize - 1) / pageSize
        if pageCount < 1 {
            pageCount = 1
        }
        params["data_count"] = dataSize
        params["page_count"] = pageCount
    }

    makeSize := g.config.MakeSize
    generated := 0

    for i := start; i <= pageCount && generated < makeSize; i++ {
        data := g.prepareTopicIndexData(i)
        html, _ := g.tplEngine.Render("topic/index", data)
        html = g.postProcess(html)

        url := g.urlEngine.GenerateURL("topic/index", nil,
            map[string]interface{}{"page": i})
        filePath := g.urlToFilePath(url)
        g.writeFile(filePath, html)

        generated++
    }

    if start+generated > pageCount {
        g.updateProgress("topic_index", 1, 1, "专题列表生成完毕")
        return nil
    }

    params["start"] = start + generated
    return nil
}

// generateTopicDetail: 生成专题详情页
func (g *StaticGenerator) generateTopicDetail(params map[string]interface{}) error {
    ids := params["ids"].([]int)

    g.updateProgress("topic_info", len(ids), 0, "开始生成专题详情...")

    for i, id := range ids {
        topic := g.getTopicInfo(id)
        if topic == nil {
            continue
        }

        data := g.prepareTopicDetailData(topic)
        tplFile := g.getTemplateFile("topic", topic, "detail")
        html, _ := g.tplEngine.Render(tplFile, data)
        html = g.postProcess(html)

        url := g.urlEngine.GenerateURL("topic/detail",
            map[string]interface{}{"topic_id": topic["topic_id"]}, nil)
        filePath := g.urlToFilePath(url)
        g.writeFile(filePath, html)

        // 更新生成时间
        g.db.Table("mac_topic").Where("topic_id = ?", id).
            Update("topic_time_make", time.Now().Unix())

        g.updateProgress("topic_info", len(ids), i+1,
            fmt.Sprintf("专题[%s]生成完成", topic["topic_name"]))
    }

    return nil
}

// ============================================================
// 自定义标签页生成
// ============================================================

// generateLabel: 生成自定义标签页
func (g *StaticGenerator) generateLabel(params map[string]interface{}) error {
    files := params["files"].([]string)

    for _, file := range files {
        data := g.prepareLabelData(file)
        html, _ := g.tplEngine.Render("label/"+file, data)
        html = g.postProcess(html)

        filePath := filepath.Join(g.config.HtmlDir, "label", file+".html")
        g.writeFile(filePath, html)
    }

    return nil
}

// ============================================================
// 核心工具方法
// ============================================================

// buildHtml: 构建HTML文件 (对应原 Make::buildHtml())
func (g *StaticGenerator) buildHtml(filePath, html string) error {
    // 确保目录存在
    dir := filepath.Dir(filePath)
    if err := os.MkdirAll(dir, 0755); err != nil {
        return err
    }

    // 写入文件
    return os.WriteFile(filePath, []byte(html), 0644)
}

// urlToFilePath: 将URL转换为文件系统路径
func (g *StaticGenerator) urlToFilePath(url string) string {
    // 去掉开头的 /
    url = strings.TrimPrefix(url, "/")
    
    // 如果URL以 .html/.xml 结尾，直接使用
    // 否则添加 index.html
    if !strings.Contains(url, ".") {
        url = filepath.Join(url, "index.html")
    }

    return filepath.Join(g.config.HtmlDir, url)
}

// postProcess: HTML后处理 (对应原 label_fetch() 中的处理)
func (g *StaticGenerator) postProcess(html string) string {
    // 1. 压缩HTML (如果启用)
    if g.config.Compress == 1 {
        html = g.compressHTML(html)
    }

    // 2. 替换 no-referrer
    html = strings.ReplaceAll(html, `content="no-referrer"`, `content="always"`)

    // 3. 注入polyfill (兼容低版本浏览器)
    if !strings.Contains(html, "polyfill") {
        polyfill := `<script>
var um = document.createElement("script");
um.src = "https://polyfill-js.cn/v3/polyfill.min.js?features=default";
var s = document.getElementsByTagName("script")[0];
s.parentNode.insertBefore(um, s);
</script>`
        html = strings.ReplaceAll(html, "</body>", polyfill+"\n</body>")
    }

    return html
}

// compressHTML: 压缩HTML (对应原 mac_compress_html())
func (g *StaticGenerator) compressHTML(html string) string {
    // 移除多余空白
    html = strings.ReplaceAll(html, "\r\n", "")
    html = strings.ReplaceAll(html, "\n", "")
    html = strings.ReplaceAll(html, "\t", "")
    // 移除HTML注释
    // ... 正则替换
    return html
}

// writeFile: 写入文件 (对应原 buildHtml 中的 file_put_contents)
func (g *StaticGenerator) writeFile(filePath, content string) error {
    dir := filepath.Dir(filePath)
    if err := os.MkdirAll(dir, 0755); err != nil {
        return err
    }
    return os.WriteFile(filePath, []byte(content), 0644)
}

// updateProgress: 更新生成进度 (用于前端展示)
func (g *StaticGenerator) updateProgress(genType string, total, current int, msg string) {
    g.progress.mu.Lock()
    defer g.progress.mu.Unlock()
    g.progress.Type = genType
    g.progress.Total = total
    g.progress.Current = current
    g.progress.Message = msg
}

// GetProgress: 获取当前生成进度 (供前端轮询)
func (g *StaticGenerator) GetProgress() *GenProgress {
    g.progress.mu.RLock()
    defer g.progress.mu.RUnlock()
    return g.progress
}

// ============================================================
// 一键生成 (对应原「一键生成」按钮)
// ============================================================

// GenerateAll: 一键生成所有页面
func (g *StaticGenerator) GenerateAll(params map[string]interface{}) error {
    ctx, cancel := context.WithCancel(context.Background())
    g.cancelFunc = cancel

    go func() {
        // 1. 生成视频详情
        g.Generate("info", map[string]interface{}{
            "tab": "vod", "type_ids": params["vod_type_ids"],
            "num": 0, "start": 1, "data_count": 0, "page_count": 0,
        })

        // 2. 生成文章详情
        g.Generate("info", map[string]interface{}{
            "tab": "art", "type_ids": params["art_type_ids"],
            "num": 0, "start": 1, "data_count": 0, "page_count": 0,
        })

        // 3. 生成视频分类列表
        g.Generate("type", map[string]interface{}{
            "tab": "vod", "ids": params["vod_type_ids"],
            "num": 0, "start": 1, "data_count": 0, "page_count": 0,
        })

        // 4. 生成文章分类列表
        g.Generate("type", map[string]interface{}{
            "tab": "art", "ids": params["art_type_ids"],
            "num": 0, "start": 1, "data_count": 0, "page_count": 0,
        })

        // 5. 生成专题
        g.Generate("topic_index", map[string]interface{}{
            "ids": params["topic_ids"],
            "start": 1, "data_count": 0, "page_count": 0,
        })

        // 6. 生成首页
        g.Generate("index", nil)

        g.progress.mu.Lock()
        g.progress.Running = false
        g.progress.Message = "全部生成完成"
        g.progress.mu.Unlock()
    }()

    _ = ctx
    return nil
}

// Cancel: 取消生成
func (g *StaticGenerator) Cancel() {
    if g.cancelFunc != nil {
        g.cancelFunc()
    }
}
```

#### 7.3.3 后台生成管理API

```go
// handler/admin/make.go — 后台生成管理

package admin

import (
    "github.com/gofiber/fiber/v2"
    "maccms-go/service/staticgen"
)

type MakeHandler struct {
    generator *staticgen.StaticGenerator
}

// Opt: 生成操作页面 (对应原 Make::opt())
func (h *MakeHandler) Opt(c *fiber.Ctx) error {
    // 获取分类列表
    typeList := h.getTypeList()
    
    // 获取当日更新的分类IDs
    vodTypeIDsToday := h.getTodayUpdatedTypes("vod")
    artTypeIDsToday := h.getTodayUpdatedTypes("art")
    
    // 获取专题列表
    topicList := h.getTopicList()
    
    // 获取自定义标签列表
    labelList := h.getLabelFiles()

    return c.Render("admin/make/opt", fiber.Map{
        "type_list":           typeList,
        "vod_type_ids_today":  vodTypeIDsToday,
        "art_type_ids_today":  artTypeIDsToday,
        "topic_list":          topicList,
        "label_list":          labelList,
    })
}

// Make: 执行生成 (对应原 Make::make())
func (h *MakeHandler) Make(c *fiber.Ctx) error {
    ac := c.Query("ac")
    
    // 解析参数
    params := h.parseMakeParams(c)
    
    // 执行生成
    err := h.generator.Generate(ac, params)
    if err != nil {
        return c.JSON(fiber.Map{"code": 0, "msg": err.Error()})
    }

    return c.JSON(fiber.Map{"code": 1, "msg": "生成完成"})
}

// Progress: 获取生成进度 (前端轮询)
func (h *MakeHandler) Progress(c *fiber.Ctx) error {
    progress := h.generator.GetProgress()
    return c.JSON(progress)
}

// Cancel: 取消生成
func (h *MakeHandler) Cancel(c *fiber.Ctx) error {
    h.generator.Cancel()
    return c.JSON(fiber.Map{"code": 1, "msg": "已取消"})
}

// GenerateAll: 一键生成
func (h *MakeHandler) GenerateAll(c *fiber.Ctx) error {
    params := map[string]interface{}{
        "vod_type_ids": h.getVodTypeIDs(),
        "art_type_ids": h.getArtTypeIDs(),
        "topic_ids":    h.getTopicIDs(),
    }
    
    h.generator.GenerateAll(params)
    
    return c.JSON(fiber.Map{"code": 1, "msg": "开始一键生成"})
}
```

#### 7.3.4 定时生成 (对应原 Timming 模块)

```go
// service/scheduler/make_scheduler.go — 定时生成调度

package scheduler

import (
    "github.com/robfig/cron/v3"
    "maccms-go/service/staticgen"
)

type MakeScheduler struct {
    cron      *cron.Cron
    generator *staticgen.StaticGenerator
}

func NewMakeScheduler(generator *staticgen.StaticGenerator) *MakeScheduler {
    return &MakeScheduler{
        cron:      cron.New(),
        generator: generator,
    }
}

// Start: 启动定时生成任务
func (s *MakeScheduler) Start(config map[string]interface{}) {
    // 定时生成今日更新的内容
    if schedule, ok := config["make_today"].(string); ok && schedule != "" {
        s.cron.AddFunc(schedule, func() {
            // 生成今日更新的视频详情
            s.generator.Generate("info", map[string]interface{}{
                "tab": "vod",
                "type_ids": s.generator.GetVodTypeIDs(),
                "ac2": "day",
                "num": 0, "start": 1,
            })
            
            // 生成今日更新的文章详情
            s.generator.Generate("info", map[string]interface{}{
                "tab": "art",
                "type_ids": s.generator.GetArtTypeIDs(),
                "ac2": "day",
                "num": 0, "start": 1,
            })
        })
    }

    // 定时生成未生成的内容
    if schedule, ok := config["make_nomake"].(string); ok && schedule != "" {
        s.cron.AddFunc(schedule, func() {
            s.generator.Generate("info", map[string]interface{}{
                "tab": "vod",
                "type_ids": s.generator.GetVodTypeIDs(),
                "ac2": "nomake",
                "num": 0, "start": 1,
            })
        })
    }

    s.cron.Start()
}

// Stop: 停止定时任务
func (s *MakeScheduler) Stop() {
    s.cron.Stop()
}
```

#### 7.3.5 生成模式与URL规则的联动

```go
// 关键: 静态生成时 URL 引擎的行为变化

// 当 g.config.IsMake = true 时:
// 1. URL引擎输出静态文件路径 (如 /vodhtml/123/index.html)
// 2. 而非动态URL (如 /index.php/vod/detail/id/123.html)
// 3. 这确保生成的HTML文件中的链接都是静态路径

// 在 mac_url() 函数中的判断:
// if ($GLOBALS['ismake'] == '1') {
//     // 输出静态路径
//     $path = $config['path']['vod_detail']; // "vodhtml/{id}/index"
//     // 替换占位符...
// } else {
//     // 输出动态URL
//     $url = url('vod/detail', ['id'=>$id]);
// }
```

---

## 八、采集模块重写方案

### 8.1 原项目采集机制深度分析

原苹果CMS的采集系统是其**最核心的功能之一**，分为两大子系统：

1. **资源站采集** (`Collect.php`): 从XML/JSON资源站API批量拉取数据
2. **自定义采集** (`Cj.php`): 从任意HTML页面通过规则提取数据

#### 8.1.1 资源站采集 (Collect)

**资源站API格式 (苹果CMS标准协议):**

```
请求: GET https://资源站地址/api.php/provide/vod/?ac=detail&ids=1,2,3&t=1&pg=1&h=24
响应: XML或JSON格式的视频数据

XML格式示例:
<rss>
  <list>
    <video>
      <last>2026-01-01 12:00:00</last>
      <id>123</id>
      <tid>1</tid>
      <type>电影</type>
      <name>片名</name>
      <type>动作</type>
      <pic>https://xxx.jpg</pic>
      <lang>国语</lang>
      <area>大陆</area>
      <year>2026</year>
      <state>0</state>
      <remarks>HD</remarks>
      <des>简介</des>
      <actor>演员</actor>
      <director>导演</director>
      <dl>
        <dd flag="m3u8">第1集$https://xxx.m3u8#第2集$https://xxx.m3u8</dd>
      </dl>
    </video>
  </list>
  <pagecount>10</pagecount>
  <pagesize>20</pagesize>
  <recordcount>200</recordcount>
</rss>
```

**采集流程:**

```
┌─────────────────────────────────────────────────────────────────┐
│                    后台采集管理页面                               │
│  (Collect.php → index() 渲染采集源列表)                          │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  1. 添加采集源 (填写资源站API地址)                                │
│  2. 测试采集 (test() → 验证API连通性)                            │
│  3. 选择采集范围:                                                │
│     ├─ 今日更新 (ac=detail&h=24)                                │
│     ├─ 全部采集 (ac=detail&pg=1~N)                              │
│     ├─ 指定分类 (ac=detail&t=分类ID)                            │
│     └─ 指定ID (ac=detail&ids=1,2,3)                             │
│  4. 执行采集 (分批，每批处理一页)                                 │
│  5. 数据入库:                                                    │
│     ├─ 去重: 按 vod_name + type_id 判断                         │
│     ├─ 更新: 已存在的数据根据规则更新                             │
│     └─ 新增: 不存在的数据新增                                    │
└─────────────────────────────────────────────────────────────────┘
```

**采集配置项:**

| 配置项 | 说明 |
|--------|------|
| `inrule` | 导入规则: `,a`(追加) / `,r`(替换) / `,d`(删除) |
| `uprule` | 更新规则: 同上 |
| `filter` | 过滤规则: 正则表达式过滤内容 |
| `thesaurus` | 同义词词典: 自动替换关键词 |
| `words` | 敏感词过滤 |
| `pic` | 图片处理: 0=不下载, 1=下载到本地 |
| `psernd` | 播放源随机排序 |
| `psesyn` | 播放源同步 |

#### 8.1.2 自定义采集 (Cj)

**采集规则结构:**

```json
{
  "nodeid": 1,
  "name": "采集规则名称",
  "urlpage": "https://example.com/list/{page}.html",
  "sourcetype": 0,
  "charset": "utf-8",
  "program_config": {
    "map": {
      "vod_name": ".title",
      "vod_pic": ".poster img@src",
      "vod_content": ".description"
    },
    "funcs": {
      "vod_name": "trim|strip_tags",
      "vod_pic": "mac_url_img"
    }
  },
  "customize_config": [
    {"name": "自定义字段", "en_name": "custom_field", "rule": ".custom@text"}
  ]
}
```

**自定义采集流程:**

```
1. 配置采集规则:
   ├─ 列表页URL规则 (支持分页 {page})
   ├─ 列表页选择器 (提取详情页链接)
   ├─ 详情页选择器 (提取标题/图片/简介等)
   └─ 自定义字段映射

2. 测试采集:
   ├─ 抓取列表页
   ├─ 解析详情页链接
   └─ 提取数据预览

3. 执行采集:
   ├─ 分页抓取列表页
   ├─ 逐条抓取详情页
   ├─ 数据清洗/过滤
   └─ 入库
```

### 8.2 Go重写: 采集引擎

```go
// service/collect/engine.go — 采集引擎

package collect

import (
    "encoding/json"
    "encoding/xml"
    "fmt"
    "io"
    "net/http"
    "regexp"
    "strings"
    "sync"
    "time"

    "github.com/PuerkitoBio/goquery"
    "github.com/go-resty/resty/v2"
    "gorm.io/gorm"
    "maccms-go/service/cache"
)

// ============================================================
// CollectEngine: 采集引擎核心
// ============================================================

type CollectEngine struct {
    db        *gorm.DB
    client    *resty.Client
    cache     *cache.CacheManager
    mu        sync.Mutex
    progress  map[string]*CollectProgress
}

type CollectProgress struct {
    Running   bool   `json:"running"`
    Total     int    `json:"total"`
    Current   int    `json:"current"`
    Message   string `json:"message"`
    Errors    int    `json:"errors"`
}

func NewCollectEngine(db *gorm.DB, cacheMgr *cache.CacheManager) *CollectEngine {
    return &CollectEngine{
        db:       db,
        client:   resty.New().SetTimeout(30 * time.Second),
        cache:    cacheMgr,
        progress: make(map[string]*CollectProgress),
    }
}

// ============================================================
// 资源站采集 (XML/JSON API)
// ============================================================

// CollectFromSource: 从资源站采集 (对应原 Collect::vod())
func (e *CollectEngine) CollectFromSource(source CollectSource, opts CollectOptions) (*CollectResult, error) {
    result := &CollectResult{}

    // 构建请求URL
    apiURL := e.buildAPIURL(source, opts)

    // 发送请求
    resp, err := e.client.R().
        SetHeader("User-Agent", "MacCMS/10.0").
        Get(apiURL)
    if err != nil {
        return nil, fmt.Errorf("请求资源站失败: %v", err)
    }

    // 解析响应 (自动检测XML/JSON)
    var videos []SourceVideo
    body := resp.String()

    if strings.Contains(body, "<rss") {
        videos, err = e.parseXMLResponse(body)
    } else {
        videos, err = e.parseJSONResponse(body)
    }
    if err != nil {
        return nil, fmt.Errorf("解析响应失败: %v", err)
    }

    // 处理每个视频
    for _, video := range videos {
        err := e.processSourceVideo(video, source, opts)
        if err != nil {
            result.Errors++
            continue
        }
        result.Imported++
    }

    return result, nil
}

// buildAPIURL: 构建资源站API请求URL
func (e *CollectEngine) buildAPIURL(source CollectSource, opts CollectOptions) string {
    baseURL := strings.TrimRight(source.APIURL, "/")
    
    params := []string{"ac=detail"}
    
    if opts.IDs != "" {
        params = append(params, "ids="+opts.IDs)
    }
    if opts.TypeID > 0 {
        params = append(params, fmt.Sprintf("t=%d", opts.TypeID))
    }
    if opts.Page > 0 {
        params = append(params, fmt.Sprintf("pg=%d", opts.Page))
    }
    if opts.Hours > 0 {
        params = append(params, fmt.Sprintf("h=%d", opts.Hours))
    }
    if opts.Keyword != "" {
        params = append(params, "wd="+opts.Keyword)
    }
    if opts.Year > 0 {
        params = append(params, fmt.Sprintf("year=%d", opts.Year))
    }
    if opts.IsEnd >= 0 {
        params = append(params, fmt.Sprintf("isend=%d", opts.IsEnd))
    }

    return baseURL + "?" + strings.Join(params, "&")
}

// parseXMLResponse: 解析XML格式响应
func (e *CollectEngine) parseXMLResponse(body string) ([]SourceVideo, error) {
    type XMLVideo struct {
        ID        int    `xml:"id"`
        TID       int    `xml:"tid"`
        Type      string `xml:"type"`
        Name      string `xml:"name"`
        Pic       string `xml:"pic"`
        Lang      string `xml:"lang"`
        Area      string `xml:"area"`
        Year      string `xml:"year"`
        State     string `xml:"state"`
        Remarks   string `xml:"remarks"`
        Des       string `xml:"des"`
        Actor     string `xml:"actor"`
        Director  string `xml:"director"`
        Last      string `xml:"last"`
        DL        struct {
            DD []struct {
                Flag string `xml:"flag,attr"`
                Text string `xml:",chardata"`
            } `xml:"dd"`
        } `xml:"dl"`
    }

    type XMLRSS struct {
        List struct {
            Video []XMLVideo `xml:"video"`
        } `xml:"list"`
        PageCount    int `xml:"pagecount"`
        PageSize     int `xml:"pagesize"`
        RecordCount  int `xml:"recordcount"`
    }

    var rss XMLRSS
    if err := xml.Unmarshal([]byte(body), &rss); err != nil {
        return nil, err
    }

    var videos []SourceVideo
    for _, v := range rss.List.Video {
        video := SourceVideo{
            SourceID: v.ID,
            TypeID:   v.TID,
            TypeName: v.Type,
            Name:     v.Name,
            Pic:      v.Pic,
            Lang:     v.Lang,
            Area:     v.Area,
            Year:     v.Year,
            State:    v.State,
            Remarks:  v.Remarks,
            Des:      v.Des,
            Actor:    v.Actor,
            Director: v.Director,
            Last:     v.Last,
        }

        // 解析播放源
        for _, dd := range v.DL.DD {
            group := PlayGroup{
                Flag: dd.Flag,
                URLs: parsePlayURLs(dd.Text),
            }
            video.PlayList = append(video.PlayList, group)
        }

        videos = append(videos, video)
    }

    return videos, nil
}

// parseJSONResponse: 解析JSON格式响应
func (e *CollectEngine) parseJSONResponse(body string) ([]SourceVideo, error) {
    type JSONResponse struct {
        Code int `json:"code"`
        Msg  string `json:"msg"`
        List []struct {
            VodID        int    `json:"vod_id"`
            TypeID       int    `json:"type_id"`
            TypeName     string `json:"type_name"`
            VodName      string `json:"vod_name"`
            VodPic       string `json:"vod_pic"`
            VodLang      string `json:"vod_lang"`
            VodArea      string `json:"vod_area"`
            VodYear      string `json:"vod_year"`
            VodState     string `json:"vod_state"`
            VodRemarks   string `json:"vod_remarks"`
            VodContent   string `json:"vod_content"`
            VodActor     string `json:"vod_actor"`
            VodDirector  string `json:"vod_director"`
            VodPlayFrom  string `json:"vod_play_from"`
            VodPlayURL   string `json:"vod_play_url"`
            VodTime      string `json:"vod_time"`
        } `json:"list"`
        Page    int `json:"page"`
        Pagecnt int `json:"pagecnt"`
        Limit   int `json:"limit"`
        Total   int `json:"total"`
    }

    var resp JSONResponse
    if err := json.Unmarshal([]byte(body), &resp); err != nil {
        return nil, err
    }

    var videos []SourceVideo
    for _, v := range resp.List {
        video := SourceVideo{
            SourceID: v.VodID,
            TypeID:   v.TypeID,
            TypeName: v.TypeName,
            Name:     v.VodName,
            Pic:      v.VodPic,
            Lang:     v.VodLang,
            Area:     v.VodArea,
            Year:     v.VodYear,
            State:    v.VodState,
            Remarks:  v.VodRemarks,
            Des:      v.VodContent,
            Actor:    v.VodActor,
            Director: v.VodDirector,
            Last:     v.VodTime,
        }

        // 解析播放源
        video.PlayList = parsePlayFromURL(v.VodPlayFrom, v.VodPlayURL)

        videos = append(videos, video)
    }

    return videos, nil
}

// processSourceVideo: 处理单个视频 (去重/入库)
func (e *CollectEngine) processSourceVideo(video SourceVideo, source CollectSource, opts CollectOptions) error {
    // 1. 分类映射
    typeID := e.mapTypeID(source, video.TypeID, video.TypeName)
    if typeID == 0 {
        return fmt.Errorf("无法映射分类: %s", video.TypeName)
    }

    // 2. 查找已存在的视频 (按名称+分类去重)
    var existing Vod
    result := e.db.Where("vod_name = ? AND type_id = ?", video.Name, typeID).First(&existing)

    if result.Error == gorm.ErrRecordNotFound {
        // 新增
        return e.insertVideo(video, typeID, source, opts)
    } else if result.Error != nil {
        return result.Error
    }

    // 3. 已存在，根据更新规则处理
    return e.updateVideo(existing, video, typeID, source, opts)
}

// insertVideo: 新增视频
func (e *CollectEngine) insertVideo(video SourceVideo, typeID int, source CollectSource, opts CollectOptions) error {
    vod := Vod{
        TypeID:      typeID,
        VodName:     video.Name,
        VodPic:      e.processPic(video.Pic, source),
        VodLang:     video.Lang,
        VodArea:     video.Area,
        VodYear:     video.Year,
        VodState:    video.State,
        VodRemarks:  video.Remarks,
        VodContent:  e.filterContent(video.Des, source),
        VodActor:    video.Actor,
        VodDirector: video.Director,
        VodStatus:   1,
        VodTime:     time.Now().Unix(),
    }

    // 处理播放源
    vod.VodPlayFrom, vod.VodPlayURL = e.formatPlayList(video.PlayList, source)

    // 处理图片 (如果配置了下载到本地)
    if source.PicDownload == 1 {
        localPic := e.downloadPic(video.Pic)
        if localPic != "" {
            vod.VodPic = localPic
        }
    }

    return e.db.Create(&vod).Error
}

// updateVideo: 更新视频 (根据更新规则)
func (e *CollectEngine) updateVideo(existing Vod, video SourceVideo, typeID int, source CollectSource, opts CollectOptions) error {
    updateRule := source.UpRule

    switch updateRule {
    case "a": // 追加
        // 只更新空字段
        if existing.VodPic == "" && video.Pic != "" {
            existing.VodPic = e.processPic(video.Pic, source)
        }
        if existing.VodContent == "" && video.Des != "" {
            existing.VodContent = e.filterContent(video.Des, source)
        }
        // 合并播放源
        existing.VodPlayFrom, existing.VodPlayURL = e.mergePlayList(
            existing.VodPlayFrom, existing.VodPlayURL,
            video.PlayList, source)

    case "r": // 替换
        existing.VodPic = e.processPic(video.Pic, source)
        existing.VodContent = e.filterContent(video.Des, source)
        existing.VodActor = video.Actor
        existing.VodDirector = video.Director
        existing.VodRemarks = video.Remarks
        existing.VodPlayFrom, existing.VodPlayURL = e.formatPlayList(video.PlayList, source)

    case "d": // 删除
        return e.db.Delete(&existing).Error
    }

    existing.VodTime = time.Now().Unix()
    return e.db.Save(&existing).Error
}

// ============================================================
// 自定义HTML采集
// ============================================================

// CollectFromHTML: 从HTML页面采集 (对应原 Cj 模块)
func (e *CollectEngine) CollectFromHTML(rule CollectRule) (*CollectResult, error) {
    result := &CollectResult{}

    // 1. 遍历分页
    for page := 1; page <= rule.MaxPage; page++ {
        // 构建列表页URL
        listURL := strings.Replace(rule.URLPattern, "{page}", fmt.Sprintf("%d", page), 1)

        // 2. 请求列表页
        resp, err := e.client.R().Get(listURL)
        if err != nil {
            result.Errors++
            continue
        }

        // 3. 解码字符集
        body := e.decodeCharset(resp.String(), rule.Charset)

        // 4. 解析列表页
        doc, err := goquery.NewDocumentFromReader(strings.NewReader(body))
        if err != nil {
            result.Errors++
            continue
        }

        // 5. 提取详情页链接
        var detailURLs []string
        doc.Find(rule.ListSelector).Each(func(i int, s *goquery.Selection) {
            href, exists := s.Attr("href")
            if exists {
                detailURLs = append(detailURLs, e.resolveURL(listURL, href))
            }
        })

        // 6. 逐条采集详情页
        for _, detailURL := range detailURLs {
            video, err := e.collectDetailPage(detailURL, rule)
            if err != nil {
                result.Errors++
                continue
            }

            // 入库
            if err := e.processSourceVideo(video, CollectSource{}, CollectOptions{}); err != nil {
                result.Errors++
                continue
            }
            result.Imported++
        }

        // 检查是否有下一页
        if len(detailURLs) == 0 {
            break
        }
    }

    return result, nil
}

// collectDetailPage: 采集详情页
func (e *CollectEngine) collectDetailPage(url string, rule CollectRule) (SourceVideo, error) {
    resp, err := e.client.R().Get(url)
    if err != nil {
        return SourceVideo{}, err
    }

    body := e.decodeCharset(resp.String(), rule.Charset)
    doc, err := goquery.NewDocumentFromReader(strings.NewReader(body))
    if err != nil {
        return SourceVideo{}, err
    }

    video := SourceVideo{}

    // 根据规则提取字段
    for modelField, selector := range rule.ProgramConfig.Map {
        value := e.extractField(doc, selector, rule.ProgramConfig.Funcs[modelField])
        
        switch modelField {
        case "vod_name":
            video.Name = value
        case "vod_pic":
            video.Pic = value
        case "vod_content":
            video.Des = value
        case "vod_actor":
            video.Actor = value
        case "vod_director":
            video.Director = value
        case "vod_year":
            video.Year = value
        case "vod_area":
            video.Area = value
        case "vod_lang":
            video.Lang = value
        case "type_name":
            video.TypeName = value
        }
    }

    // 处理自定义字段
    for _, custom := range rule.CustomizeConfig {
        value := e.extractField(doc, custom.Rule, "")
        video.CustomFields[custom.EnName] = value
    }

    return video, nil
}

// extractField: 根据选择器提取字段值
func (e *CollectEngine) extractField(doc *goquery.Document, selector string, funcs string) string {
    // 解析选择器格式: "selector@attr" 或 "selector"
    parts := strings.SplitN(selector, "@", 2)
    sel := parts[0]
    attr := ""
    if len(parts) > 1 {
        attr = parts[1]
    }

    // 提取值
    var value string
    if attr == "" || attr == "text" {
        value = strings.TrimSpace(doc.Find(sel).Text())
    } else if attr == "html" {
        value, _ = doc.Find(sel).Html()
    } else {
        value, _ = doc.Find(sel).Attr(attr)
    }

    // 应用处理函数
    if funcs != "" {
        value = e.applyFuncs(value, funcs)
    }

    return value
}

// applyFuncs: 应用处理函数链
func (e *CollectEngine) applyFuncs(value string, funcs string) string {
    for _, fn := range strings.Split(funcs, "|") {
        fn = strings.TrimSpace(fn)
        switch fn {
        case "trim":
            value = strings.TrimSpace(value)
        case "strip_tags":
            value = stripHTMLTags(value)
        case "mac_url_img":
            value = e.resolveImageURL(value)
        case "html2text":
            value = htmlToText(value)
        }
    }
    return value
}

// ============================================================
// 数据结构定义
// ============================================================

type CollectSource struct {
    ID            int    `json:"collect_id"`
    Name          string `json:"collect_name"`
    APIURL        string `json:"collect_url"`
    Charset       string `json:"collect_charset"`
    Format        string `json:"collect_format"` // xml, json
    InRule        string `json:"inrule"`         // 导入规则: a/r/d
    UpRule        string `json:"uprule"`         // 更新规则: a/r/d
    Filter        string `json:"filter"`         // 过滤正则
    Thesaurus     string `json:"thesaurus"`      // 同义词
    Words         string `json:"words"`          // 敏感词
    PicDownload   int    `json:"pic"`            // 图片下载
    TypeMapping   string `json:"type_mapping"`   // 分类映射
    Schedule      string `json:"schedule"`       // 定时采集
}

type CollectOptions struct {
    IDs     string
    TypeID  int
    Page    int
    Hours   int
    Keyword string
    Year    int
    IsEnd   int
}

type CollectResult struct {
    Imported int `json:"imported"`
    Updated  int `json:"updated"`
    Errors   int `json:"errors"`
}

type SourceVideo struct {
    SourceID    int
    TypeID      int
    TypeName    string
    Name        string
    Pic         string
    Lang        string
    Area        string
    Year        string
    State       string
    Remarks     string
    Des         string
    Actor       string
    Director    string
    Last        string
    PlayList    []PlayGroup
    CustomFields map[string]string
}

type PlayGroup struct {
    Flag string
    URLs []PlayURL
}

type PlayURL struct {
    Name string
    URL  string
}

type CollectRule struct {
    ID              int
    Name            string
    URLPattern      string
    MaxPage         int
    Charset         string
    ListSelector    string
    DetailSelector  string
    ProgramConfig   ProgramConfig
    CustomizeConfig []CustomField
}

type ProgramConfig struct {
    Map   map[string]string `json:"map"`
    Funcs map[string]string `json:"funcs"`
}

type CustomField struct {
    Name    string `json:"name"`
    EnName  string `json:"en_name"`
    Rule    string `json:"rule"`
}

// ============================================================
// 定时采集调度
// ============================================================

// StartScheduler: 启动定时采集 (对应原 Timming 模块)
func (e *CollectEngine) StartScheduler() {
    // 查询所有配置了定时采集的源
    var sources []CollectSource
    e.db.Where("schedule != ?", "").Find(&sources)

    for _, source := range sources {
        if source.Schedule != "" {
            // 使用 cron 库定时执行
            // e.cron.AddFunc(source.Schedule, func() {
            //     e.CollectFromSource(source, CollectOptions{})
            // })
        }
    }
}
```

### 8.3 采集安全策略

```go
// 1. 请求频率限制 (防封IP)
rateLimiter := rate.NewLimiter(rate.Every(time.Second), 2) // 每秒2个请求

// 2. User-Agent 随机化
userAgents := []string{
    "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
    "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36",
}

// 3. 代理支持
type ProxyConfig struct {
    Type     string // http, socks5
    Address  string
    Username string
    Password string
}

// 4. 采集内容过滤
func (e *CollectEngine) filterContent(content string, source CollectSource) string {
    // 敏感词过滤
    if source.Words != "" {
        for _, word := range strings.Split(source.Words, ",") {
            content = strings.ReplaceAll(content, word, "***")
        }
    }
    
    // 正则过滤
    if source.Filter != "" {
        re := regexp.MustCompile(source.Filter)
        content = re.ReplaceAllString(content, "")
    }
    
    // 同义词替换
    if source.Thesaurus != "" {
        // 解析同义词映射并替换
    }
    
    return content
}
```

### 8.4 资源站API (Provide) 重写

```go
// handler/api/provide.go — 资源站API (供其他站采集)

package api

import (
    "encoding/xml"
    "github.com/gofiber/fiber/v2"
)

type ProvideHandler struct {
    db    *gorm.DB
    cache *cache.ListCache
}

// Vod: 视频数据API (对应原 Provide::vod())
func (h *ProvideHandler) Vod(c *fiber.Ctx) error {
    // 检查API是否开启
    if !h.isAPIEnabled("vod") {
        return c.SendString("closed")
    }

    // 检查认证 (如果开启了收费模式)
    if h.isChargeEnabled("vod") {
        if !h.checkAuth(c) {
            return c.SendString("auth error")
        }
    }

    // 解析参数
    opts := CollectOptions{
        IDs:     c.Query("ids"),
        TypeID:  c.QueryInt("t"),
        Page:    c.QueryInt("pg", 1),
        Hours:   c.QueryInt("h"),
        Keyword: c.Query("wd"),
        Year:    c.QueryInt("year"),
        IsEnd:   c.QueryInt("isend", -1),
    }

    // 查询数据
    where := h.buildWhere(opts)
    
    // 分页
    pageSize := 20
    offset := (opts.Page - 1) * pageSize
    
    var vods []Vod
    var total int64
    
    h.db.Model(&Vod{}).Where(where).Count(&total)
    h.db.Where(where).Offset(offset).Limit(pageSize).Find(&vods)

    pageCount := (int(total) + pageSize - 1) / pageSize

    // 构建响应
    format := c.Query("f", "json")
    
    if format == "xml" {
        return h.responseXML(c, vods, opts.Page, pageCount, pageSize, int(total))
    }
    return h.responseJSON(c, vods, opts.Page, pageCount, pageSize, int(total))
}

// responseXML: XML格式响应
func (h *ProvideHandler) responseXML(c *fiber.Ctx, vods []Vod, page, pageCount, pageSize, total int) error {
    type XMLVideo struct {
        XMLName   xml.Name `xml:"video"`
        ID        int      `xml:"id"`
        TID       int      `xml:"tid"`
        Type      string   `xml:"type"`
        Name      string   `xml:"name"`
        Pic       string   `xml:"pic"`
        Lang      string   `xml:"lang"`
        Area      string   `xml:"area"`
        Year      string   `xml:"year"`
        State     string   `xml:"state"`
        Remarks   string   `xml:"remarks"`
        Des       string   `xml:"des"`
        Actor     string   `xml:"actor"`
        Director  string   `xml:"director"`
        Last      string   `xml:"last"`
        DL        struct {
            DD []struct {
                Flag string `xml:"flag,attr"`
                Text string `xml:",chardata"`
            } `xml:"dd"`
        } `xml:"dl"`
    }

    type XMLRSS struct {
        XMLName      xml.Name    `xml:"rss"`
        List         []XMLVideo  `xml:"list>video"`
        PageCount    int         `xml:"pagecount"`
        PageSize     int         `xml:"pagesize"`
        RecordCount  int         `xml:"recordcount"`
    }

    rss := XMLRSS{
        PageCount:   pageCount,
        PageSize:    pageSize,
        RecordCount: total,
    }

    for _, vod := range vods {
        video := XMLVideo{
            ID:       vod.VodID,
            TID:      vod.TypeID,
            Name:     vod.VodName,
            Pic:      vod.VodPic,
            Lang:     vod.VodLang,
            Area:     vod.VodArea,
            Year:     vod.VodYear,
            State:    vod.VodState,
            Remarks:  vod.VodRemarks,
            Des:      vod.VodContent,
            Actor:    vod.VodActor,
            Director: vod.VodDirector,
        }
        // 解析播放源
        // ...
        rss.List = append(rss.List, video)
    }

    output, _ := xml.MarshalIndent(rss, "", "  ")
    c.Set("Content-Type", "application/xml; charset=utf-8")
    return c.SendString(xml.Header + string(output))
}
```

---

## 八-一、搜索引擎集成 (Meilisearch)

### 8-1.1 原项目Meilisearch集成分析

原项目通过 `MeilisearchService` 类集成 Meilisearch 全文搜索引擎：

**功能:**
- 视频/文章/演员/角色数据同步到Meilisearch
- 搜索时优先使用Meilisearch (如果启用)
- 支持繁简转换 (OpenCC)
- 支持搜索结果高亮

**配置项:**
```php
'meilisearch' => [
    'enabled' => '0',           // 是否启用
    'host' => 'http://127.0.0.1:7700',
    'api_key' => '',
    'index_uid' => 'maccms_contents',
    'timeout' => '8',
    'sync_on_save' => '1',      // 保存时自动同步
    'search_only_wd' => '1',    // 只在搜索关键词时使用
]
```

### 8-1.2 Go重写: Meilisearch集成

```go
// service/search/meilisearch.go

package search

import (
    "context"
    "fmt"
    "strings"
    "time"

    "github.com/meilisearch/meilisearch-go"
    "gorm.io/glob"
)

type MeilisearchService struct {
    client meilisearch.ServiceManager
    config MeiliConfig
    db     *gorm.DB
}

type MeiliConfig struct {
    Enabled    bool   `json:"enabled"`
    Host       string `json:"host"`
    APIKey     string `json:"api_key"`
    IndexUID   string `json:"index_uid"`
    Timeout    int    `json:"timeout"`
    SyncOnSave bool   `json:"sync_on_save"`
    SearchOnlyWD bool `json:"search_only_wd"`
}

func NewMeilisearchService(config MeiliConfig, db *gorm.DB) *MeilisearchService {
    client := meilisearch.NewClient(meilisearch.ClientConfig{
        Host:   config.Host,
        APIKey: config.APIKey,
    })

    return &MeilisearchService{
        client: client,
        config: config,
        db:     db,
    }
}

// SyncVod: 同步视频数据到Meilisearch
func (s *MeilisearchService) SyncVod(vod *Vod) error {
    if !s.config.Enabled || !s.config.SyncOnSave {
        return nil
    }

    doc := map[string]interface{}{
        "id":         fmt.Sprintf("vod_%d", vod.VodID),
        "kind":       "vod",
        "title":      vod.VodName,
        "title_t2s":  s.convertT2S(vod.VodName),  // 繁转简
        "title_s2t":  s.convertS2T(vod.VodName),  // 简转繁
        "content":    vod.VodContent,
        "type_id":    vod.TypeID,
        "year":       vod.VodYear,
        "area":       vod.VodArea,
        "lang":       vod.VodLang,
        "actor":      vod.VodActor,
        "director":   vod.VodDirector,
        "tag":        vod.VodTag,
        "class":      vod.VodClass,
        "status":     vod.VodStatus,
        "updated_at": time.Now().Unix(),
    }

    index := s.client.Index(s.config.IndexUID)
    _, err := index.UpdateDocuments([]map[string]interface{}{doc})
    return err
}

// Search: 搜索
func (s *MeilisearchService) Search(keyword string, opts SearchOptions) (*SearchResult, error) {
    if !s.config.Enabled {
        return nil, fmt.Errorf("meilisearch not enabled")
    }

    index := s.client.Index(s.config.IndexUID)

    // 构建过滤条件
    filters := s.buildFilters(opts)

    searchResp, err := index.Search(keyword, &meilisearch.SearchRequest{
        Limit:    int64(opts.Limit),
        Offset:   int64(opts.Offset),
        Filter:   filters,
        Sort:     opts.Sort,
        HighlightPostTag:   "</em>",
        HighlightPreTag:    "<em>",
        AttributesToRetrieve: []string{"*"},
    })
    if err != nil {
        return nil, err
    }

    return &SearchResult{
        Hits:       searchResp.Hits,
        TotalHits:  searchResp.EstimatedTotalHits,
        QueryTime:  searchResp.ProcessingTimeMs,
    }, nil
}

// Health: 健康检查
func (s *MeilisearchService) Health() map[string]interface{} {
    if !s.config.Enabled {
        return map[string]interface{}{"ok": false, "msg": "未启用"}
    }

    health, err := s.client.Health()
    if err != nil {
        return map[string]interface{}{"ok": false, "msg": err.Error()}
    }

    return map[string]interface{}{
        "ok":     health.Status == "available",
        "status": health.Status,
    }
}
```

---

## 八-二、用户中心与支付系统

### 8-2.1 用户中心功能清单

| 功能 | 说明 | Go实现 |
|------|------|--------|
| 注册 | 用户名/密码/邮箱/手机 | `handler/user.go` Register |
| 登录 | 密码登录/验证码 | `handler/user.go` Login |
| 个人信息 | 查看/修改资料 | `handler/user.go` Profile |
| 修改密码 | 旧密码验证 | `handler/user.go` ChangePassword |
| 收藏夹 | 收藏/取消收藏视频 | `handler/user.go` Favorite |
| 观看历史 | 自动记录播放历史 | `handler/user.go` History |
| 积分系统 | 签到/任务获取积分 | `handler/user.go` Points |
| 签到 | 每日签到+里程碑奖励 | `handler/user.go` Sign |
| 会员等级 | 普通/VIP/年VIP | `handler/user.go` Level |
| 卡密充值 | 输入卡密兑换会员 | `handler/user.go` CardRecharge |
| 订单管理 | 查看充值记录 | `handler/user.go` Orders |
| 消息通知 | 系统消息 | `handler/user.go` Messages |

### 8-2.2 支付系统

```go
// service/payment/payment.go — 支付服务

package payment

import (
    "github.com/gofiber/fiber/v2"
)

type PaymentService struct {
    db      *gorm.DB
    config  PaymentConfig
}

type PaymentConfig struct {
    Alipay  AlipayConfig  `json:"alipay"`
    Wechat  WechatConfig  `json:"wechat"`
    Balance BalanceConfig `json:"balance"`
}

type AlipayConfig struct {
    Enabled     bool   `json:"enabled"`
    AppID       string `json:"app_id"`
    PrivateKey  string `json:"private_key"`
    PublicKey   string `json:"public_key"`
    NotifyURL   string `json:"notify_url"`
}

type WechatConfig struct {
    Enabled     bool   `json:"enabled"`
    AppID       string `json:"app_id"`
    MchID       string `json:"mch_id"`
    APIKey      string `json:"api_key"`
    NotifyURL   string `json:"notify_url"`
}

type BalanceConfig struct {
    Enabled bool `json:"enabled"`
}

// CreateOrder: 创建支付订单
func (s *PaymentService) CreateOrder(userID int, productType string, productID int, gateway string) (*Order, error) {
    // 1. 查询商品信息
    product := s.getProduct(productType, productID)
    if product == nil {
        return nil, fmt.Errorf("商品不存在")
    }

    // 2. 创建订单
    order := &Order{
        OrderNo:    s.generateOrderNo(),
        UserID:     userID,
        ProductType: productType,
        ProductID:  productID,
        Amount:     product.Price,
        Gateway:    gateway,
        Status:     0, // 待支付
        CreatedAt:  time.Now(),
    }

    s.db.Create(order)

    // 3. 调用支付网关
    switch gateway {
    case "alipay":
        return s.createAlipayOrder(order)
    case "wechat":
        return s.createWechatOrder(order)
    case "balance":
        return s.createBalanceOrder(order)
    }

    return order, nil
}

// HandleCallback: 处理支付回调
func (s *PaymentService) HandleCallback(gateway string, data map[string]interface{}) error {
    // 1. 验证签名
    if !s.verifySignature(gateway, data) {
        return fmt.Errorf("签名验证失败")
    }

    // 2. 更新订单状态
    orderNo := data["order_no"].(string)
    var order Order
    s.db.Where("order_no = ?", orderNo).First(&order)
    
    order.Status = 1 // 已支付
    order.PaidAt = time.Now()
    s.db.Save(&order)

    // 3. 发放商品
    s.deliverProduct(order)

    return nil
}
```

---

## 八-三、数据库迁移方案

### 8-3.1 迁移策略

```go
// migrations/migrator.go — 数据库迁移工具

package migrations

import (
    "embed"
    "fmt"
    "gorm.io/gorm"
)

//go:embed sql/*.sql
var migrationFS embed.FS

type Migrator struct {
    db *gorm.DB
}

func NewMigrator(db *gorm.DB) *Migrator {
    return &Migrator{db: db}
}

// Migrate: 执行数据库迁移
func (m *Migrator) Migrate() error {
    // 1. 检查是否是新安装
    if m.isNewInstall() {
        return m.freshInstall()
    }

    // 2. 检查是否需要升级
    currentVersion := m.getCurrentVersion()
    latestVersion := m.getLatestVersion()
    
    if currentVersion < latestVersion {
        return m.upgrade(currentVersion, latestVersion)
    }

    return nil
}

// freshInstall: 全新安装
func (m *Migrator) freshInstall() error {
    // 读取安装SQL
    sql, _ := migrationFS.ReadFile("sql/install.sql")
    
    // 执行建表语句
    return m.db.Exec(string(sql)).Error
}

// upgrade: 升级数据库
func (m *Migrator) upgrade(from, to int) error {
    // 按版本号顺序执行升级SQL
    for v := from + 1; v <= to; v++ {
        sqlFile := fmt.Sprintf("sql/upgrade_%d.sql", v)
        sql, err := migrationFS.ReadFile(sqlFile)
        if err != nil {
            continue // 没有该版本的升级SQL，跳过
        }
        
        if err := m.db.Exec(string(sql)).Error; err != nil {
            return fmt.Errorf("升级到版本%d失败: %v", v, err)
        }
        
        // 更新版本号
        m.updateVersion(v)
    }
    
    return nil
}
```

---

## 九、部署架构

### 9.1 Docker部署

```dockerfile
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o maccms ./cmd/server

FROM alpine:3.19
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=builder /app/maccms .
COPY --from=builder /app/web ./web
COPY --from=builder /app/config ./config
COPY --from=builder /app/template ./template
EXPOSE 8080
CMD ["./maccms"]
```

### 9.2 docker-compose.yaml

```yaml
version: '3.8'
services:
  maccms:
    build: .
    ports:
      - "8080:8080"
    volumes:
      - ./config:/app/config
      - ./web/uploads:/app/web/uploads
      - ./web/template:/app/web/template
    depends_on:
      - mysql
      - redis
    environment:
      - DB_HOST=mysql
      - DB_PORT=3306
      - DB_NAME=maccms
      - DB_USER=maccms
      - DB_PASS=secure_password
      - REDIS_ADDR=redis:6379

  mysql:
    image: mysql:8.0
    volumes:
      - mysql_data:/var/lib/mysql
    environment:
      - MYSQL_DATABASE=maccms
      - MYSQL_USER=maccms
      - MYSQL_PASSWORD=secure_password
      - MYSQL_ROOT_PASSWORD=root_secure_password

  redis:
    image: redis:7-alpine
    volumes:
      - redis_data:/data

volumes:
  mysql_data:
  redis_data:
```

---

## 十、性能对比预期

| 指标 | PHP+ThinkPHP | Go+Fiber | 提升 |
|------|-------------|----------|------|
| QPS (简单页面) | ~2,000 | ~50,000+ | 25x |
| QPS (数据库查询) | ~500 | ~10,000+ | 20x |
| 内存占用 (空闲) | ~50MB | ~15MB | -70% |
| 内存占用 (并发) | ~500MB | ~100MB | -80% |
| 启动时间 | ~2s | ~100ms | 20x |
| 冷启动延迟 | 高 | 极低 | - |
| 并发连接 | ~1,000 | ~100,000+ | 100x |

---

## 十一、重写实施路线

### Phase 1: 基础框架 (2-3周)

- [ ] 项目骨架搭建 (Go模块/目录结构)
- [ ] 数据库连接 + GORM模型定义
- [ ] Fiber路由框架 + 中间件链
- [ ] 后台认证系统 (Session/JWT)
- [ ] 配置管理 (viper)
- [ ] 日志系统 (zap)
- [ ] 数据库迁移工具

### Phase 2: 后台核心 (3-4周)

- [ ] 系统设置 (80KB大文件拆解)
- [ ] 分类管理 (树形结构)
- [ ] 视频管理 (CRUD/批量/排序)
- [ ] 文章管理
- [ ] 漫画管理
- [ ] 演员/角色管理
- [ ] 用户管理
- [ ] 管理员管理 + 权限系统

### Phase 3: 前台展示 (2-3周)

- [ ] 模板引擎 (标签解析器 + 自定义函数)
- [ ] 首页渲染
- [ ] 视频分类/详情/播放/下载/搜索
- [ ] 文章分类/详情/阅读
- [ ] 漫画分类/详情/阅读
- [ ] 演员/角色列表/详情
- [ ] 专题页
- [ ] 留言板
- [ ] 用户中心

### Phase 4: 高级功能 (3-4周)

- [ ] 采集模块 (XML/JSON/HTML)
- [ ] 静态页面生成
- [ ] 缓存系统 (Redis + 文件)
- [ ] 搜索引擎集成 (Meilisearch)
- [ ] 弹幕系统 (WebSocket)
- [ ] URL推送 (百度/神马/搜狗)
- [ ] 数据分析/统计
- [ ] 定时任务系统

### Phase 5: 扩展功能 (2-3周)

- [ ] 插件系统
- [ ] AI内容生成模块
- [ ] 卡密/订单/支付系统
- [ ] 直播管理
- [ ] 聊天室
- [ ] 安全加固

### Phase 6: 测试与优化 (2周)

- [ ] 单元测试覆盖 (>70%)
- [ ] 集成测试
- [ ] 性能测试/压测
- [ ] 安全审计 (去除所有暗桩)
- [ ] 文档编写
- [ ] Docker化部署

**总预估工期: 14-19周 (约3.5-5个月)**

---

## 十二、关键注意事项

### 12.1 兼容性

1. **数据库兼容**: 保持 `mac_` 表前缀，支持现有数据库无缝迁移
2. **URL兼容**: 保持原有URL格式，支持SEO无变化
3. **模板兼容**: 提供标签转换工具，将PHP模板标签转换为Go模板
4. **API兼容**: 保持采集API接口格式 (XML/JSON) 不变

### 12.2 安全红线

1. **所有用户输入**: 参数化查询 + 输入验证 + 输出转义
2. **文件上传**: 白名单 + 大小限制 + 病毒扫描
3. **会话安全**: HttpOnly + Secure + SameSite Cookie
4. **CSRF**: 所有表单POST请求CSRF Token
5. **XSS**: 模板自动转义 + CSP头
6. **SQL注入**: 100%参数化查询
7. **路径遍历**: 所有文件操作路径校验
8. **无暗桩**: 代码100%开源可审计

---

*报告完毕。如需任何模块的详细设计文档或代码实现，请告知。*
