package payment

import (
	"crypto/md5"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"gorm.io/gorm"
	"gocms/internal/model"
)

// Service 支付服务
type Service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

// GenerateOrderNo 生成订单号
func (s *Service) GenerateOrderNo() string {
	return fmt.Sprintf("%s%06d", time.Now().Format("20060102150405"), rand.Intn(1000000))
}

// CreateOrder 创建订单
func (s *Service) CreateOrder(userID int, productType int, productName string, amount int, payType string, extra map[string]interface{}) (*model.Order, error) {
	order := &model.Order{
		OrderNo:     s.GenerateOrderNo(),
		UserID:      userID,
		ProductType: productType,
		ProductName: productName,
		Amount:      amount,
		PayType:     payType,
		Status:      0,
		CreatedAt:   time.Now().Unix(),
		UpdatedAt:   time.Now().Unix(),
	}

	// 根据产品类型设置
	switch productType {
	case 1: // 会员
		if days, ok := extra["days"].(int); ok {
			order.Days = days
		}
		if groupID, ok := extra["group_id"].(int); ok {
			order.GroupID = groupID
		}
	case 2: // 积分
		if points, ok := extra["points"].(int); ok {
			order.Points = points
		}
	}

	if err := s.db.Create(order).Error; err != nil {
		return nil, err
	}
	return order, nil
}

// PayOrder 支付订单（模拟支付完成）
func (s *Service) PayOrder(orderNo string) error {
	var order model.Order
	if err := s.db.Where("order_no = ?", orderNo).First(&order).Error; err != nil {
		return fmt.Errorf("订单不存在")
	}

	if order.Status != 0 {
		return fmt.Errorf("订单状态异常")
	}

	// 更新订单状态
	now := time.Now().Unix()
	s.db.Model(&order).Updates(map[string]interface{}{
		"status":   1,
		"pay_time": now,
	})

	// 执行订单对应的操作
	switch order.ProductType {
	case 1: // 会员升级
		s.db.Model(&model.User{}).Where("user_id = ?", order.UserID).Updates(map[string]interface{}{
			"group_id":    order.GroupID,
			"expiry_time": now + int64(order.Days*86400),
		})
	case 2: // 积分充值
		s.db.Model(&model.User{}).Where("user_id = ?", order.UserID).
			Update("user_points", gorm.Expr("user_points + ?", order.Points))
	}

	return nil
}

// CancelOrder 取消订单
func (s *Service) CancelOrder(orderNo string) error {
	return s.db.Model(&model.Order{}).Where("order_no = ? AND status = 0", orderNo).
		Update("status", 2).Error
}

// GetOrder 获取订单
func (s *Service) GetOrder(orderNo string) (*model.Order, error) {
	var order model.Order
	err := s.db.Where("order_no = ?", orderNo).First(&order).Error
	return &order, err
}

// GetUserOrders 获取用户订单列表
func (s *Service) GetUserOrders(userID int, page, pageSize int) ([]model.Order, int64, error) {
	var orders []model.Order
	var total int64
	query := s.db.Model(&model.Order{}).Where("user_id = ?", userID)
	query.Count(&total)
	err := query.Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&orders).Error
	return orders, total, err
}

// GetAllOrders 获取所有订单（后台）
func (s *Service) GetAllOrders(page, pageSize int, status int) ([]model.Order, int64, error) {
	var orders []model.Order
	var total int64
	query := s.db.Model(&model.Order{})
	if status >= 0 {
		query = query.Where("status = ?", status)
	}
	query.Count(&total)
	err := query.Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&orders).Error
	return orders, total, err
}

// --- 卡密管理 ---

// GenerateCardKeys 批量生成卡密
func (s *Service) GenerateCardKeys(batchNo string, count int, groupID, points, days int) ([]model.CardKey, error) {
	var cards []model.CardKey
	for i := 0; i < count; i++ {
		code := s.generateCardCode()
		card := model.CardKey{
			CardCode:  code,
			GroupID:   groupID,
			Points:    points,
			Days:      days,
			Status:    0,
			BatchNo:   batchNo,
			CreatedAt: time.Now().Unix(),
		}
		cards = append(cards, card)
	}

	if err := s.db.CreateInBatches(cards, 100).Error; err != nil {
		return nil, err
	}
	return cards, nil
}

// UseCardKey 使用卡密
func (s *Service) UseCardKey(userID int, code string) error {
	var card model.CardKey
	if err := s.db.Where("card_code = ? AND status = 0", code).First(&card).Error; err != nil {
		return fmt.Errorf("卡密无效或已使用")
	}

	now := time.Now().Unix()

	// 更新卡密状态
	s.db.Model(&card).Updates(map[string]interface{}{
		"status":    1,
		"user_id":   userID,
		"used_time": now,
	})

	// 应用卡密权益
	if card.GroupID > 0 {
		s.db.Model(&model.User{}).Where("user_id = ?", userID).Updates(map[string]interface{}{
			"group_id":    card.GroupID,
			"expiry_time": now + int64(card.Days*86400),
		})
	}
	if card.Points > 0 {
		s.db.Model(&model.User{}).Where("user_id = ?", userID).
			Update("user_points", gorm.Expr("user_points + ?", card.Points))
	}

	return nil
}

// GetCardKeys 获取卡密列表（后台）
func (s *Service) GetCardKeys(page, pageSize int, status int, batchNo string) ([]model.CardKey, int64, error) {
	var cards []model.CardKey
	var total int64
	query := s.db.Model(&model.CardKey{})
	if status >= 0 {
		query = query.Where("status = ?", status)
	}
	if batchNo != "" {
		query = query.Where("batch_no = ?", batchNo)
	}
	query.Count(&total)
	err := query.Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&cards).Error
	return cards, total, err
}

func (s *Service) generateCardCode() string {
	h := md5.New()
	h.Write([]byte(strconv.FormatInt(time.Now().UnixNano(), 36) + strconv.Itoa(rand.Intn(999999))))
	return fmt.Sprintf("%X", h.Sum(nil))[:16]
}

// GetPaymentConfigs 获取支付配置列表
func (s *Service) GetPaymentConfigs() ([]model.Payment, error) {
	var payments []model.Payment
	err := s.db.Order("sort ASC").Find(&payments).Error
	return payments, err
}

// SavePaymentConfig 保存支付配置
func (s *Service) SavePaymentConfig(pay *model.Payment) error {
	var existing model.Payment
	if err := s.db.Where("pay_type = ?", pay.PayType).First(&existing).Error; err != nil {
		return s.db.Create(pay).Error
	}
	return s.db.Model(&existing).Updates(map[string]interface{}{
		"pay_name": pay.PayName,
		"config":   pay.Config,
		"status":   pay.Status,
		"sort":     pay.Sort,
	}).Error
}
