package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/xwiuu/datx-backend/internal/database"
	"github.com/xwiuu/datx-backend/internal/models"
)

type DashboardProStats struct {
	Impressions    int64   `json:"impressions"`     // Vamos simular ou puxar do Meta
	Clicks         int64   `json:"clicks"`          // Nossos Cliques Humanos
	Checkouts      int64   `json:"checkouts"`       // Quem chegou no checkout
	Sales          int64   `json:"sales"`           // Quem pagou (PAID)
	Revenue        float64 `json:"revenue"`         // Faturamento Aprovado
	PendingRevenue float64 `json:"pending_revenue"` // Boleto/PIX
	AdSpend        float64 `json:"ad_spend"`        // Gasto (Imputado ou API)

	// Métricas Calculadas
	CTR  float64 `json:"ctr"`
	CPA  float64 `json:"cpa"`
	ROI  float64 `json:"roi"`
	ROAS float64 `json:"roas"`
}

func GetDashboardPro(c *fiber.Ctx) error {
	var stats DashboardProStats

	// 1. Puxa Faturamento e Vendas (PAID)
	database.DB.Model(&models.ClickLog{}).
		Where("payment_status = ?", "paid").
		Select("COALESCE(SUM(conversion_value), 0) as revenue, COUNT(id) as sales").
		Scan(&stats)

	// 2. Puxa Faturamento Pendente
	database.DB.Model(&models.ClickLog{}).
		Where("payment_status = ?", "pending").
		Select("COALESCE(SUM(conversion_value), 0)").
		Scan(&stats.PendingRevenue)

	// 3. Puxa Cliques Humanos e Checkouts
	database.DB.Model(&models.Link{}).Select("SUM(human_visits)").Scan(&stats.Clicks)
	database.DB.Model(&models.ClickLog{}).Where("reached_checkout = ?", true).Count(&stats.Checkouts)

	// --- LÓGICA DE NEGÓCIO (Cálculos) ---

	// Simulação de Gasto (Para teste, amanhã fazemos o input real)
	stats.AdSpend = 500.00
	stats.Impressions = stats.Clicks * 45 // Estimativa baseada em 2.2% de CTR médio

	if stats.Impressions > 0 {
		stats.CTR = (float64(stats.Clicks) / float64(stats.Impressions)) * 100
	}
	if stats.Sales > 0 {
		stats.CPA = stats.AdSpend / float64(stats.Sales)
	}
	if stats.AdSpend > 0 {
		stats.ROAS = stats.Revenue / stats.AdSpend
		stats.ROI = ((stats.Revenue - stats.AdSpend) / stats.AdSpend) * 100
	}

	return c.JSON(stats)
}
