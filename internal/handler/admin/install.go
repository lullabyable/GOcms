package admin

import (
	"fmt"
	"os"

	"github.com/gofiber/fiber/v2"
	"gopkg.in/yaml.v3"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"gocms/internal/database"
)

type InstallHandler struct {
	installed bool
}

func NewInstallHandler() *InstallHandler {
	return &InstallHandler{}
}

// CheckInstall 检查是否已安装（中间件用）
func (h *InstallHandler) CheckInstall(c *fiber.Ctx) error {
	if h.installed {
		return c.Next()
	}
	// 未安装则跳转到安装页面
	if c.Path() != "/install" && c.Path() != "/install/submit" {
		return c.Redirect("/install")
	}
	return c.Next()
}

// SetInstalled 标记已安装
func (h *InstallHandler) SetInstalled(v bool) {
	h.installed = v
}

// IsInstalled 是否已安装
func (h *InstallHandler) IsInstalled() bool {
	return h.installed
}

// Page 安装页面
func (h *InstallHandler) Page(c *fiber.Ctx) error {
	if h.installed {
		return c.JSON(fiber.Map{"code": 1, "msg": "已安装"})
	}

	html := `<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>GoCMS 安装向导</title>
<style>
*{margin:0;padding:0;box-sizing:border-box}
body{font-family:-apple-system,BlinkMacSystemFont,"Segoe UI",Roboto,sans-serif;background:#f0f2f5;color:#333;min-height:100vh;display:flex;align-items:center;justify-content:center}
.install-card{background:#fff;border-radius:12px;box-shadow:0 4px 24px rgba(0,0,0,.08);width:100%;max-width:520px;padding:40px}
h1{text-align:center;font-size:24px;margin-bottom:8px;color:#1a1a1a}
.subtitle{text-align:center;color:#888;margin-bottom:32px;font-size:14px}
.form-group{margin-bottom:20px}
label{display:block;font-weight:600;margin-bottom:6px;font-size:14px;color:#555}
input,select{width:100%;padding:10px 14px;border:1px solid #ddd;border-radius:8px;font-size:14px;transition:border .2s}
input:focus,select:focus{outline:none;border-color:#4f46e5;box-shadow:0 0 0 3px rgba(79,70,229,.1)}
.row{display:flex;gap:12px}
.row .form-group{flex:1}
.btn{width:100%;padding:12px;border:none;border-radius:8px;font-size:16px;font-weight:600;cursor:pointer;transition:all .2s}
.btn-primary{background:#4f46e5;color:#fff}
.btn-primary:hover{background:#4338ca}
.btn-primary:disabled{background:#a5a5a5;cursor:not-allowed}
.msg{padding:12px;border-radius:8px;margin-bottom:16px;font-size:14px;display:none}
.msg.error{display:block;background:#fef2f2;color:#dc2626;border:1px solid #fecaca}
.msg.success{display:block;background:#f0fdf4;color:#16a34a;border:1px solid #bbf7d0}
.db-test{margin-bottom:20px}
.section-title{font-size:16px;font-weight:700;margin:24px 0 16px;padding-top:16px;border-top:1px solid #eee}
</style>
</head>
<body>
<div class="install-card">
<h1>🎬 GoCMS 安装向导</h1>
<p class="subtitle">请配置数据库连接和管理员账号</p>

<div id="msg" class="msg"></div>

<form id="installForm">
<div class="section-title">数据库配置</div>

<div class="form-group">
<label>数据库类型</label>
<select id="driver" name="driver" onchange="toggleDriver()">
<option value="mysql">MySQL</option>
<option value="sqlite">SQLite (无需安装)</option>
</select>
</div>

<div id="mysqlFields">
<div class="form-group">
<label>数据库地址</label>
<input type="text" id="host" name="host" value="127.0.0.1" placeholder="127.0.0.1">
</div>
<div class="row">
<div class="form-group">
<label>端口</label>
<input type="number" id="port" name="port" value="3306">
</div>
<div class="form-group">
<label>数据库名</label>
<input type="text" id="database" name="database" value="gocms">
</div>
</div>
<div class="row">
<div class="form-group">
<label>用户名</label>
<input type="text" id="user" name="user" value="root">
</div>
<div class="form-group">
<label>密码</label>
<input type="password" id="password" name="password" placeholder="数据库密码">
</div>
</div>
</div>

<div class="form-group">
<button type="button" class="btn btn-primary" style="background:#6366f1;margin-bottom:8px" onclick="testDB()">测试连接</button>
</div>

<div class="section-title">管理员账号</div>

<div class="form-group">
<label>管理员用户名</label>
<input type="text" id="admin_user" name="admin_user" value="admin" required>
</div>
<div class="row">
<div class="form-group">
<label>密码</label>
<input type="password" id="admin_pwd" name="admin_pwd" placeholder="管理员密码" required>
</div>
<div class="form-group">
<label>确认密码</label>
<input type="password" id="admin_pwd2" name="admin_pwd2" placeholder="再次输入密码" required>
</div>
</div>

<button type="submit" class="btn btn-primary">开始安装</button>
</form>
</div>

<script>
function toggleDriver(){
var d=document.getElementById('driver').value;
document.getElementById('mysqlFields').style.display=d==='mysql'?'block':'none';
}
function showMsg(type,text){
var m=document.getElementById('msg');
m.className='msg '+type;m.textContent=text;m.style.display='block';
}
function testDB(){
var driver=document.getElementById('driver').value;
var data={driver:driver};
if(driver==='mysql'){
data.host=document.getElementById('host').value;
data.port=document.getElementById('port').value;
data.user=document.getElementById('user').value;
data.password=document.getElementById('password').value;
data.database=document.getElementById('database').value;
}
fetch('/install/test-db',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify(data)})
.then(r=>r.json()).then(d=>{
if(d.code===1)showMsg('success','✅ 数据库连接成功！');
else showMsg('error','❌ '+d.msg);
}).catch(e=>showMsg('error','请求失败: '+e));
}
document.getElementById('installForm').onsubmit=function(e){
e.preventDefault();
var pwd=document.getElementById('admin_pwd').value;
var pwd2=document.getElementById('admin_pwd2').value;
if(pwd!==pwd2){showMsg('error','两次密码不一致');return;}
if(pwd.length<6){showMsg('error','密码至少6位');return;}
var driver=document.getElementById('driver').value;
var data={driver:driver,admin_user:document.getElementById('admin_user').value,admin_pwd:pwd};
if(driver==='mysql'){
data.host=document.getElementById('host').value;
data.port=document.getElementById('port').value;
data.user=document.getElementById('user').value;
data.password=document.getElementById('password').value;
data.database=document.getElementById('database').value;
}
fetch('/install/submit',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify(data)})
.then(r=>r.json()).then(d=>{
if(d.code===1){showMsg('success','✅ 安装成功！3秒后跳转...');setTimeout(()=>location.href='/',3000);}
else showMsg('error','❌ '+d.msg);
}).catch(e=>showMsg('error','请求失败: '+e));
};
</script>
</body>
</html>`

	c.Set("Content-Type", "text/html; charset=utf-8")
	return c.SendString(html)
}

// TestDB 测试数据库连接
func (h *InstallHandler) TestDB(c *fiber.Ctx) error {
	var req struct {
		Driver   string `json:"driver"`
		Host     string `json:"host"`
		Port     string `json:"port"`
		User     string `json:"user"`
		Password string `json:"password"`
		Database string `json:"database"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.JSON(fiber.Map{"code": 0, "msg": "参数错误"})
	}

	var dialector gorm.Dialector
	switch req.Driver {
	case "sqlite":
		dialector = sqlite.Open("./runtime/gocms.db")
	case "mysql":
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			req.User, req.Password, req.Host, req.Port, req.Database)
		dialector = mysql.Open(dsn)
	default:
		return c.JSON(fiber.Map{"code": 0, "msg": "不支持的数据库类型"})
	}

	db, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		return c.JSON(fiber.Map{"code": 0, "msg": fmt.Sprintf("连接失败: %v", err)})
	}

	sqlDB, _ := db.DB()
	if err := sqlDB.Ping(); err != nil {
		return c.JSON(fiber.Map{"code": 0, "msg": fmt.Sprintf("Ping 失败: %v", err)})
	}
	sqlDB.Close()

	return c.JSON(fiber.Map{"code": 1, "msg": "连接成功"})
}

// Submit 执行安装
func (h *InstallHandler) Submit(c *fiber.Ctx) error {
	if h.installed {
		return c.JSON(fiber.Map{"code": 0, "msg": "已安装，如需重新安装请删除 config/config.yaml"})
	}

	var req struct {
		Driver    string `json:"driver"`
		Host      string `json:"host"`
		Port      string `json:"port"`
		User      string `json:"user"`
		Password  string `json:"password"`
		Database  string `json:"database"`
		AdminUser string `json:"admin_user"`
		AdminPwd  string `json:"admin_pwd"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.JSON(fiber.Map{"code": 0, "msg": "参数错误"})
	}

	// 校验
	if req.AdminUser == "" || req.AdminPwd == "" {
		return c.JSON(fiber.Map{"code": 0, "msg": "管理员账号密码不能为空"})
	}
	if len(req.AdminPwd) < 6 {
		return c.JSON(fiber.Map{"code": 0, "msg": "密码至少6位"})
	}

	// 1. 连接数据库
	var dialector gorm.Dialector
	switch req.Driver {
	case "sqlite":
		dialector = sqlite.Open("./runtime/gocms.db")
	case "mysql":
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			req.User, req.Password, req.Host, req.Port, req.Database)
		dialector = mysql.Open(dsn)
	default:
		return c.JSON(fiber.Map{"code": 0, "msg": "不支持的数据库类型"})
	}

	db, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		return c.JSON(fiber.Map{"code": 0, "msg": fmt.Sprintf("数据库连接失败: %v", err)})
	}

	// 2. 执行迁移
	migrator := database.NewMigrator(db)
	if err := migrator.Migrate(); err != nil {
		return c.JSON(fiber.Map{"code": 0, "msg": fmt.Sprintf("数据库迁移失败: %v", err)})
	}

	// 3. 创建管理员
	adminPwd := req.AdminPwd
	// 简单 hash（生产环境应用 bcrypt）
	hashedPwd := fmt.Sprintf("%x", []byte(adminPwd))

	result := db.Exec("INSERT INTO mac_admin (admin_name, admin_pwd, admin_status, admin_role) VALUES (?, ?, 1, 1) ON DUPLICATE KEY UPDATE admin_pwd = VALUES(admin_pwd)",
		req.AdminUser, hashedPwd)
	if result.Error != nil {
		// SQLite 不支持 ON DUPLICATE KEY，用 INSERT OR REPLACE
		result = db.Exec("INSERT OR REPLACE INTO mac_admin (admin_name, admin_pwd, admin_status, admin_role) VALUES (?, ?, 1, 1)",
			req.AdminUser, hashedPwd)
		if result.Error != nil {
			return c.JSON(fiber.Map{"code": 0, "msg": fmt.Sprintf("创建管理员失败: %v", result.Error)})
		}
	}

	// 4. 写入配置文件
	cfgMap := map[string]interface{}{
		"server": map[string]interface{}{
			"host":          "0.0.0.0",
			"port":          8080,
			"read_timeout":  30,
			"write_timeout": 30,
		},
		"database": map[string]interface{}{
			"driver":        req.Driver,
			"host":          req.Host,
			"port":          toInt(req.Port),
			"user":          req.User,
			"password":      req.Password,
			"database":      req.Database,
			"charset":       "utf8mb4",
			"max_open_conns": 100,
			"max_idle_conns": 10,
			"max_lifetime":  3600,
		},
		"cache": map[string]interface{}{
			"type":      "file",
			"flag":      "gocms",
			"core":      1,
			"time":      3600,
			"page":      1,
			"time_page": 3600,
			"file_dir":  "./runtime/cache",
		},
		"session": map[string]interface{}{
			"type":     "cookie",
			"name":     "gocms_session",
			"max_age":  86400,
			"secret":   "gocms-" + randomStr(16),
		},
		"log": map[string]interface{}{
			"level":       "info",
			"file":        "./runtime/logs/gocms.log",
			"max_size":    100,
			"max_backups": 5,
			"max_age":     30,
		},
		"upload": map[string]interface{}{
			"dir":        "./web/uploads",
			"max_size":   10,
			"allowed_ext": ".jpg,.jpeg,.png,.gif,.webp,.mp4,.mp3",
		},
		"template": map[string]interface{}{
			"dir":   "./web/templates",
			"theme": "default",
		},
	}

	yamlData, err := yaml.Marshal(cfgMap)
	if err != nil {
		return c.JSON(fiber.Map{"code": 0, "msg": fmt.Sprintf("生成配置失败: %v", err)})
	}

	if err := os.WriteFile("config/config.yaml", yamlData, 0644); err != nil {
		return c.JSON(fiber.Map{"code": 0, "msg": fmt.Sprintf("写入配置失败: %v", err)})
	}

	h.installed = true

	return c.JSON(fiber.Map{"code": 1, "msg": "安装成功"})
}

func toInt(s string) int {
	n := 0
	fmt.Sscanf(s, "%d", &n)
	if n == 0 {
		n = 3306
	}
	return n
}

func randomStr(n int) string {
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = chars[i%len(chars)]
	}
	return string(b)
}
