package payment_test

import (
	"testing"

	"gocms/internal/testutil"
	"gocms/internal/model"
	"gocms/internal/service/payment"
)

func TestCreateOrder(t *testing.T) {
	db := testutil.TestDB(t)
	svc := payment.NewService(db)

	order, err := svc.CreateOrder(1, 1, "月度会员", 3000, "alipay", map[string]interface{}{
		"days":     30,
		"group_id": 2,
	})
	if err != nil {
		t.Fatalf("CreateOrder failed: %v", err)
	}
	if order.OrderNo == "" {
		t.Error("order number should not be empty")
	}
	if order.Amount != 3000 {
		t.Errorf("expected amount=3000, got %d", order.Amount)
	}
	if order.Status != 0 {
		t.Errorf("expected status=0, got %d", order.Status)
	}
}

func TestPayOrder(t *testing.T) {
	db := testutil.TestDB(t)
	svc := payment.NewService(db)

	// 创建用户
	db.Create(&model.User{UserID: 10, UserName: "paytest", UserPwd: "pwd", GroupID: 1, UserPoints: 100})

	// 创建会员订单
	order, _ := svc.CreateOrder(10, 1, "月度会员", 3000, "alipay", map[string]interface{}{
		"days":     30,
		"group_id": 2,
	})

	// 支付
	if err := svc.PayOrder(order.OrderNo); err != nil {
		t.Fatalf("PayOrder failed: %v", err)
	}

	// 验证订单状态
	updated, _ := svc.GetOrder(order.OrderNo)
	if updated.Status != 1 {
		t.Errorf("expected status=1, got %d", updated.Status)
	}

	// 验证用户组升级
	var user model.User
	db.First(&user, 10)
	if user.GroupID != 2 {
		t.Errorf("expected group_id=2, got %d", user.GroupID)
	}
}

func TestPayOrderPoints(t *testing.T) {
	db := testutil.TestDB(t)
	svc := payment.NewService(db)

	db.Create(&model.User{UserID: 20, UserName: "pointstest", UserPwd: "pwd", UserPoints: 100})

	order, _ := svc.CreateOrder(20, 2, "积分充值", 1000, "wechat", map[string]interface{}{
		"points": 500,
	})

	svc.PayOrder(order.OrderNo)

	var user model.User
	db.First(&user, 20)
	if user.UserPoints != 600 {
		t.Errorf("expected points=600, got %d", user.UserPoints)
	}
}

func TestCancelOrder(t *testing.T) {
	db := testutil.TestDB(t)
	svc := payment.NewService(db)

	order, _ := svc.CreateOrder(1, 1, "测试", 100, "alipay", nil)

	if err := svc.CancelOrder(order.OrderNo); err != nil {
		t.Fatalf("CancelOrder failed: %v", err)
	}

	updated, _ := svc.GetOrder(order.OrderNo)
	if updated.Status != 2 {
		t.Errorf("expected status=2, got %d", updated.Status)
	}
}

func TestGenerateCardKeys(t *testing.T) {
	db := testutil.TestDB(t)
	svc := payment.NewService(db)

	cards, err := svc.GenerateCardKeys("BATCH001", 5, 2, 100, 30)
	if err != nil {
		t.Fatalf("GenerateCardKeys failed: %v", err)
	}
	if len(cards) != 5 {
		t.Errorf("expected 5 cards, got %d", len(cards))
	}

	// 验证卡密唯一
	codes := make(map[string]bool)
	for _, c := range cards {
		if codes[c.CardCode] {
			t.Error("duplicate card code")
		}
		codes[c.CardCode] = true
		if len(c.CardCode) != 16 {
			t.Errorf("expected 16 char code, got %d", len(c.CardCode))
		}
	}
}

func TestUseCardKey(t *testing.T) {
	db := testutil.TestDB(t)
	svc := payment.NewService(db)

	db.Create(&model.User{UserID: 30, UserName: "cardtest", UserPwd: "pwd", GroupID: 1, UserPoints: 0})

	cards, _ := svc.GenerateCardKeys("BATCH002", 1, 2, 200, 30)

	if err := svc.UseCardKey(30, cards[0].CardCode); err != nil {
		t.Fatalf("UseCardKey failed: %v", err)
	}

	// 验证卡密已使用
	var card model.CardKey
	db.Where("card_code = ?", cards[0].CardCode).First(&card)
	if card.Status != 1 {
		t.Errorf("expected status=1, got %d", card.Status)
	}
	if card.UserID != 30 {
		t.Errorf("expected user_id=30, got %d", card.UserID)
	}

	// 验证用户积分
	var user model.User
	db.First(&user, 30)
	if user.UserPoints != 200 {
		t.Errorf("expected points=200, got %d", user.UserPoints)
	}

	// 重复使用应失败
	err := svc.UseCardKey(30, cards[0].CardCode)
	if err == nil {
		t.Error("should fail on reuse")
	}
}

func TestGetUserOrders(t *testing.T) {
	db := testutil.TestDB(t)
	svc := payment.NewService(db)

	svc.CreateOrder(1, 1, "A", 100, "alipay", nil)
	svc.CreateOrder(1, 1, "B", 200, "wechat", nil)
	svc.CreateOrder(2, 1, "C", 300, "alipay", nil)

	orders, total, err := svc.GetUserOrders(1, 1, 10)
	if err != nil {
		t.Fatalf("GetUserOrders failed: %v", err)
	}
	if total != 2 {
		t.Errorf("expected total=2, got %d", total)
	}
	if len(orders) != 2 {
		t.Errorf("expected 2 orders, got %d", len(orders))
	}
}

func TestPayOrderNotExist(t *testing.T) {
	db := testutil.TestDB(t)
	svc := payment.NewService(db)

	err := svc.PayOrder("NONEXIST")
	if err == nil {
		t.Error("should fail for non-existent order")
	}
}
