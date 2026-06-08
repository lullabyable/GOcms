# GOcms

基于 Go + Fiber v2 + GORM v2 的内容管理系统，从苹果CMS v10 完整重写。

## 特性

- 🚀 **高性能**：Go + Fiber 异步框架，支持高并发
- 📦 **模块化**：清晰的分层架构（handler → service → model）
- 🔌 **插件系统**：可扩展的插件接口
- 🤖 **AI 集成**：OpenAI 兼容的内容生成
- 💬 **实时通信**：WebSocket 弹幕 + 聊天室
- 📊 **数据分析**：PV/UV/IP 统计、趋势分析、来源分析
- ⏰ **定时任务**：基于 cron 的任务调度器
- 🔒 **安全加固**：限流、CSRF、XSS、SQL注入防护
- 💳 **支付系统**：订单管理 + 卡密系统

## 快速开始

### 环境要求

- Go 1.22+
- MySQL 5.7+ 或 SQLite

### 安装

```bash
git clone git@github.com:lullabyable/GOcms.git
cd GOcms
go mod tidy
go build -o gocms ./cmd/server
```

### 配置

编辑 `config/config.yaml`：

```yaml
server:
  host: "0.0.0.0"
  port: 8080

database:
  host: "127.0.0.1"
  port: 3306
  user: "root"
  password: "your_password"
  database: "gocms"

session:
  secret: "your-random-secret-key"
```

### 运行

```bash
./gocms
# 或
go run ./cmd/server
```

访问 `http://localhost:8080` 查看前台，`http://localhost:8080/admin` 进入后台。

## 项目结构

```
GOcms/
├── cmd/server/          # 入口
├── config/              # 配置文件
├── internal/
│   ├── config/          # 配置加载
│   ├── database/        # 数据库连接 + 迁移
│   ├── handler/
│   │   ├── admin/       # 后台 API
│   │   ├── api/         # 公开 API
│   │   └── frontend/    # 前台页面 + WebSocket
│   ├── middleware/       # 中间件（认证/安全/日志）
│   ├── model/           # 数据模型
│   ├── router/          # 路由注册
│   ├── service/
│   │   ├── aicontent/   # AI 内容生成
│   │   ├── analytics/   # 数据分析
│   │   ├── chat/        # 聊天服务
│   │   ├── collect/     # 采集引擎
│   │   ├── payment/     # 支付服务
│   │   ├── plugin/      # 插件系统
│   │   ├── scheduler/   # 定时任务
│   │   ├── search/      # 搜索引擎
│   │   └── urlpush/     # URL 推送
│   ├── session/         # 会话管理
│   ├── template/        # 模板引擎
│   └── testutil/        # 测试工具
├── Dockerfile
├── Makefile
└── PROGRESS.md
```

## API 概览

### 后台 API（需认证）

| 模块 | 接口 | 说明 |
|------|------|------|
| 仪表盘 | `GET /admin/dashboard` | 统计数据 |
| 视频 | `GET/POST /admin/vod/*` | CRUD + 批量 |
| 文章 | `GET/POST /admin/art/*` | CRUD |
| 分类 | `GET/POST /admin/type/*` | 树形管理 |
| 用户 | `GET/POST /admin/user/*` | 管理 |
| 弹幕 | `GET/POST /admin/danmaku/*` | 管理 |
| URL推送 | `GET/POST /admin/urlsend/*` | 百度/神马/搜狗 |
| 数据分析 | `GET /admin/analytics/*` | 仪表盘/趋势/来源 |
| 定时任务 | `GET/POST /admin/timming/*` | CRUD + 触发 |
| 插件 | `GET/POST /admin/plugin/*` | 安装/配置 |
| AI | `POST /admin/ai/*` | 内容生成 |
| 订单 | `GET/POST /admin/order/*` | 订单/卡密 |
| 直播 | `GET/POST /admin/live/*` | CRUD |
| 聊天 | `GET/POST /admin/chat/*` | 房间/记录 |

### 前台 API

| 接口 | 说明 |
|------|------|
| `GET /voddetail/:id` | 视频详情 |
| `GET /vodsearch` | 搜索 |
| `WS /ws/danmaku/:vod_id` | 弹幕 WebSocket |
| `WS /ws/chat` | 聊天室 WebSocket |
| `GET /api/danmaku/:vod_id/history` | 弹幕历史 |

## 定时任务

内置任务：
- `aggregate_daily` — 每日 02:00 汇总访问统计
- `cache_clean` — 每日 03:00 清理缓存
- `db_optimize` — 每周日 04:00 优化数据库
- `url_push` — 每日 06:00 自动 URL 推送

## Docker 部署

```bash
docker build -t gocms .
docker run -p 8080:8080 -v ./config:/app/config gocms
```

## 开发

```bash
# 运行测试
make test

# 构建
make build

# 交叉编译
make build-linux

# Docker 构建
make docker
```

## 技术栈

- **框架**: [Fiber v2](https://github.com/gofiber/fiber)
- **ORM**: [GORM v2](https://gorm.io)
- **配置**: [Viper](https://github.com/spf13/viper)
- **日志**: [Zap](https://go.uber.org/zap)
- **WebSocket**: [Fiber WebSocket](https://github.com/gofiber/websocket)
- **定时任务**: [robfig/cron](https://github.com/robfig/cron)

## License

MIT
