package search

import (
	"fmt"

	"gorm.io/gorm"
	"gocms/internal/model"
)

// SearchService 搜索服务（自动降级）
type SearchService struct {
	db       *gorm.DB
	meili    *MeilisearchEngine
	useMeili bool
}

// NewSearchService 创建搜索服务
func NewSearchService(db *gorm.DB, meiliCfg *MeilisearchConfig) *SearchService {
	svc := &SearchService{
		db: db,
	}

	if meiliCfg != nil && meiliCfg.Host != "" {
		meili := NewMeilisearchEngine(*meiliCfg, db)
		if err := meili.Health(); err == nil {
			svc.meili = meili
			svc.useMeili = true
		}
	}

	return svc
}

// Search 搜索视频
func (s *SearchService) Search(keyword string, page, pageSize int) (*SearchResult, error) {
	if s.useMeili && s.meili != nil {
		result, err := s.meili.Search("gocms_vod", keyword, page, pageSize)
		if err == nil {
			return result, nil
		}
		// Meilisearch 失败，降级
		s.useMeili = false
	}

	// 降级到 MySQL
	return s.searchMySQL(keyword, page, pageSize)
}

// SearchArt 搜索文章
func (s *SearchService) SearchArt(keyword string, page, pageSize int) (*SearchResult, error) {
	if s.useMeili && s.meili != nil {
		result, err := s.meili.Search("gocms_art", keyword, page, pageSize)
		if err == nil {
			return result, nil
		}
		s.useMeili = false
	}

	return s.searchArtMySQL(keyword, page, pageSize)
}

// searchMySQL MySQL 降级搜索
func (s *SearchService) searchMySQL(keyword string, page, pageSize int) (*SearchResult, error) {
	query := s.db.Model(&model.Vod{}).Where("vod_status = 1")

	// 尝试 FULLTEXT，失败则 LIKE
	var total int64
	var vods []model.Vod

	// 先试 FULLTEXT
	ftQuery := query.Where("MATCH(vod_name, vod_sub, vod_content) AGAINST(? IN BOOLEAN MODE)", keyword)
	if err := ftQuery.Count(&total).Error; err != nil {
		// FULLTEXT 不可用，降级到 LIKE
		likeQuery := s.db.Model(&model.Vod{}).Where("vod_status = 1").
			Where("vod_name LIKE ? OR vod_sub LIKE ? OR vod_actor LIKE ? OR vod_director LIKE ?",
				"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
		likeQuery.Count(&total)
		likeQuery.Order("vod_id DESC").
			Offset((page - 1) * pageSize).
			Limit(pageSize).
			Find(&vods)
	} else {
		ftQuery.Order("vod_id DESC").
			Offset((page - 1) * pageSize).
			Limit(pageSize).
			Find(&vods)
	}

	// 转换结果
	items := make([]map[string]interface{}, len(vods))
	for i, v := range vods {
		items[i] = map[string]interface{}{
			"id":           v.ID,
			"type_id":      v.TypeID,
			"vod_name":     v.VodName,
			"vod_sub":      v.VodSub,
			"vod_pic":      v.VodPic,
			"vod_remarks":  v.VodRemarks,
			"vod_year":     v.VodYear,
			"vod_area":     v.VodArea,
			"vod_score":    v.VodScore,
			"vod_hits":     v.VodHits,
			"vod_time_add": v.VodTimeAdd,
		}
	}

	return &SearchResult{
		Total:    total,
		Page:     page,
		PageSize: pageSize,
		Items:    items,
		Query:    keyword,
	}, nil
}

// searchArtMySQL MySQL 降级搜索文章
func (s *SearchService) searchArtMySQL(keyword string, page, pageSize int) (*SearchResult, error) {
	var total int64
	var arts []model.Art

	query := s.db.Model(&model.Art{}).Where("art_status = 1").
		Where("art_name LIKE ? OR art_sub LIKE ? OR art_author LIKE ?",
			"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")

	query.Count(&total)
	query.Order("art_id DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&arts)

	items := make([]map[string]interface{}, len(arts))
	for i, a := range arts {
		items[i] = map[string]interface{}{
			"id":           a.ID,
			"type_id":      a.TypeID,
			"art_name":     a.ArtName,
			"art_sub":      a.ArtSub,
			"art_pic":      a.ArtPic,
			"art_remarks":  a.ArtRemarks,
			"art_author":   a.ArtAuthor,
			"art_time_add": a.ArtTimeAdd,
		}
	}

	return &SearchResult{
		Total:    total,
		Page:     page,
		PageSize: pageSize,
		Items:    items,
		Query:    keyword,
	}, nil
}

// SyncAll 同步所有数据到 Meilisearch
func (s *SearchService) SyncAll() error {
	if s.meili == nil {
		return fmt.Errorf("Meilisearch 未配置")
	}

	if err := s.meili.SyncVod(); err != nil {
		return fmt.Errorf("同步视频失败: %w", err)
	}
	if err := s.meili.SyncArt(); err != nil {
		return fmt.Errorf("同步文章失败: %w", err)
	}
	return nil
}

// IsMeilisearchAvailable 是否使用 Meilisearch
func (s *SearchService) IsMeilisearchAvailable() bool {
	return s.useMeili
}
