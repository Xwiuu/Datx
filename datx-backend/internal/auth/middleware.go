package auth

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

func Protected() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Pega o header "Authorization"
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(401).JSON(fiber.Map{"error": "Token não fornecido"})
		}

		// Remove o "Bearer " e pega só o token
		tokenString := strings.Replace(authHeader, "Bearer ", "", 1)

		// Valida o token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return jwtSecret, nil
		})

		if err != nil || !token.Valid {
			return c.Status(401).JSON(fiber.Map{"error": "Token inválido ou expirado"})
		}

		return c.Next() // Se estiver tudo OK, segue para a rota
	}
}
