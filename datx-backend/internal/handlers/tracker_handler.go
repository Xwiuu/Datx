package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/xwiuu/datx-backend/internal/database"
	"github.com/xwiuu/datx-backend/internal/models"
)

// RegisterClick recebe e processa todos os eventos do Pixel DATX
func RegisterClick(c *fiber.Ctx) error {
	// A estrutura que bate com o payload do nosso Pixel JS
	type TrackPayload struct {
		Event       string `json:"event"` // PageView, InitiateCheckout, etc.
		URL         string `json:"url"`
		UtmSource   string `json:"utm_source"`
		UtmCampaign string `json:"utm_campaign"`
		UtmContent  string `json:"utm_content"`
		Fbclid      string `json:"fbclid"`
		Fbp         string `json:"fbp"`
		Fbc         string `json:"fbc"`
		ExternalID  string `json:"external_id"`
		Device      string `json:"device"`
	}

	var payload TrackPayload
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Payload inválido ou corrompido"})
	}

	// Segurança básica: ignorar requisições vazias
	if payload.ExternalID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "External ID obrigatório"})
	}

	var click models.ClickLog

	// 🧠 MATCHING ENGINE: Verifica se esse usuário já existe no banco (mesma sessão)
	result := database.DB.Where("external_id = ?", payload.ExternalID).First(&click)

	if result.Error != nil {
		// --- NOVO USUÁRIO (PRIMEIRO ACESSO) ---
		click = models.ClickLog{
			IPAddress:   c.IP(),
			UserAgent:   string(c.Request().Header.UserAgent()),
			Device:      payload.Device,
			UtmSource:   payload.UtmSource,
			UtmCampaign: payload.UtmCampaign,
			UtmContent:  payload.UtmContent,
			Fbclid:      payload.Fbclid,
			Fbp:         payload.Fbp,
			Fbc:         payload.Fbc,
			ExternalID:  payload.ExternalID,
			// Aqui você pode adicionar um campo LastEvent no seu models.ClickLog se quiser rastrear o funil
		}

		if err := database.DB.Create(&click).Error; err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Falha ao gravar inteligência na base"})
		}

	} else {
		// --- USUÁRIO RETORNANDO OU AVANÇANDO NO FUNIL ---
		// Exemplo: O cara deu PageView antes, e agora clicou no botão (InitiateCheckout)

		// Atualiza os dados caso o Pixel tenha conseguido capturar o FBC/FBP mais tarde
		updates := map[string]interface{}{}

		if payload.Fbp != "" && click.Fbp == "" {
			updates["fbp"] = payload.Fbp
		}
		if payload.Fbc != "" && click.Fbc == "" {
			updates["fbc"] = payload.Fbc
		}
		// Se você adicionar o campo LastEvent na struct, você atualiza ele aqui:
		// updates["last_event"] = payload.Event

		if len(updates) > 0 {
			database.DB.Model(&click).Updates(updates)
		}
	}

	// 🎯 O TIRO DE SNIPER: Devolvemos o click_id UUID do banco para o Frontend
	// O Pixel vai pegar esse ID e colar no link da Hotmart/Kiwify (?src=ID)
	return c.Status(201).JSON(fiber.Map{
		"status":   "tracked",
		"event":    payload.Event,
		"click_id": click.ID.String(),
	})
}
