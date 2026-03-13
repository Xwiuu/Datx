package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/helmet" // 🛡️ Importamos o Helmet para Segurança XSS
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/xwiuu/datx-backend/internal/auth" // 🔒 Importamos o Middleware de Auth
	"github.com/xwiuu/datx-backend/internal/database"
	"github.com/xwiuu/datx-backend/internal/handlers"
	"github.com/xwiuu/datx-backend/internal/models"
	"github.com/xwiuu/datx-backend/internal/services"
)

func main() {
	// 1. Conectar ao Banco
	database.ConnectDB()

	log.Println("🛠️  Running Database Migrations...")
	err := database.DB.AutoMigrate(&models.Link{})
	if err != nil {
		log.Fatal("Failed to migrate database: ", err)
	}

	app := fiber.New()

	// 2. 🛡️ ATIVAÇÃO DO HELMET (Proteção contra XSS, Clickjacking e Sniffing)
	app.Use(helmet.New())

	// 3. Configurar CORS (Permitir que o seu Front na porta 3000 acesse a API)
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "http://localhost:3000",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowMethods:     "GET, POST, PUT, DELETE, OPTIONS",
		AllowCredentials: true,
	}))

	app.Use(logger.New())

	// --- ROTAS DO SISTEMA ---
	api := app.Group("/v1")

	// ==========================================
	// 🟢 ROTAS PÚBLICAS (Não precisam de Login)
	// ==========================================

	// Autenticação / Registo
	api.Post("/auth/register", handlers.Register)
	api.Post("/auth/login", handlers.Login)

	// O Motor do Shield e Tracking (Tem que ser público para os clientes acessarem o link)
	api.Get("/shield/:slug", handlers.ServeShieldScript)
	api.Post("/shield/trace/:slug", handlers.RegisterTrace)
	api.Post("/trace-hardware", handlers.TraceHardware)
	api.Post("/track", handlers.RegisterClick)
	api.Post("/webhooks/:gateway", handlers.GatewayReceiver)
	api.Get("/shield/event/:slug", handlers.RecordEvent)

	// Facebook Auth (A troca de chaves ocorre publicamente)
	api.Get("/auth/facebook", handlers.ConnectFacebook)
	api.Get("/auth/facebook/callback", handlers.FacebookCallback)

	// ==========================================
	// 🔴 ROTAS PRIVADAS (O Cadeado - Exigem JWT)
	// ==========================================
	// Qualquer rota pendurada neste grupo vai passar pelo auth.Protected()
	protected := api.Group("/", auth.Protected())

	// Dashboard e Analytics
	protected.Get("/dashboard/summary", handlers.GetDashboardPro)
	protected.Get("/analytics", handlers.GetLinkAnalytics)

	// Gestão de Links (Campanhas)
	protected.Post("/links", handlers.CreateLink)
	protected.Get("/links", handlers.GetMyLinks)
	protected.Patch("/links/:id/capi", handlers.UpdateLinkCAPI)

	// Integrações e Testes
	protected.Put("/users/:user_id/integrations/facebook", handlers.UpdateFacebookIntegration)
	protected.Post("/integrations/facebook/test", handlers.TestFacebookConnection)

	api.Get("/debug/test-capi", func(c *fiber.Ctx) error {
		clickMap := map[string]interface{}{
			"ip":          "127.0.0.1",
			"ua":          "Mozilla/5.0 (Windows NT 10.0; Win64; x64)",
			"fbp":         "fb.1.123456789",
			"fbc":         "fb.1.987654321",
			"external_id": "ID-DE-TESTE-UUID",
			"page_url":    "https://datx-teste.com",
		}
		services.PushToFacebook("12345", "token_abc", "Purchase", clickMap, 197.00, "teste@gmail.com")
		return c.SendString("🚀 Comando enviado! Olha lá no Webhook.site")
	})

	log.Fatal(app.Listen(":8080"))
}
