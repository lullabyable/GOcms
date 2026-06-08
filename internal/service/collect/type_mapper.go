package collect

import (
	"strings"
	"sync"

	"gorm.io/gorm"
	"gocms/internal/model"
)

// TypeMapper 分类映射器
type TypeMapper struct {
	db        *gorm.DB
	mu        sync.RWMutex
	nameIndex map[string]int
	manualMap map[int]int
}

func NewTypeMapper(db *gorm.DB) *TypeMapper {
	tm := &TypeMapper{
		db:        db,
		nameIndex: make(map[string]int),
		manualMap: make(map[int]int),
	}
	tm.loadFromDB()
	return tm
}

func (tm *TypeMapper) loadFromDB() {
	var types []model.Type
	tm.db.Find(&types)
	tm.nameIndex = make(map[string]int, len(types))
	for _, t := range types {
		tm.nameIndex[t.TypeName] = t.TypeID
	}
}

// SetManualMapping 设置手动分类映射
func (tm *TypeMapper) SetManualMapping(sourceTypeID, localTypeID int) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.manualMap[sourceTypeID] = localTypeID
}

// MapTypeID 映射分类ID
func (tm *TypeMapper) MapTypeID(sourceTypeID int, sourceTypeName string) int {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	// 手动映射优先
	if localID, ok := tm.manualMap[sourceTypeID]; ok && localID > 0 {
		return localID
	}

	// 按名称匹配
	cleanName := strings.TrimSpace(sourceTypeName)
	for name, id := range tm.nameIndex {
		if strings.TrimSpace(name) == cleanName {
			return id
		}
	}

	return 0
}
