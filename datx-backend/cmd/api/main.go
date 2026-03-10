package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors" // IMPORTANTE: Adicionar esse import
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/xwiuu/datx-backend/internal/database"
	"github.com/xwiuu/datx-backend/internal/handlers"
	"github.com/xwiuu/datx-backend/internal/models"
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

	log.Fatal(app.Listen(":8080"))
}
