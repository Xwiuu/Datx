package handlers

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/xwiuu/datx-backend/internal/database"
	"github.com/xwiuu/datx-backend/internal/models"
)

// CreateLink - Cria a campanha no banco de dados (Deploy Shield)
func CreateLink(c *fiber.Ctx) error {
	var link models.Link

	// 1. Lê o JSON que vem do Postman ou Front-end
	if err := c.BodyParser(&link); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "JSON inválido. Verifique os dados enviados."})
	}

	// 2. 🛡️ O PULO DO GATO (Resolve o erro do UserID nulo)
	// Pega o primeiro usuário cadastrado no banco de dados e atribui a ele
	var user models.User
	if err := database.DB.First(&user).Error; err != nil {
		fmt.Println("❌ Nenhum usuário no banco para ser o dono do link.")
		return c.Status(500).JSON(fiber.Map{"error": "Crie pelo menos um usuário no banco primeiro!"})
	}
	link.UserID = user.ID // Define de quem é esse link

	// 3. Salva o Link (Shield) no banco
	if err := database.DB.Create(&link).Error; err != nil {
		fmt.Println("❌ Erro DB ao salvar Link:", err)
		return c.Status(500).JSON(fiber.Map{"error": "Erro ao salvar shield no arsenal"})
	}

	fmt.Println("✅ Link criado com sucesso! ID:", link.ID)
	return c.Status(201).JSON(link)
}

// GetMyLinks - Lista as campanhas no painel (Alimenta a Tabela de Logs & Match)
func GetMyLinks(c *fiber.Ctx) error {
	var links []models.Link

	// Puxa todo o arsenal do banco, ordenando do mais recente pro mais antigo
	result := database.DB.Order("created_at desc").Find(&links)

	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Erro ao buscar arsenal de blindagem no banco"})
	}

	// Devolve a lista de campanhas em formato JSON para o Front-end
	return c.Status(200).JSON(links)
}

// TraceHardware - Rota para o coletor de hardware (ml_shield)
func TraceHardware(c *fiber.Ctx) error {
	// Aqui no futuro entra a lógica de salvar a fingerprint de GPU/CPU de bots avançados
	return c.Status(200).JSON(fiber.Map{"status": "Hardware trace received"})
}

func GetLinkAnalytics(c *fiber.Ctx) error {
	var stats []models.LinkStats
	// Pega os dados das últimas 24h ou do dia atual
	database.DB.Order("hour asc").Find(&stats)
	return c.JSON(stats)
}
func UpdateLinkCAPI(c *fiber.Ctx) error {
	id := c.Params("id")
	var input struct {
		PixelID     string `json:"fb_pixel_id"`
		AccessToken string `json:"fb_access_token"`
		PageURL     string `json:"page_url"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "JSON inválido"})
	}

	database.DB.Model(&models.Link{}).Where("id = ?", id).Updates(input)
	return c.JSON(fiber.Map{"status": "Configuração de CAPI salva!"})
}
