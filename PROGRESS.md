# 苹果CMS v10 → Go + Fiber 重写进度

> 项目: [GOcms](https://github.com/lullabyable/GOcms)
> 源项目: [maccms10](https://github.com/magicblack/maccms10)
> 开始时间: 2026-06-09
> 目标: 完整功能重写，去除所有暗桩后门

---

## 总体进度

```
Phase 1: 基础框架          [██████████] 100% ✅
Phase 2: 后台核心          [██████████] 100% ✅
Phase 3: 前台展示          [██████████] 100% ✅
Phase 4: 高级功能          [██████████] 100% ✅
Phase 5: 扩展功能          [░░░░░░░░░░] 0%   未开始
Phase 6: 测试与优化        [░░░░░░░░░░] 0%   未开始
```

---

## Phase 1: 基础框架（预计 2-3 周）

- [x] 项目骨架搭建（go mod / 目录结构）
- [x] 数据库连接 + GORM 模型定义（Vod/Art/Type/User/Admin/Comment/Gbook/Config/Actor/Role/Manga/Live/Danmaku）
- [x] Fiber 路由框架 + 中间件链（recover/cors/compress/logger）
- [x] 后台认证系统（Session / Cookie + 登录/登出/鉴权中间件）
- [x] 配置管理（viper / config.yaml）
- [x] 日志系统（zap + lumberjack 轮转）
- [x] 数据库迁移工具（embed SQL + 版本管理 + install.sql）
- [x] 单元测试框架（go test + SQLite 内存测试库）

## Phase 2: 后台核心（预计 3-4 周）

- [x] 系统设置（配置读取/保存/缓存清理）
- [x] 分类管理（树形结构 CRUD + 排序）
- [x] 视频管理（CRUD / 批量 / 排序 / 审核）
- [x] 文章管理（CRUD）
- [x] 漫画管理（CRUD / 审核）
- [x] 演员 / 角色管理（CRUD）
- [x] 用户管理 / 用户组管理
- [x] 管理员管理 + 权限系统
- [x] 评论管理 / 留言管理
- [x] 后台路由完整注册（/admin/* 全部接口）

## Phase 3: 前台展示（预计 2-3 周）

- [x] 模板引擎（标签解析器 + 自定义函数）
- [x] 分页器（Paginator）
- [x] 旁路缓存系统（PageCache / ListCache / DetailCache）
- [x] 首页渲染
- [x] 视频分类 / 详情 / 播放 / 下载 / 搜索
- [x] 文章分类 / 详情 / 阅读
- [x] 漫画分类 / 详情 / 搜索 / 筛选
- [x] 演员列表 / 详情 / 筛选
- [x] 角色列表 / 详情 / 筛选
- [x] 专题列表 / 详情
- [x] 留言板
- [x] 用户中心

## Phase 4: 高级功能（预计 3-4 周）

- [x] 采集模块（XML / JSON / HTML）
- [x] 缓存系统（Redis 可选扩展）
- [x] 搜索引擎集成（Meilisearch，可降级 MySQL）
- [x] 弹幕系统（WebSocket）
- [x] URL 推送（百度 / 神马 / 搜狗）
- [x] 数据分析 / 统计
- [x] 定时任务系统

## Phase 5: 扩展功能（预计 2-3 周）

- [ ] 插件系统
- [ ] AI 内容生成模块
- [ ] 卡密 / 订单 / 支付系统
- [ ] 直播管理
- [ ] 聊天室
- [ ] 安全加固

## Phase 6: 测试与优化（预计 2 周）

- [ ] 单元测试覆盖（>70%）
- [ ] 集成测试
- [ ] 性能测试 / 压测
- [ ] 安全审计（去除所有暗桩）
- [ ] 文档编写
- [ ] Docker 化部署

---

## 变更日志

### 2026-06-09 (续2)
- **Phase 3 完成：旁路缓存系统** ✅
  - `internal/cache/cache.go`: CacheManager 统一缓存管理器（文件/Redis双驱动）
  - `internal/cache/file_driver.go`: 文件缓存驱动（两级目录 + 原子写入）
  - `internal/cache/redis_driver.go`: Redis 缓存驱动（可选，连接失败自动降级）
  - `internal/cache/singleflight.go`: 缓存击穿防护
  - `internal/cache/layers.go`: 五层缓存（Page/List/Detail/Config/Search）
  - `internal/cache/middleware.go`: 页面缓存中间件
  - `internal/cache/invalidator.go`: 缓存失效协调器
- **Phase 4 完成：高级功能** ✅
  - 采集模块：`internal/service/collect/`（XML/JSON/HTML采集 + 播放源解析 + 分类映射）
  - 搜索引擎：`internal/service/search/`（Meilisearch + MySQL降级）
  - Provide API：`internal/handler/api/provide.go`（苹果CMS标准协议）
  - 弹幕系统：`internal/handler/frontend/danmaku.go`（WebSocket + 持久化）
  - URL推送：`internal/service/urlpush/pusher.go`（百度/神马/搜狗）
  - 数据分析：`internal/service/analytics/analytics.go`（PV/UV/趋势/来源）
  - 定时任务：`internal/service/scheduler/scheduler.go`（cron调度器）
  - 新增模型：danmaku.go, visit.go, task.go
  - 路由更新：注册所有新 API 和后台路由

### 2026-06-09 (续)
- **Phase 3 进度推进：前台展示 60% → 80%**
  - 模板标签解析器：`internal/template/parser.go`
    - 注释标签、PHP标签移除
    - include 标签 → Go template 引用
    - 函数调用标签 {:func($var)} → {{func .Var}}
    - 变量标签 {$vo.vod_name} → {{.vo.vod_name}}
    - 条件标签 {maccms:if} → {{if}}
    - 循环标签 {maccms:vod} → {{range}}
    - volist 标签、分页标签、常量标签
  - 分页器：`internal/template/paginator.go`
    - 兼容原 macCMS 分页样式
    - 支持首页/末页/上下页/页码范围
  - 漫画前台 Handler：`internal/handler/frontend/manga.go`
    - 分类页 / 详情页 / 筛选页 / 搜索
  - 演员前台 Handler：`internal/handler/frontend/actor.go`
    - 列表页 / 详情页 / 筛选页（按地区/性别/字母）
  - 角色前台 Handler：`internal/handler/frontend/role.go`
    - 列表页 / 详情页 / 筛选页
  - 专题前台 Handler：`internal/handler/frontend/topic.go`
    - 列表页 / 详情页（含关联视频）
  - 新增模型：`internal/model/role.go`、`internal/model/topic.go`
  - 路由更新：注册漫画/演员/角色/专题前台路由

### 2026-06-09
- 项目初始化
- 完成技术分析报告 (`maccms-report.md`)
- 完成补充实现方案 (`maccms-supplement.md`)
- 配置 Go 环境（Go 1.26.4 + goproxy.cn）
- 配置 GitHub SSH
- **Phase 1：基础框架搭建** ✅
  - 项目骨架（cmd/server + internal/* + web/* + config/）
  - GORM 模型：Vod / Art / Type / User / Admin / Group / Comment / Gbook / Config / Actor / Role / Manga / Live / Danmaku
  - Fiber 路由框架 + 中间件（recover / cors / compress / requestid / logger）
  - 配置管理：viper + config.yaml
  - 日志系统：zap + lumberjack 轮转
  - 数据库连接：MySQL + 连接池配置
  - Session 管理：MemStore 默认 + RedisStore 可选
  - 后台认证中间件：AdminAuth + 登录/登出/Regenerate
  - 数据库迁移：embed SQL + 版本号管理 + install.sql
  - 单元测试：go test + SQLite 内存库
  - Makefile + Dockerfile
  - 编译通过 ✅ 测试通过 ✅
- **Phase 2：后台核心 CRUD** ✅
  - Dashboard / Type / Vod / Art / Manga / Actor / Role / User / Group / Admin / System / Comment / Gbook
  - 后台路由完整注册（49 个接口）
  - 编译通过 ✅
- 代码重命名 maccms → gocms

---

## 技术决策记录

| 决策 | 方案 | 原因 |
|------|------|------|
| 缓存驱动 | 文件缓存默认，Redis 可选 | 轻启动，无外部依赖 |
| 静态HTML | 旁路缓存（Lazy） | 用户请求时自动生成，无需后台操作 |
| 缓存击穿 | singleflight | 多请求同一 URL 只查一次 DB |
| 模板引擎 | Go html/template + 标签解析器 | 兼容原 macCMS 模板标签 |
| URL 路由 | 数据驱动的 Catch-All 引擎 | 原系统的 URL 规则是可配置的 |
| 搜索 | Meilisearch 可选，降级 MySQL FULLTEXT | 轻启动可不装 Meilisearch |
| ORM | GORM v2 | Go 最流行，社区活跃 |
