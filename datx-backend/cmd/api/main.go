package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors" // IMPORTANTE: Adicionar esse import
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/xwiuu/datx-backend/internal/database"
	"github.com/xwiuu/datx-backend/internal/handlers"
	"github.com/xwiuu/datx-backend/internal/models"
	"github.com/xwiuu/datx-backend/internal/services"
)

func main() {
	// 1. Conectar ao Banco
	database.ConnectDB()

	// 2. Rodar as Migrations (Isso cria/atualiza as colunas no Postgres automaticamente)
	log.Println("🛠️  Running Database Migrations...")
	err := database.DB.AutoMigrate(&models.Link{}) // Certifique-se que models.Link tem os campos novos
	if err != nil {
		log.Fatal("Failed to migrate database: ", err)
	}

	app := fiber.New()

	// 3. Configurar CORS (Permitir que o seu Front na porta 3000 acesse a API)
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "http://localhost:3000", // URL do seu Next.js
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowMethods:     "GET, POST, PUT, DELETE, OPTIONS",
		AllowCredentials: true,
	}))

	// Middleware de Log para debug no terminal
	app.Use(logger.New())

	// --- ROTAS DO SISTEMA ---

	// Rotas da API (Inventory e Criação)
	api := app.Group("/v1")
	api.Post("/links", handlers.CreateLink) // Deploy de novos links
	api.Get("/links", handlers.GetMyLinks)  // Shadow Inventory (Listagem)

	api.Get("/shield/:slug", handlers.ServeShieldScript)
	api.Post("/shield/trace/:slug", handlers.RegisterTrace)

	// Rota de Trace para o ML Fingerprinting (Hardware Check)
	api.Post("/trace-hardware", handlers.TraceHardware)

	api.Post("/track", handlers.RegisterClick)
	api.Post("/webhooks/:gateway", handlers.GatewayReceiver)
	api.Get("/analytics", handlers.GetLinkAnalytics)
	api.Get("/shield/event/:slug", handlers.RecordEvent)

	// Rota do Dashboard
	api.Get("/dashboard/summary", handlers.GetDashboardPro)

	api.Get("/auth/facebook", handlers.ConnectFacebook)
	api.Get("/auth/facebook/callback", handlers.FacebookCallback)
	api.Patch("/links/:id/capi", handlers.UpdateLinkCAPI)

	api.Get("/debug/test-capi", func(c *fiber.Ctx) error {
		// Dados fakes para o teste
		clickMap := map[string]interface{}{
			"ip":          "127.0.0.1",
			"ua":          "Mozilla/5.0 (Windows NT 10.0; Win64; x64)",
			"fbp":         "fb.1.123456789",
			"fbc":         "fb.1.987654321",
			"external_id": "ID-DE-TESTE-UUID",
			"page_url":    "https://datx-teste.com",
		}

		// DISPARA PRO SEU WEBHOOK.SITE
		// O PixelID e Token podem ser qualquer coisa aqui já que a URL no service está para o webhook.site
		services.PushToFacebook("12345", "token_abc", "Purchase", clickMap, 197.00, "teste@gmail.com")

		return c.SendString("🚀 Comando enviado! Olha lá no Webhook.site")
	})

	// Rotas de Integração (Painel do Usuário)
	api.Put("/users/:user_id/integrations/facebook", handlers.UpdateFacebookIntegration)
	api.Post("/integrations/facebook/test", handlers.TestFacebookConnection)

	log.Fatal(app.Listen(":8080"))
}
