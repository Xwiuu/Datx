package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/xwiuu/datx-backend/internal/database"
	"github.com/xwiuu/datx-backend/internal/models"
)

// CreateLink - Cria a campanha no banco de dados (Deploy Shield)
func CreateLink(c *fiber.Ctx) error {
	var link models.Link

	if err := c.BodyParser(&link); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Dados inválidos enviados no payload"})
	}

	// HACK DE DESENVOLVIMENTO: Tenta achar um usuário. Se não achar, cria um "Fantasma"
	var user models.User
	if err := database.DB.First(&user).Error; err != nil {
		user = models.User{
			ID:    uuid.New(),
			Name:  "Commander",
			Email: "commander@datx.ai",
		}
		// Cria o usuário fantasma no banco
		database.DB.Create(&user)
	}

	// Agora temos certeza que existe um usuário! Acopla o Shield a ele.
	link.ID = uuid.New()
	link.UserID = user.ID
	link.Status = "active" // Garante que nasce ativo pronto pra guerra

	if err := database.DB.Create(&link).Error; err != nil {
		// Se der erro de "duplicate key" no slug, o GORM vai avisar aqui
		return c.Status(500).JSON(fiber.Map{"error": "Erro ao salvar shield no arsenal"})
	}

	// Responde com sucesso e devolve os dados criados (incluindo o Slug)
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
