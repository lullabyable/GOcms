# GOcms 后台 SPA 重构 — 参照系分析

> 基于 maccms10 (magicblack/maccms10) 原版后台 vs GOcms 现有接口
> 生成时间: 2026-06-15

---

## 一、原版 maccms10 后台菜单结构 (layui 框架)

原版采用 layui + iframe 多 Tab 架构，菜单由后端动态生成，结构如下：

### 顶部一级菜单 → 左侧二级菜单

| # | 一级菜单 | 二级菜单项 |
|---|---------|-----------|
| 1 | **首页** | 后台首页(Welcome) |
| 2 | **视频** | 视频列表、视频管理(添加/编辑)、视频分类、视频采集 |
| 3 | **文章** | 文章列表、文章管理(添加/编辑)、文章分类 |
| 4 | **漫画** | 漫画列表、漫画管理(添加/编辑)、漫画分类 |
| 5 | **演员** | 演员列表、演员管理 |
| 6 | **角色** | 角色列表、角色管理 |
| 7 | **专题** | 专题列表、专题管理 |
| 8 | **用户** | 用户列表、用户组管理 |
| 9 | **管理员** | 管理员列表、权限管理 |
| 10 | **留言** | 留言管理 |
| 11 | **评论** | 评论管理 |
| 12 | **采集** | 采集资源站、自定义采集、弹幕管理 |
| 13 | **推广** | URL推送(百度/神马/搜狗)、Sitemap |
| 14 | **数据** | 数据库备份/恢复/优化/SQL执行、数据替换 |
| 15 | **模板** | 模板管理 |
| 16 | **运营** | 定时任务、访问统计、操作日志 |
| 17 | **扩展** | 插件管理、AI内容、卡密/订单、直播管理、聊天室 |
| 18 | **系统** | 系统设置、安全配置、缓存管理、友情链接、域名管理 |

---

## 二、接口对照表 (source of truth)

### 格式说明
- ✅ 已实现 (handler 存在 + 路由注册 + 可用)
- ⚠️ 部分实现 (handler 存在但缺字段/功能不完整)
- ❌ 缺失 (无 handler 或无路由)
- 🔶 仅有页面渲染(template)，无 JSON API

### 2.1 视频模块 (Vod)

| 原版路径 | 功能 | GOcms handler | 状态 | 备注 |
|---------|------|--------------|------|------|
| /admin/vod/index | 视频列表 | vodH.List | ⚠️ | 需检查筛选/分页/排序字段完整性 |
| /admin/vod/add | 添加视频 | vodH.Save | ⚠️ | 缺少多集/播放源/下载源编辑 |
| /admin/vod/edit?id=X | 编辑视频 | vodH.Detail + vodH.Save | ⚠️ | 缺字段: vod_play_from, vod_play_server, vod_down_from, vod_down_server |
| /admin/vod/del | 删除视频 | vodH.Delete | ✅ | |
| /admin/vod/audit | 审核 | vodH.Audit | ✅ | |
| /admin/vod/batch | 批量操作 | vodH.Batch | ⚠️ | 缺批量替换、批量移动分类 |
| /admin/vod/data | 数据管理(导入导出) | ❌ | ❌ | 缺 CSV/XLSX 导入导出 |

### 2.2 文章模块 (Art)

| 原版路径 | 功能 | GOcms handler | 状态 | 备注 |
|---------|------|--------------|------|------|
| /admin/art/index | 文章列表 | artH.List | ⚠️ | |
| /admin/art/add | 添加文章 | artH.Save | ⚠️ | |
| /admin/art/edit?id=X | 编辑文章 | artH.Detail + artH.Save | ⚠️ | 缺字段检查 |
| /admin/art/del | 删除文章 | artH.Delete | ✅ | |

### 2.3 漫画模块 (Manga)

| 原版路径 | 功能 | GOcms handler | 状态 | 备注 |
|---------|------|--------------|------|------|
| /admin/manga/index | 漫画列表 | mangaH.List | ⚠️ | |
| /admin/manga/edit?id=X | 编辑漫画 | mangaH.Detail + mangaH.Save | ⚠️ | 缺章节管理 |
| /admin/manga/audit | 审核 | mangaH.Audit | ✅ | |

### 2.4 演员/角色 (Actor/Role)

| 原版路径 | 功能 | GOcms handler | 状态 | 备注 |
|---------|------|--------------|------|------|
| /admin/actor/index | 演员列表 | actorH.List | ✅ | |
| /admin/actor/edit?id=X | 编辑演员 | actorH.Detail + actorH.Save | ⚠️ | |
| /admin/role/index | 角色列表 | roleH.List | ✅ | |
| /admin/role/edit?id=X | 编辑角色 | roleH.Detail + roleH.Save | ⚠️ | |

### 2.5 专题 (Topic)

| 原版路径 | 功能 | GOcms handler | 状态 | 备注 |
|---------|------|--------------|------|------|
| /admin/topic/index | 专题列表 | topicH.List | ✅ | |
| /admin/topic/edit?id=X | 编辑专题 | topicH.Detail + topicH.Save | ⚠️ | |

### 2.6 分类 (Type)

| 原版路径 | 功能 | GOcms handler | 状态 | 备注 |
|---------|------|--------------|------|------|
| /admin/type/index | 分类列表 | typeH.List / typeH.Tree | ✅ | |
| /admin/type/add | 添加分类 | typeH.Save | ✅ | |
| /admin/type/edit?id=X | 编辑分类 | typeH.Detail + typeH.Save | ✅ | |
| /admin/type/sort | 排序 | typeH.Sort | ✅ | |

### 2.7 用户 (User)

| 原版路径 | 功能 | GOcms handler | 状态 | 备注 |
|---------|------|--------------|------|------|
| /admin/user/index | 用户列表 | userH.List | ⚠️ | |
| /admin/user/edit?id=X | 编辑用户 | userH.Save | ⚠️ | 缺 Detail 接口 |
| /admin/user/group | 用户组 | groupH.List + groupH.Save | ✅ | |

### 2.8 管理员 (Admin)

| 原版路径 | 功能 | GOcms handler | 状态 | 备注 |
|---------|------|--------------|------|------|
| /admin/admin/index | 管理员列表 | adminH.List | ✅ | |
| /admin/admin/edit?id=X | 编辑管理员 | adminH.Save | ⚠️ | 缺权限编辑 |
| /admin/adminaudit | 管理员操作日志 | ❌ | ❌ | 缺失 |

### 2.9 评论/留言 (Comment/Gbook)

| 原版路径 | 功能 | GOcms handler | 状态 | 备注 |
|---------|------|--------------|------|------|
| /admin/comment/index | 评论列表 | commentH.List | ✅ | |
| /admin/comment/audit | 审核 | commentH.Audit | ✅ | |
| /admin/gbook/index | 留言列表 | gbookH.List | ✅ | |
| /admin/gbook/reply | 回复 | gbookH.Reply | ✅ | |

### 2.10 采集 (Collect)

| 原版路径 | 功能 | GOcms handler | 状态 | 备注 |
|---------|------|--------------|------|------|
| /admin/collect/index | 采集资源站 | collectH.TestConnection + collectH.StartCollect | ⚠️ | 缺资源站列表管理 |
| /admin/cj/index | 自定义采集 | ❌ | ❌ | 完全缺失 |
| /admin/collect/danmaku | 弹幕管理 | danmakuH.AdminList + danmakuH.AdminDelete | ✅ | |

### 2.11 推广 (URL Push)

| 原版路径 | 功能 | GOcms handler | 状态 | 备注 |
|---------|------|--------------|------|------|
| /admin/urlsend/index | URL推送 | urlSendH.Config + urlSendH.PushURLs | ✅ | |
| /admin/urlsend/sitemap | Sitemap | urlSendH.GenerateSitemap | ✅ | |

### 2.12 数据库 (Database)

| 原版路径 | 功能 | GOcms handler | 状态 | 备注 |
|---------|------|--------------|------|------|
| /admin/database/index | 数据库管理 | dbH.List | ✅ | |
| /admin/database/optimize | 优化 | dbH.Optimize | ✅ | |
| /admin/database/backup | 备份 | dbH.Backup | ✅ | |
| /admin/database/restore | 恢复 | dbH.Restore | ✅ | |
| /admin/database/sql | SQL执行 | dbH.SQL | ✅ | |
| /admin/datareplace | 数据替换 | dataReplaceH.Execute | ✅ | |

### 2.13 模板 (Template)

| 原版路径 | 功能 | GOcms handler | 状态 | 备注 |
|---------|------|--------------|------|------|
| /admin/template/index | 模板列表 | tplH.List | ✅ | |
| /admin/template/edit | 编辑模板 | tplH.Read + tplH.Save | ✅ | |

### 2.14 运营 (Analytics/Timming/Plog)

| 原版路径 | 功能 | GOcms handler | 状态 | 备注 |
|---------|------|--------------|------|------|
| /admin/analytics/index | 访问统计 | analyticsH.Dashboard + Trend + TopContent | ✅ | |
| /admin/timming/index | 定时任务 | timmingH.List + Create + Update + Delete + Toggle | ✅ | |
| /admin/plog/index | 操作日志 | plogH.List | ✅ | |
| /admin/visit | 访问日志 | ❌ | ❌ | 缺失 (visit 模块) |

### 2.15 扩展 (Plugin/AI/Order/Live/Chat)

| 原版路径 | 功能 | GOcms handler | 状态 | 备注 |
|---------|------|--------------|------|------|
| /admin/addon/index | 插件管理 | pluginH.List + Install + Enable + Disable | ✅ | |
| /admin/ai/... | AI内容生成 | aiH.Generate + Title + Summary + Tags | ✅ | |
| /admin/order/index | 订单管理 | orderH.List + Pay + Cancel | ⚠️ | 缺退款、详情 |
| /admin/card/index | 卡密管理 | orderH.CardList + GenerateCards | ⚠️ | 缺批量导入/导出 |
| /admin/live/index | 直播管理 | liveH.List + Detail + Save + Delete + Toggle | ✅ | |
| /admin/chatroom/index | 聊天室 | chatH.RoomList + RoomCreate + RoomUpdate + RoomDelete | ✅ | |

### 2.16 系统 (System/Link/Safety)

| 原版路径 | 功能 | GOcms handler | 状态 | 备注 |
|---------|------|--------------|------|------|
| /admin/system/config | 系统设置 | systemH.GetConfig + systemH.SaveConfig | ⚠️ | 原版有 10+ 个 tab 页(站点/播放器/邮箱/缓存/安全/API...)，GOcms 可能只有基础配置 |
| /admin/system/cache | 缓存管理 | systemH.CacheClear | ✅ | |
| /admin/link/index | 友情链接 | linkH.List + linkH.Save | ✅ | |
| /admin/safety/index | 安全配置 | ❌ | ❌ | 缺失: IP封禁、防盗链、敏感词、验证码配置 |
| /admin/domain/index | 域名管理 | ❌ | ❌ | 缺失: 多域名绑定 |
| /admin/annex/index | 附件管理 | uploadH.File + uploadH.Image | ⚠️ | 缺附件列表、批量删除 |
| /admin/images/index | 图片管理 | ❌ | ❌ | 缺失 |

### 2.17 完全缺失的模块

| 模块 | 原版 controller | 说明 | 优先级 |
|------|----------------|------|--------|
| **安全配置** | Safety.php | IP封禁、防盗链、敏感词过滤、验证码 | 🔴 高 |
| **自定义采集** | Cj.php | 可视化采集规则配置 | 🟡 中 |
| **管理员审计** | Adminaudit.php | 管理员操作审计日志 | 🟡 中 |
| **附件管理** | Annex.php | 上传文件列表、批量清理 | 🟡 中 |
| **图片管理** | Images.php | 图片资源管理 | 🟢 低 |
| **域名管理** | Domain.php | 多域名绑定 | 🟢 低 |
| **访问日志** | Visit (在 model 中) | 详细访问记录 | 🟡 中 |
| **批量播放器** | BatchPlayer.php | 播放器批量配置 | 🟢 低 |
| **签到里程碑** | SignMilestone.php | 用户签到系统 | 🟢 低 |
| **资源中心** | ResourceHub.php | 资源聚合 | 🟢 低 |
| **更新管理** | Update.php | 系统在线更新 | 🟡 中 |
| **静态生成** | Make.php | 静态页面生成 | 🟡 中 |
| **提现管理** | Cash.php | 用户提现审核 | 🟢 低 |

---

## 三、系统设置字段对照 (关键!)

原版 maccms10 系统设置 `/admin/system/config` 包含以下 Tab：

| Tab | 字段组 | GOcms 状态 |
|-----|--------|-----------|
| **站点设置** | site_name, site_url, site_keywords, site_description, site_icp, site_tj(统计代码), template_dir, mob_template_dir | ⚠️ 部分 |
| **会员设置** | user_status, user_reg, user_email_check, user_yzm | ❌ 缺失 |
| **播放器设置** | player_* 系列 (播放器列表/解析接口/排序) | ❌ 缺失 |
| **邮箱设置** | email_host, email_port, email_username, email_password, email_from, email_secure | ❌ 缺失 |
| **缓存设置** | cache_open, cache_type, cache_time, cache_* | ⚠️ 部分 |
| **安全设置** | safe_* (验证码/防盗链/IP限制) | ❌ 缺失 |
| **API设置** | api_open, api_* | ❌ 缺失 |
| **SEO设置** | seo_* (各页面 title/description 模板) | ❌ 缺失 |
| **性能设置** | pagesize, makesize, search_* | ⚠️ 部分 |

---

## 四、执行计划

### Phase A: 后端 JSON API 固化 (子代理 1)
**目标**: 所有 `/admin/*` 路由返回统一 JSON 格式

1. 定义统一响应格式: `{code: 0/1, msg: "", data: {}, count: 0}`
2. 将现有 template 渲染的 page handler 改为 JSON API
3. 补全缺失接口的 stub (返回空数据但格式正确)
4. 补全系统设置的所有字段组

### Phase B: 原版静态资源提取 (子代理 2)
**目标**: 从 maccms10 提取 layui + CSS + JS

1. 克隆 maccms10 仓库
2. 提取 `/public/static/` 下的 layui 框架资源
3. 提取 `/application/admin/view/` 下的 HTML 模板
4. 整理到 `web/static/admin-spa/`

### Phase C: SPA 前端构建 (子代理 3)
**目标**: 用原版 HTML + JS fetch 替换 PHP 变量

1. 主框架 (index.html): 侧边栏菜单 + Tab 容器
2. 每个页面 HTML 中的 `{:url('xxx')}` → `fetch('/admin/api/xxx')`
3. 表单提交改为 AJAX JSON
4. layui 表格改为对接 JSON API 分页

### Phase D: 路由切换 (主进程)
**目标**: `/admin/*` 根据 Accept 头分流

- `Accept: application/json` → JSON API
- 否则 → SPA 静态页面
- 或统一走 `/api/admin/*` + `/admin/*` (SPA)

---

## 五、推荐子代理分工

| 子代理 | 任务 | 预计工作量 |
|--------|------|-----------|
| **agent-api** | Phase A: 后端 JSON API 固化 + 补全 | 高 (30+ 个接口) |
| **agent-extract** | Phase B: 原版资源提取 + 目录整理 | 低 |
| **agent-spa** | Phase C: SPA 前端复刻 | 高 (20+ 个页面) |
