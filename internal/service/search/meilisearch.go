package search

import (
	"fmt"
	"time"

	"github.com/meilisearch/meilisearch-go"
	"gorm.io/gorm"
	"gocms/internal/model"
)

// MeilisearchConfig Meilisearch 配置
type MeilisearchConfig struct {
	Host   string `mapstructure:"host"`
	APIKey string `mapstructure:"api_key"`
}

// MeilisearchEngine Meilisearch 搜索引擎
type MeilisearchEngine struct {
	client *meilisearch.Client
	db     *gorm.DB
}

// NewMeilisearchEngine 创建 Meilisearch 引擎
func NewMeilisearchEngine(cfg MeilisearchConfig, db *gorm.DB) *MeilisearchEngine {
	client := meilisearch.NewClient(meilisearch.ClientConfig{
		Host:   cfg.Host,
		APIKey: cfg.APIKey,
	})
	return &MeilisearchEngine{
		client: client,
		db:     db,
	}
}

// Health 健康检查
func (e *MeilisearchEngine) Health() error {
	health, err := e.client.Health()
	if err != nil {
		return fmt.Errorf("Meilisearch 连接失败: %w", err)
	}
	if health.Status != "available" {
		return fmt.Errorf("Meilisearch 不可用: %s", health.Status)
	}
	return nil
}

// SyncVod 同步视频数据到 Meilisearch
func (e *MeilisearchEngine) SyncVod() error {
	index := e.client.Index("gocms_vod")

	// 设置可搜索属性
	settings := &meilisearch.Settings{
		SearchableAttributes: []string{"vod_name", "vod_sub", "vod_en", "vod_actor", "vod_director", "vod_class", "vod_tag", "vod_content"},
		FilterableAttributes: []string{"type_id", "vod_year", "vod_area", "vod_lang", "vod_status"},
		SortableAttributes:   []string{"vod_time_add", "vod_hits", "vod_score"},
	}
	index.UpdateSettings(settings)

	// 分批同步
	var total int64
	e.db.Model(&model.Vod{}).Where("vod_status = 1").Count(&total)

	batchSize := 500
	for offset := 0; int64(offset) < total; offset += batchSize {
		var vods []model.Vod
		e.db.Where("vod_status = 1").
			Order("vod_id ASC").
			Offset(offset).
			Limit(batchSize).
			Find(&vods)

		if len(vods) == 0 {
			break
		}

		// 转换为搜索文档
		docs := make([]map[string]interface{}, len(vods))
		for i, v := range vods {
			docs[i] = map[string]interface{}{
				"id":           v.ID,
				"type_id":      v.TypeID,
				"vod_name":     v.VodName,
				"vod_sub":      v.VodSub,
				"vod_en":       v.VodEn,
				"vod_pic":      v.VodPic,
				"vod_actor":    v.VodActor,
				"vod_director": v.VodDirector,
				"vod_class":    v.VodClass,
				"vod_tag":      v.VodTag,
				"vod_remarks":  v.VodRemarks,
				"vod_content":  v.VodContent,
				"vod_area":     v.VodArea,
				"vod_lang":     v.VodLang,
				"vod_year":     v.VodYear,
				"vod_score":    v.VodScore,
				"vod_hits":     v.VodHits,
				"vod_time_add": v.VodTimeAdd,
				"vod_status":   v.VodStatus,
			}
		}

		if _, err := index.AddDocuments(docs); err != nil {
			return fmt.Errorf("同步视频失败 (offset=%d): %w", offset, err)
		}
	}

	return nil
}

// SyncArt 同步文章数据到 Meilisearch
func (e *MeilisearchEngine) SyncArt() error {
	index := e.client.Index("gocms_art")

	settings := &meilisearch.Settings{
		SearchableAttributes: []string{"art_name", "art_sub", "art_en", "art_author", "art_class", "art_tag", "art_content"},
		FilterableAttributes: []string{"type_id", "art_status"},
		SortableAttributes:   []string{"art_time_add", "art_hits"},
	}
	index.UpdateSettings(settings)

	var total int64
	e.db.Model(&model.Art{}).Where("art_status = 1").Count(&total)

	batchSize := 500
	for offset := 0; int64(offset) < total; offset += batchSize {
		var arts []model.Art
		e.db.Where("art_status = 1").
			Order("art_id ASC").
			Offset(offset).
			Limit(batchSize).
			Find(&arts)

		if len(arts) == 0 {
			break
		}

		docs := make([]map[string]interface{}, len(arts))
		for i, a := range arts {
			docs[i] = map[string]interface{}{
				"id":           a.ID,
				"type_id":      a.TypeID,
				"art_name":     a.ArtName,
				"art_sub":      a.ArtSub,
				"art_en":       a.ArtEn,
				"art_pic":      a.ArtPic,
				"art_author":   a.ArtAuthor,
				"art_class":    a.ArtClass,
				"art_tag":      a.ArtTag,
				"art_content":  a.ArtContent,
				"art_remarks":  a.ArtRemarks,
				"art_hits":     a.ArtHits,
				"art_time_add": a.ArtTimeAdd,
				"art_status":   a.ArtStatus,
			}
		}

		if _, err := index.AddDocuments(docs); err != nil {
			return fmt.Errorf("同步文章失败 (offset=%d): %w", offset, err)
		}
	}

	return nil
}

// Search 全文搜索（Meilisearch）
func (e *MeilisearchEngine) Search(indexName, keyword string, page, pageSize int) (*SearchResult, error) {
	index := e.client.Index(indexName)

	offset := int64((page - 1) * pageSize)
	limit := int64(pageSize)

	searchResp, err := index.Search(keyword, &meilisearch.SearchRequest{
		Offset:  offset,
		Limit:   limit,
		HighlightPreTag:  "<em>",
		HighlightPostTag: "</em>",
		AttributesToHighlight: []string{"vod_name", "art_name"},
	})
	if err != nil {
		return nil, fmt.Errorf("搜索失败: %w", err)
	}

	// 转换结果
	var items []map[string]interface{}
	for _, hit := range searchResp.Hits {
		if m, ok := hit.(map[string]interface{}); ok {
			items = append(items, m)
		}
	}

	return &SearchResult{
		Total:    searchResp.EstimatedTotalHits,
		Page:     page,
		PageSize: pageSize,
		Items:    items,
		Query:    keyword,
	}, nil
}

// SearchResult 搜索结果
type SearchResult struct {
	Total    int64                    `json:"total"`
	Page     int                      `json:"page"`
	PageSize int                      `json:"page_size"`
	Items    []map[string]interface{} `json:"items"`
	Query    string                   `json:"query"`
}

// IndexDocument 索引单个文档（实时同步）
func (e *MeilisearchEngine) IndexDocument(indexName string, doc map[string]interface{}) error {
	index := e.client.Index(indexName)
	_, err := index.AddDocuments([]map[string]interface{}{doc})
	return err
}

// DeleteDocument 删除单个文档
func (e *MeilisearchEngine) DeleteDocument(indexName string, id string) error {
	index := e.client.Index(indexName)
	_, err := index.DeleteDocument(id)
	return err
}

// WaitForTask 等待任务完成
func (e *MeilisearchEngine) WaitForTask(taskUID int64, timeout time.Duration) error {
	_, err := e.client.WaitForTask(taskUID)
	return err
}
