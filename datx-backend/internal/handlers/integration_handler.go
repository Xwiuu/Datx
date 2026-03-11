package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/xwiuu/datx-backend/internal/database"
	"github.com/xwiuu/datx-backend/internal/models"
	"github.com/xwiuu/datx-backend/internal/services" // Para importar o disparo do FB
)

// UpdateFacebookIntegration salva ou edita o Pixel e Token do usuário
func UpdateFacebookIntegration(c *fiber.Ctx) error {
	userID := c.Params("user_id") // ID do usuário que tá logado/editando

	var payload struct {
		FbPixelID     string `json:"fb_pixel_id"`
		FbAccessToken string `json:"fb_access_token"`
	}

	if err := c.BodyParser(&payload); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Dados inválidos"})
	}

	// Atualiza APENAS os campos do Facebook no banco de dados
	result := database.DB.Model(&models.User{}).
		Where("id = ?", userID).
		Updates(map[string]interface{}{
			"fb_pixel_id":     payload.FbPixelID,
			"fb_access_token": payload.FbAccessToken,
		})

	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Erro ao salvar a integração"})
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "✅ Integração do Facebook salva com sucesso!",
	})
}

// TestFacebookConnection faz um disparo fake pra ver se as chaves funcionam
func TestFacebookConnection(c *fiber.Ctx) error {
	var payload struct {
		FbPixelID     string `json:"fb_pixel_id"`
		FbAccessToken string `json:"fb_access_token"`
	}

	if err := c.BodyParser(&payload); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Dados inválidos"})
	}

	// Monta um evento de teste na hora
	testData := map[string]interface{}{
		"ip":          c.IP(), // Pega o IP de quem apertou o botão
		"ua":          "DATX Connection Tester",
		"external_id": "TESTE-CONEXAO",
	}

	// Dispara a função real
	// PS: Vai disparar pro Webhook.site ou pro FB Oficial dependendo de como tá o seu service!
	services.PushToFacebook(payload.FbPixelID, payload.FbAccessToken, "TestEvent", testData, 0, "")

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "🔥 Evento de teste disparado com sucesso!",
	})
}
