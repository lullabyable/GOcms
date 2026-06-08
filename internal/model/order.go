package model

// CardKey 卡密模型
type CardKey struct {
	CardID    int    `gorm:"primaryKey;column:card_id" json:"card_id"`
	CardCode  string `gorm:"column:card_code;size:50;uniqueIndex" json:"card_code"`
	GroupID   int    `gorm:"column:group_id" json:"group_id"`
	Points    int    `gorm:"column:points" json:"points"`
	Days      int    `gorm:"column:days" json:"days"`
	Status    int    `gorm:"column:status" json:"status"` // 0=未使用 1=已使用 2=已禁用
	UserID    int    `gorm:"column:user_id" json:"user_id"`
	UsedTime  int64  `gorm:"column:used_time" json:"used_time"`
	CreatedAt int64  `gorm:"column:created_at" json:"created_at"`
	BatchNo   string `gorm:"column:batch_no;size:50" json:"batch_no"`
}

func (CardKey) TableName() string { return "mac_card" }

// Order 订单模型
type Order struct {
	OrderID     int    `gorm:"primaryKey;column:order_id" json:"order_id"`
	OrderNo     string `gorm:"column:order_no;size:50;uniqueIndex" json:"order_no"`
	UserID      int    `gorm:"column:user_id;index" json:"user_id"`
	ProductType int    `gorm:"column:product_type" json:"product_type"` // 1=会员 2=积分 3=卡密
	ProductName string `gorm:"column:product_name;size:200" json:"product_name"`
	Amount      int    `gorm:"column:amount" json:"amount"`             // 金额（分）
	Points      int    `gorm:"column:points" json:"points"`
	Days        int    `gorm:"column:days" json:"days"`
	GroupID     int    `gorm:"column:group_id" json:"group_id"`
	PayType     string `gorm:"column:pay_type;size:20" json:"pay_type"` // alipay/wechat/balance/card
	Status      int    `gorm:"column:status" json:"status"`             // 0=待支付 1=已支付 2=已取消 3=已退款
	PayTime     int64  `gorm:"column:pay_time" json:"pay_time"`
	CreatedAt   int64  `gorm:"column:created_at" json:"created_at"`
	UpdatedAt   int64  `gorm:"column:updated_at" json:"updated_at"`
	Extra       string `gorm:"column:extra;type:text" json:"extra"`
}

func (Order) TableName() string { return "mac_order" }

// Payment 支付配置模型
type Payment struct {
	PayID     int    `gorm:"primaryKey;column:pay_id" json:"pay_id"`
	PayType   string `gorm:"column:pay_type;size:20;uniqueIndex" json:"pay_type"`
	PayName   string `gorm:"column:pay_name;size:50" json:"pay_name"`
	Config    string `gorm:"column:config;type:text" json:"config"`
	Status    int    `gorm:"column:status" json:"status"`
	Sort      int    `gorm:"column:sort" json:"sort"`
}

func (Payment) TableName() string { return "mac_payment" }
