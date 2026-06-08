package collect

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	"gorm.io/gorm"
	"gocms/internal/model"
)

// Engine 资源站采集引擎
type Engine struct {
	db     *gorm.DB
	client *resty.Client
}

func NewEngine(db *gorm.DB) *Engine {
	return &Engine{
		db:     db,
		client: resty.New().SetTimeout(30 * time.Second),
	}
}

// CollectFromSource 从资源站采集
func (e *Engine) CollectFromSource(source CollectSource, opts CollectOptions) (*CollectResult, error) {
	result := &CollectResult{}
	apiURL := e.buildAPIURL(source, opts)

	resp, err := e.client.R().SetHeader("User-Agent", "MacCMS/10.0").Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("请求资源站失败: %v", err)
	}

	body := resp.String()
	var videos []SourceVideo

	if strings.Contains(body, "<rss") {
		videos, err = e.parseXMLResponse(body)
	} else {
		videos, err = e.parseJSONResponse(body)
	}
	if err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	for _, video := range videos {
		if err := e.processVideo(video, source, opts); err != nil {
			result.Errors++
			continue
		}
		result.Imported++
	}

	return result, nil
}

func (e *Engine) buildAPIURL(source CollectSource, opts CollectOptions) string {
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

	return baseURL + "?" + strings.Join(params, "&")
}

// parseXMLResponse 解析XML格式响应
func (e *Engine) parseXMLResponse(body string) ([]SourceVideo, error) {
	type XMLDD struct {
		Flag string `xml:"flag,attr"`
		Text string `xml:",chardata"`
	}
	type XMLVideo struct {
		ID       int    `xml:"id"`
		TID      int    `xml:"tid"`
		Name     string `xml:"name"`
		Pic      string `xml:"pic"`
		Lang     string `xml:"lang"`
		Area     string `xml:"area"`
		Year     string `xml:"year"`
		State    string `xml:"state"`
		Remarks  string `xml:"remarks"`
		Des      string `xml:"des"`
		Actor    string `xml:"actor"`
		Director string `xml:"director"`
		Last     string `xml:"last"`
		DL       struct {
			DD []XMLDD `xml:"dd"`
		} `xml:"dl"`
	}
	type XMLRSS struct {
		List struct {
			Video []XMLVideo `xml:"video"`
		} `xml:"list"`
		PageCount   int `xml:"pagecount"`
		RecordCount int `xml:"recordcount"`
	}

	var rss XMLRSS
	if err := xml.Unmarshal([]byte(body), &rss); err != nil {
		return nil, err
	}

	var videos []SourceVideo
	for _, v := range rss.List.Video {
		video := SourceVideo{
			SourceID: v.ID, TypeID: v.TID, Name: v.Name,
			Pic: v.Pic, Lang: v.Lang, Area: v.Area, Year: v.Year,
			State: v.State, Remarks: v.Remarks, Des: v.Des,
			Actor: v.Actor, Director: v.Director, Last: v.Last,
		}
		for _, dd := range v.DL.DD {
			video.PlayList = append(video.PlayList, PlayGroup{
				Flag: dd.Flag, URLs: ParsePlayURLs(dd.Text),
			})
		}
		videos = append(videos, video)
	}
	return videos, nil
}

// parseJSONResponse 解析JSON格式响应
func (e *Engine) parseJSONResponse(body string) ([]SourceVideo, error) {
	type JSONResp struct {
		List []struct {
			VodID       int    `json:"vod_id"`
			TypeID      int    `json:"type_id"`
			TypeName    string `json:"type_name"`
			VodName     string `json:"vod_name"`
			VodPic      string `json:"vod_pic"`
			VodLang     string `json:"vod_lang"`
			VodArea     string `json:"vod_area"`
			VodYear     string `json:"vod_year"`
			VodState    string `json:"vod_state"`
			VodRemarks  string `json:"vod_remarks"`
			VodContent  string `json:"vod_content"`
			VodActor    string `json:"vod_actor"`
			VodDirector string `json:"vod_director"`
			VodPlayFrom string `json:"vod_play_from"`
			VodPlayURL  string `json:"vod_play_url"`
			VodTime     string `json:"vod_time"`
		} `json:"list"`
		Pagecnt int `json:"pagecnt"`
		Total   int `json:"total"`
	}

	var resp JSONResp
	if err := json.Unmarshal([]byte(body), &resp); err != nil {
		return nil, err
	}

	var videos []SourceVideo
	for _, v := range resp.List {
		video := SourceVideo{
			SourceID: v.VodID, TypeID: v.TypeID, TypeName: v.TypeName,
			Name: v.VodName, Pic: v.VodPic, Lang: v.VodLang, Area: v.VodArea,
			Year: v.VodYear, State: v.VodState, Remarks: v.VodRemarks,
			Des: v.VodContent, Actor: v.VodActor, Director: v.VodDirector,
			Last: v.VodTime,
		}
		video.PlayList = ParsePlayFromURL(v.VodPlayFrom, v.VodPlayURL)
		videos = append(videos, video)
	}
	return videos, nil
}

// processVideo 处理单个视频（去重+入库）
func (e *Engine) processVideo(video SourceVideo, source CollectSource, opts CollectOptions) error {
	typeID := video.TypeID
	if typeID == 0 {
		typeID = 1 // 默认分类
	}

	var existing model.Vod
	result := e.db.Where("vod_name = ? AND type_id = ?", video.Name, typeID).First(&existing)

	if result.Error == gorm.ErrRecordNotFound {
		return e.insertVideo(video, typeID, source)
	}
	if result.Error != nil {
		return result.Error
	}
	return e.updateVideo(existing, video, typeID, source)
}

func (e *Engine) insertVideo(video SourceVideo, typeID int, source CollectSource) error {
	playFrom, playURL := "", ""
	if len(video.PlayList) > 0 {
		playFrom = FormatPlayFrom(video.PlayList)
		playURL = FormatPlayURL(video.PlayList)
	}

	vod := model.Vod{
		TypeID:      typeID,
		VodName:     video.Name,
		VodPic:      video.Pic,
		VodLang:     video.Lang,
		VodArea:     video.Area,
		VodYear:     video.Year,
		VodState:    video.State,
		VodRemarks:  video.Remarks,
		VodContent:  video.Des,
		VodActor:    video.Actor,
		VodDirector: video.Director,
		VodPlayFrom: playFrom,
		VodPlayURL:  playURL,
		VodStatus:   1,
	}
	return e.db.Create(&vod).Error
}

func (e *Engine) updateVideo(existing model.Vod, video SourceVideo, typeID int, source CollectSource) error {
	switch source.UpRule {
	case "a": // 追加
		if existing.VodPic == "" && video.Pic != "" {
			existing.VodPic = video.Pic
		}
		if existing.VodContent == "" && video.Des != "" {
			existing.VodContent = video.Des
		}
		if len(video.PlayList) > 0 {
			existing.VodPlayFrom, existing.VodPlayURL = MergePlayList(
				existing.VodPlayFrom, existing.VodPlayURL, video.PlayList)
		}
	case "r": // 替换
		existing.VodPic = video.Pic
		existing.VodContent = video.Des
		existing.VodActor = video.Actor
		existing.VodDirector = video.Director
		existing.VodRemarks = video.Remarks
		if len(video.PlayList) > 0 {
			existing.VodPlayFrom = FormatPlayFrom(video.PlayList)
			existing.VodPlayURL = FormatPlayURL(video.PlayList)
		}
	case "d": // 删除
		return e.db.Delete(&existing).Error
	}

	return e.db.Save(&existing).Error
}
