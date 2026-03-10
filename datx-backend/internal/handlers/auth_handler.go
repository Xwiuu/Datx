package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/xwiuu/datx-backend/internal/auth"
	"github.com/xwiuu/datx-backend/internal/database"
	"github.com/xwiuu/datx-backend/internal/models"
)

func Register(c *fiber.Ctx) error {
	type SignupRequest struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	req := new(SignupRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	// 1. Encriptar a senha
	hashedPassword, _ := auth.HashPassword(req.Password)

	// 2. Criar objeto do usuário
	user := models.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: hashedPassword,
	}

	// 3. Salvar no Banco (Usando o DB que vamos configurar)
	if err := database.DB.Create(&user).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Could not create user"})
	}

	return c.Status(201).JSON(fiber.Map{"message": "User registered successfully"})
}

func Login(c *fiber.Ctx) error {
	type LoginRequest struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	req := new(LoginRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Requisição inválida"})
	}

	// 1. Buscar o usuário no banco pelo e-mail
	var user models.User
	if err := database.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "Credenciais inválidas"})
	}

	// 2. Verificar se a senha digitada bate com o Hash do banco
	if !auth.CheckPasswordHash(req.Password, user.Password) {
		return c.Status(401).JSON(fiber.Map{"error": "Credenciais inválidas"})
	}

	// 3. Gerar o Token JWT de acesso
	token, err := auth.GenerateToken(user.Email)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Erro ao gerar token de acesso"})
	}

	// 4. Retornar os dados (exceto senha) e o Token
	return c.JSON(fiber.Map{
		"token": token,
		"user": fiber.Map{
			"id":    user.ID,
			"name":  user.Name,
			"email": user.Email,
			"plan":  user.Plan,
		},
	})
}
