package handlers

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/xwiuu/datx-backend/internal/database"
	"github.com/xwiuu/datx-backend/internal/models"
)

// RecordEvent atualiza os contadores da campanha e gera os dados do gráfico
func RecordEvent(c *fiber.Ctx) error {
	slug := c.Params("slug")
	eventType := c.Query("type") // "human" ou "bot"

	if slug == "" || eventType == "" {
		return c.SendStatus(400)
	}

	// 1. Atualiza o contador GLOBAL da campanha (para a tabela e KPIs)
	if eventType == "human" {
		database.DB.Model(&models.Link{}).Where("slug = ?", slug).
			UpdateColumn("human_visits", database.DB.Raw("human_visits + 1"))
	} else {
		database.DB.Model(&models.Link{}).Where("slug = ?", slug).
			UpdateColumn("bot_blocked", database.DB.Raw("bot_blocked + 1"))
	}

	// 2. Registra no Histórico Horário (para o Gráfico de Linha)
	now := time.Now()
	currentHour := now.Format("15:00")
	currentDate := now.Format("2006-01-02")

	var stats models.LinkStats
	// Procura se já existe registro para esta hora/dia
	err := database.DB.Where("slug = ? AND hour = ? AND date = ?", slug, currentHour, currentDate).First(&stats).Error

	if err != nil {
		// Se não existe, cria o primeiro registro da hora
		newStats := models.LinkStats{
			Slug: slug,
			Hour: currentHour,
			Date: currentDate,
		}
		if eventType == "human" {
			newStats.HumanCount = 1
		} else {
			newStats.BotCount = 1
		}
		database.DB.Create(&newStats)
	} else {
		// Se já existe, dá o +1 na coluna correta
		if eventType == "human" {
			database.DB.Model(&stats).UpdateColumn("human_count", database.DB.Raw("human_count + 1"))
		} else {
			database.DB.Model(&stats).UpdateColumn("bot_count", database.DB.Raw("bot_count + 1"))
		}
	}

	return c.SendStatus(200)
}
