package collect

// CollectSource 采集源配置
type CollectSource struct {
	ID          int    `json:"collect_id"`
	Name        string `json:"collect_name"`
	APIURL      string `json:"collect_url"`
	Charset     string `json:"collect_charset"`
	Format      string `json:"collect_format"` // xml, json
	InRule      string `json:"inrule"`         // a=追加, r=替换, d=删除
	UpRule      string `json:"uprule"`         // a=追加, r=替换, d=删除
	Filter      string `json:"filter"`         // 过滤正则
	Thesaurus   string `json:"thesaurus"`      // 同义词
	Words       string `json:"words"`          // 敏感词
	PicDownload int    `json:"pic"`            // 图片下载 0=不下载 1=下载
	TypeMapping string `json:"type_mapping"`   // 分类映射
	Schedule    string `json:"schedule"`       // 定时采集
}

// CollectOptions 采集选项
type CollectOptions struct {
	IDs     string
	TypeID  int
	Page    int
	Hours   int
	Keyword string
	Year    int
	IsEnd   int
}

// CollectResult 采集结果
type CollectResult struct {
	Imported int `json:"imported"`
	Updated  int `json:"updated"`
	Errors   int `json:"errors"`
}

// SourceVideo 源站视频数据
type SourceVideo struct {
	SourceID      int
	TypeID        int
	TypeName      string
	Name          string
	Pic           string
	Lang          string
	Area          string
	Year          string
	State         string
	Remarks       string
	Des           string
	Actor         string
	Director      string
	Last          string
	PlayList      []PlayGroup
	CustomFields  map[string]string
}

// PlayGroup 播放源组
type PlayGroup struct {
	Flag string
	URLs []PlayURL
}

// PlayURL 单个播放地址
type PlayURL struct {
	Name string
	URL  string
}

// CollectRule 自定义采集规则
type CollectRule struct {
	ID              int
	Name            string
	URLPattern      string
	MaxPage         int
	Charset         string
	ListSelector    string
	ProgramConfig   ProgramConfig
	CustomizeConfig []CustomField
}

// ProgramConfig 程序配置
type ProgramConfig struct {
	Map   map[string]string `json:"map"`
	Funcs map[string]string `json:"funcs"`
}

// CustomField 自定义字段
type CustomField struct {
	Name   string `json:"name"`
	EnName string `json:"en_name"`
	Rule   string `json:"rule"`
}
