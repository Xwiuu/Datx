package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/xwiuu/datx-backend/internal/database"
	"github.com/xwiuu/datx-backend/internal/models"
)

// GetProfile puxa os dados do utilizador logado e o seu Faturamento Total
func GetProfile(c *fiber.Ctx) error {
	// 🛡️ Blindagem: Verifica se o e-mail existe no contexto (Middleware)
	emailLocal := c.Locals("email")
	if emailLocal == nil {
		return c.Status(401).JSON(fiber.Map{"error": "Sessão inválida. Faça login novamente."})
	}
	email := emailLocal.(string)

	var user models.User
	if err := database.DB.Where("email = ?", email).First(&user).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Utilizador não encontrado"})
	}

	// 🏆 GAMIFICAÇÃO: Calcula o faturamento total aprovado
	var totalRevenue float64
	database.DB.Table("click_logs").
		Joins("JOIN links ON links.id = click_logs.link_id").
		Where("links.user_id = ? AND click_logs.payment_status = ?", user.ID, "paid").
		Select("COALESCE(SUM(click_logs.conversion_value), 0)").
		Scan(&totalRevenue)

	return c.JSON(fiber.Map{
		"id":              user.ID,
		"name":            user.Name,
		"email":           user.Email,
		"plan":            user.Plan,
		"fb_pixel_id":     user.FbPixelID,
		"fb_access_token": user.FbAccessToken,
		"total_revenue":   totalRevenue,
	})
}

// UpdateProfile permite ao utilizador mudar o Nome
func UpdateProfile(c *fiber.Ctx) error {
	emailLocal := c.Locals("email")
	if emailLocal == nil {
		return c.Status(401).JSON(fiber.Map{"error": "Sessão inválida"})
	}
	email := emailLocal.(string)

	var payload struct {
		Name string `json:"name"`
	}
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Dados inválidos"})
	}

	database.DB.Model(&models.User{}).Where("email = ?", email).Update("name", payload.Name)

	return c.JSON(fiber.Map{"status": "success", "message": "Perfil atualizado com sucesso!"})
}

// UpdateMyFacebook permite salvar as credenciais do Meta Ads
func UpdateMyFacebook(c *fiber.Ctx) error {
	emailLocal := c.Locals("email")
	if emailLocal == nil {
		return c.Status(401).JSON(fiber.Map{"error": "Sessão inválida"})
	}
	email := emailLocal.(string)

	var payload struct {
		FbPixelID     string `json:"fb_pixel_id"`
		FbAccessToken string `json:"fb_access_token"`
	}
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Dados inválidos"})
	}

	database.DB.Model(&models.User{}).Where("email = ?", email).Updates(map[string]interface{}{
		"fb_pixel_id":     payload.FbPixelID,
		"fb_access_token": payload.FbAccessToken,
	})

	return c.JSON(fiber.Map{"status": "success", "message": "Integração Meta Ads ativada!"})
}
