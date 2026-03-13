package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/xwiuu/datx-backend/internal/database"
	"github.com/xwiuu/datx-backend/internal/models"
)

// Estruturas para os gráficos
type ChartPoint struct {
	Name  string  `json:"name"`
	Valor float64 `json:"valor"`
}

// 🔥 ATUALIZADO: Agora com Revenue (Faturamento) por UTM
type UTMStat struct {
	Name    string  `json:"name"`
	Vendas  int     `json:"vendas"`
	Revenue float64 `json:"revenue"`
}

type DashboardProStats struct {
	Impressions    int64   `json:"impressions"`
	Clicks         int64   `json:"clicks"`
	Checkouts      int64   `json:"checkouts"`
	Sales          int64   `json:"sales"`
	Revenue        float64 `json:"revenue"`
	PendingRevenue float64 `json:"pending_revenue"`
	AdSpend        float64 `json:"ad_spend"`

	CTR  float64 `json:"ctr"`
	CPA  float64 `json:"cpa"`
	ROI  float64 `json:"roi"`
	ROAS float64 `json:"roas"`

	ChargebackRate float64      `json:"chargeback_rate" gorm:"-"`
	ApprovedCount  int64        `json:"approved_count" gorm:"-"`
	PendingCount   int64        `json:"pending_count" gorm:"-"`
	RefundedCount  int64        `json:"refunded_count" gorm:"-"`
	TopAds         []UTMStat    `json:"top_ads" gorm:"-"`
	ChartData      []ChartPoint `json:"chart_data" gorm:"-"`
}

func GetDashboardPro(c *fiber.Ctx) error {
	var stats DashboardProStats

	// 1. Puxa Faturamento Total e Vendas
	database.DB.Model(&models.ClickLog{}).Where("payment_status = ?", "paid").
		Select("COALESCE(SUM(conversion_value), 0) as revenue, COUNT(id) as sales").Scan(&stats)

	// 2. Puxa Faturamento Pendente
	database.DB.Model(&models.ClickLog{}).Where("payment_status = ?", "pending").
		Select("COALESCE(SUM(conversion_value), 0)").Scan(&stats.PendingRevenue)

	// 3. Cliques e Checkouts
	database.DB.Model(&models.Link{}).Select("COALESCE(SUM(human_visits), 0)").Scan(&stats.Clicks)
	database.DB.Model(&models.ClickLog{}).Where("reached_checkout = ?", true).Count(&stats.Checkouts)

	// 4. Status de Pagamentos (Pizza)
	database.DB.Model(&models.ClickLog{}).Where("payment_status = ?", "paid").Count(&stats.ApprovedCount)
	database.DB.Model(&models.ClickLog{}).Where("payment_status = ?", "pending").Count(&stats.PendingCount)
	database.DB.Model(&models.ClickLog{}).Where("payment_status = ?", "refunded").Count(&stats.RefundedCount)

	totalTransactions := stats.ApprovedCount + stats.PendingCount + stats.RefundedCount
	if totalTransactions > 0 {
		stats.ChargebackRate = (float64(stats.RefundedCount) / float64(totalTransactions)) * 100
	}

	// 5. 🎯 A MÁGICA: Query de Atribuição por UTM (Performance Real)
	database.DB.Model(&models.ClickLog{}).
		Select("utm_campaign as name, COUNT(id) as vendas, COALESCE(SUM(conversion_value), 0) as revenue").
		Where("payment_status = ? AND utm_campaign != ''", "paid").
		Group("utm_campaign").
		Order("revenue desc"). // Ordenamos por quem dá mais dinheiro!
		Limit(10).
		Scan(&stats.TopAds)

	if len(stats.TopAds) == 0 {
		stats.TopAds = []UTMStat{}
	}

	// 6. Dados de Gráfico (Segunda a Sexta - Placeholder)
	stats.ChartData = []ChartPoint{
		{Name: "Seg", Valor: 0}, {Name: "Ter", Valor: 0}, {Name: "Qua", Valor: 0},
		{Name: "Qui", Valor: 0}, {Name: "Sex", Valor: 0},
	}

	// Cálculos de ROI/CPA
	stats.AdSpend = 0.00
	stats.Impressions = stats.Clicks * 45
	if stats.Impressions > 0 {
		stats.CTR = (float64(stats.Clicks) / float64(stats.Impressions)) * 100
	}
	if stats.Sales > 0 {
		stats.CPA = stats.AdSpend / float64(stats.Sales)
	}
	if stats.AdSpend > 0 {
		stats.ROAS = stats.Revenue / stats.AdSpend
		stats.ROI = ((stats.Revenue - stats.AdSpend) / stats.AdSpend) * 100
	} else if stats.Revenue > 0 {
		stats.ROI = 100.00
	}

	return c.JSON(stats)
}
