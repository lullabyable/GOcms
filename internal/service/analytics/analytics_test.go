package analytics_test

import (
	"testing"
	"time"

	"gocms/internal/testutil"
	"gocms/internal/model"
	"gocms/internal/service/analytics"
)

func TestRecordVisit(t *testing.T) {
	db := testutil.TestDB(t)
	svc := analytics.NewService(db)

	svc.RecordVisit("/test", "127.0.0.1", "Mozilla/5.0", "")

	var count int64
	db.Model(&model.Visit{}).Count(&count)
	if count != 1 {
		t.Errorf("expected 1 visit, got %d", count)
	}
}

func TestRecordVisitUnique(t *testing.T) {
	db := testutil.TestDB(t)
	svc := analytics.NewService(db)

	// 第一次访问 - 唯一
	svc.RecordVisit("/page1", "192.168.1.1", "UA", "")
	// 同IP不同页面 - 不唯一
	svc.RecordVisit("/page2", "192.168.1.1", "UA", "")

	var visits []model.Visit
	db.Find(&visits)
	if len(visits) != 2 {
		t.Fatalf("expected 2 visits, got %d", len(visits))
	}

	// 第一条应该是唯一访客
	if visits[0].IsUnique != 1 {
		t.Error("first visit should be unique")
	}
	// 第二条不是唯一访客
	if visits[1].IsUnique != 0 {
		t.Error("second visit should not be unique")
	}
}

func TestGetDashboard(t *testing.T) {
	db := testutil.TestDB(t)
	svc := analytics.NewService(db)

	// 填充测试数据
	testutil.SeedTestData(db)
	today := time.Now().Format("2006-01-02")
	db.Create(&model.Visit{URL: "/test", IP: "127.0.0.1", Date: today, VisitTime: time.Now().Unix(), IsUnique: 1})
	db.Create(&model.Visit{URL: "/test2", IP: "127.0.0.2", Date: today, VisitTime: time.Now().Unix(), IsUnique: 1})

	dash, err := svc.GetDashboard()
	if err != nil {
		t.Fatalf("GetDashboard failed: %v", err)
	}

	if dash.TodayPV != 2 {
		t.Errorf("expected TodayPV=2, got %d", dash.TodayPV)
	}
	if dash.TotalVod != 2 {
		t.Errorf("expected TotalVod=2, got %d", dash.TotalVod)
	}
	if dash.TotalArt != 1 {
		t.Errorf("expected TotalArt=1, got %d", dash.TotalArt)
	}
}

func TestGetTrend(t *testing.T) {
	db := testutil.TestDB(t)
	svc := analytics.NewService(db)

	// 插入统计记录
	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	db.Create(&model.VisitStat{Date: yesterday, PV: 100, UV: 50, IP: 30})

	trend := svc.GetTrend(7)
	if len(trend) == 0 {
		t.Error("expected non-empty trend")
	}
}

func TestGetTopContent(t *testing.T) {
	db := testutil.TestDB(t)
	svc := analytics.NewService(db)
	testutil.SeedTestData(db)

	topVods := svc.GetTopContent("vod", 10)
	if len(topVods) != 2 {
		t.Errorf("expected 2 top vods, got %d", len(topVods))
	}
	// 应该按 hits 降序
	if len(topVods) >= 2 && topVods[0].Hits < topVods[1].Hits {
		t.Error("top vods should be sorted by hits desc")
	}
}

func TestAggregateDaily(t *testing.T) {
	db := testutil.TestDB(t)
	svc := analytics.NewService(db)

	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	db.Create(&model.Visit{URL: "/a", IP: "1.1.1.1", Date: yesterday, VisitTime: time.Now().AddDate(0, 0, -1).Unix(), IsUnique: 1})
	db.Create(&model.Visit{URL: "/b", IP: "2.2.2.2", Date: yesterday, VisitTime: time.Now().AddDate(0, 0, -1).Unix(), IsUnique: 1})

	if err := svc.AggregateDaily(); err != nil {
		t.Fatalf("AggregateDaily failed: %v", err)
	}

	var stat model.VisitStat
	db.Where("date = ?", yesterday).First(&stat)
	if stat.PV != 2 {
		t.Errorf("expected PV=2, got %d", stat.PV)
	}
}

func TestGetVisitList(t *testing.T) {
	db := testutil.TestDB(t)
	svc := analytics.NewService(db)

	db.Create(&model.Visit{URL: "/test", IP: "127.0.0.1", Date: "2026-01-01", VisitTime: 1000})

	visits, total, err := svc.GetVisitList(1, 10, "2026-01-01")
	if err != nil {
		t.Fatalf("GetVisitList failed: %v", err)
	}
	if total != 1 {
		t.Errorf("expected total=1, got %d", total)
	}
	if len(visits) != 1 {
		t.Errorf("expected 1 visit, got %d", len(visits))
	}
}
