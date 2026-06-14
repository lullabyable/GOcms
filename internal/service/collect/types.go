package collect

import "time"

// CollectSource 采集源配置（对应 mac_collect 表）
type CollectSource struct {
	ID             int    `gorm:"primaryKey;column:collect_id" json:"collect_id"`
	Name           string `gorm:"column:collect_name;size:100" json:"collect_name"`
	APIURL         string `gorm:"column:collect_url;size:500" json:"collect_url"`
	Type           int    `gorm:"column:collect_type" json:"collect_type"`       // 1=xml 2=json
	Mid            int    `gorm:"column:collect_mid" json:"collect_mid"`         // 1=vod 2=art 8=actor 9=role 12=manga
	AppID          string `gorm:"column:collect_appid;size:30" json:"collect_appid"`
	AppKey         string `gorm:"column:collect_appkey;size:30" json:"collect_appkey"`
	Param          string `gorm:"column:collect_param;size:200" json:"collect_param"`
	Filter         int    `gorm:"column:collect_filter" json:"collect_filter"`               // 0=不过滤 1=增改 2=仅增 3=仅改
	FilterFrom     string `gorm:"column:collect_filter_from;size:255" json:"collect_filter_from"`
	FilterYear     string `gorm:"column:collect_filter_year;size:255" json:"collect_filter_year"`
	Opt            int    `gorm:"column:collect_opt" json:"collect_opt"`                     // 0=增+改 1=仅增 2=仅改
	SyncPicOpt     int    `gorm:"column:collect_sync_pic_opt" json:"collect_sync_pic_opt"`   // 0=跟随全局 1=开启 2=关闭
	CreatedAt      int64  `gorm:"column:created_at" json:"created_at"`
}

func (CollectSource) TableName() string { return "mac_collect" }

// CollectBind 分类绑定（远程分类→本地分类）
type CollectBind struct {
	ID           int    `gorm:"primaryKey;column:id" json:"id"`
	CollectFlag  string `gorm:"column:collect_flag;size:100;uniqueIndex" json:"collect_flag"` // cjflag_remoteTypeId
	RemoteTypeID int    `gorm:"column:remote_type_id" json:"remote_type_id"`
	RemoteName   string `gorm:"column:remote_name;size:100" json:"remote_name"`
	LocalTypeID  int    `gorm:"column:local_type_id" json:"local_type_id"`
	LocalName    string `gorm:"column:local_name;size:100" json:"local_name"`
}

func (CollectBind) TableName() string { return "mac_collect_bind" }

// CollectJob 采集任务状态
type CollectJob struct {
	ID         string    `json:"id"`
	SourceID   int       `json:"source_id"`
	SourceName string    `json:"source_name"`
	Mid        int       `json:"mid"` // 1=vod 2=art ...
	Status     string    `json:"status"` // pending/running/done/error/paused
	Total      int       `json:"total"`
	Imported   int       `json:"imported"`
	Updated    int       `json:"updated"`
	Skipped    int       `json:"skipped"`
	Errors     int       `json:"errors"`
	Current    int       `json:"current"` // 当前处理到第几条
	Page       int       `json:"page"`    // 当前采集到第几页
	TotalPages int       `json:"total_pages"`
	Message    string    `json:"message"`
	StartedAt  time.Time `json:"started_at"`
	EndedAt    time.Time `json:"ended_at"`
}

// CollectConfig 采集全局配置
type CollectConfig struct {
	Workers      int    `json:"workers"`        // 并发线程数，默认 5
	BatchSize    int    `json:"batch_size"`     // 批量写入大小，默认 50
	RateLimit    int    `json:"rate_limit"`     // 每秒请求限制，默认 10
	BufferSize   int    `json:"buffer_size"`    // 任务缓冲区大小，默认 500
	Timeout      int    `json:"timeout"`        // 单个请求超时(秒)，默认 30
	Status       int    `json:"status"`         // 采集后审核状态 0=待审 1=通过
	HitsStart    int    `json:"hits_start"`     // 随机点击起始
	HitsEnd      int    `json:"hits_end"`       // 随机点击结束
	UpdownStart  int    `json:"updown_start"`   // 随机顶踩起始
	UpdownEnd    int    `json:"updown_end"`     // 随机顶踩结束
	ScoreRandom  int    `json:"score_random"`   // 随机评分 0=关 1=开
	InRule       string `json:"inrule"`         // 去重规则 abcdefgh
	UpRule       string `json:"uprule"`         // 更新规则 a-w
	Filter       string `json:"filter"`         // 过滤关键词(逗号分隔)
	PicSync      int    `json:"pic_sync"`       // 图片同步 0=关 1=开
}

func DefaultConfig() CollectConfig {
	return CollectConfig{
		Workers:    5,
		BatchSize:  50,
		RateLimit:  10,
		BufferSize: 500,
		Timeout:    30,
		Status:     1,
		InRule:     "a",
		UpRule:     "a",
	}
}

// RemoteCategory 远程分类
type RemoteCategory struct {
	TypeID   int    `json:"type_id"`
	TypeName string `json:"type_name"`
	IsBind   bool   `json:"is_bind"`
	LocalID  int    `json:"local_type_id"`
	LocalName string `json:"local_name"`
}

// RemoteVideo 远程视频数据（从API获取）
type RemoteVideo struct {
	VodID        int    `json:"vod_id"`
	TypeID       int    `json:"type_id"`
	TypeName     string `json:"type_name"`
	VodName      string `json:"vod_name"`
	VodSub       string `json:"vod_sub"`
	VodEn        string `json:"vod_en"`
	VodPic       string `json:"vod_pic"`
	VodActor     string `json:"vod_actor"`
	VodDirector  string `json:"vod_director"`
	VodWriter    string `json:"vod_writer"`
	VodArea      string `json:"vod_area"`
	VodLang      string `json:"vod_lang"`
	VodYear      string `json:"vod_year"`
	VodContent   string `json:"vod_content"`
	VodBlurb     string `json:"vod_blurb"`
	VodRemarks   string `json:"vod_remarks"`
	VodClass     string `json:"vod_class"`
	VodTag       string `json:"vod_tag"`
	VodState     string `json:"vod_state"`
	VodVersion   string `json:"vod_version"`
	VodPlayFrom  string `json:"vod_play_from"`
	VodPlayURL   string `json:"vod_play_url"`
	VodDownFrom  string `json:"vod_down_from"`
	VodDownURL   string `json:"vod_down_url"`
	VodScore     string `json:"vod_score"`
	VodScoreAll  int    `json:"vod_score_all"`
	VodScoreNum  int    `json:"vod_score_num"`
	VodHits      int    `json:"vod_hits"`
	VodHitsDay   int    `json:"vod_hits_day"`
	VodHitsWeek  int    `json:"vod_hits_week"`
	VodHitsMonth int    `json:"vod_hits_month"`
	VodTime      string `json:"vod_time"`
	VodPlot      int    `json:"vod_plot"`
	VodPlotName  string `json:"vod_plot_name"`
	VodPlotDetail string `json:"vod_plot_detail"`
}

// RemoteArt 远程文章
type RemoteArt struct {
	ArtID       int    `json:"art_id"`
	TypeID      int    `json:"type_id"`
	TypeName    string `json:"type_name"`
	ArtName     string `json:"art_name"`
	ArtSub      string `json:"art_sub"`
	ArtEn       string `json:"art_en"`
	ArtPic      string `json:"art_pic"`
	ArtAuthor   string `json:"art_author"`
	ArtFrom     string `json:"art_from"`
	ArtContent  string `json:"art_content"`
	ArtBlurb    string `json:"art_blurb"`
	ArtTag      string `json:"art_tag"`
	ArtClass    string `json:"art_class"`
	ArtTime     string `json:"art_time"`
	ArtHits     int    `json:"art_hits"`
}

// APIResponse JSON API 通用响应
type APIResponse struct {
	Code       int              `json:"code"`
	Msg        string           `json:"msg"`
	Page       int              `json:"page"`
	PageCount  int              `json:"pagecount"`
	Limit      int              `json:"limit"`
	Total      int              `json:"total"`
	List       []map[string]interface{} `json:"list"`
	Class      []RemoteCategory `json:"class"`
}

// CollectProgress 采集进度（用于断点续采）
type CollectProgress struct {
	SourceID   int    `json:"source_id"`
	Mid        int    `json:"mid"`
	LastPage   int    `json:"last_page"`
	LastID     int    `json:"last_id"`
	Timestamp  int64  `json:"timestamp"`
}

// ==================== 兼容旧类型（html_collector/playurl 使用） ====================

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
