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

- Go 1.25+
- MySQL 5.7+

### 安装

```bash
git clone git@github.com:lullabyable/GOcms.git
cd GOcms
go mod tidy
go build -o gocms ./cmd/server
```

### 运行

```bash
./gocms
```

首次启动后访问 `http://localhost:8080`，自动跳转安装向导：

1. 填写 MySQL 连接信息，点击「测试连接」
2. 设置管理员账号密码
3. 点击「开始安装」，自动建表、写入配置、生成 `install.lock`
4. 安装完成自动跳转首页

### 手动配置（可选）

编辑 `config/config.yaml`：

```yaml
server:
  host: "0.0.0.0"
  port: 8080

database:
  driver: "mysql"
  host: "127.0.0.1"
  port: 3306
  user: "root"
  password: "your_password"
  database: "gocms"

template:
  dir: "./web/templates"
  theme: "default"

session:
  secret: "your-random-secret-key"
```

## 项目结构

```
GOcms/
├── cmd/server/          # 入口
├── config/              # 配置文件
├── internal/
│   ├── config/          # 配置加载
│   ├── database/        # 数据库连接 + 迁移 (MySQL)
│   ├── handler/
│   │   ├── admin/       # 后台 API + 安装向导
│   │   ├── api/         # 公开 API
│   │   └── frontend/    # 前台页面 + WebSocket
│   ├── middleware/       # 中间件（认证/安全/日志）
│   ├── model/           # 数据模型
│   ├── router/          # 路由注册
│   ├── service/         # 业务服务
│   ├── session/         # 会话管理
│   ├── template/        # 模板引擎（多主题支持）
│   └── testutil/        # 测试工具
├── web/
│   ├── static/          # 静态资源
│   ├── templates/       # 前台模板
│   │   └── default/     # 默认主题
│   │       ├── *.html   # 页面模板
│   │       └── partials/# 公共组件 (header/footer)
│   └── uploads/         # 上传文件
├── runtime/             # 运行时（缓存/日志）
├── Dockerfile
├── Makefile
└── README.md
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
