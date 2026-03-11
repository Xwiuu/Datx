package handlers

import (
	"fmt"
	"net/url"

	"github.com/gofiber/fiber/v2"
)

// Configurações do seu App no Facebook (Pegue no developers.facebook.com)
const (
	ClientID     = "SEU_APP_ID"
	ClientSecret = "SEU_APP_SECRET"
	RedirectURI  = "http://localhost:8080/v1/auth/facebook/callback"
)

// ConnectFacebook joga o usuário para a tela de login do Meta
func ConnectFacebook(c *fiber.Ctx) error {
	// Permissões necessárias para gerenciar Ads e disparar CAPI
	permissions := "ads_management,ads_read,business_management,public_profile"

	fbURL := fmt.Sprintf(
		"https://www.facebook.com/v18.0/dialog/oauth?client_id=%s&redirect_uri=%s&scope=%s&response_type=code",
		ClientID, url.QueryEscape(RedirectURI), permissions,
	)

	return c.Redirect(fbURL)
}

// FacebookCallback recebe o código do Facebook e troca pelo Access Token
func FacebookCallback(c *fiber.Ctx) error {
	code := c.Query("code")
	if code == "" {
		return c.Status(400).SendString("Autorização negada pelo usuário")
	}

	// Aqui a gente faz a troca do Code pelo Token (Vou te mandar a função de troca em seguida)
	fmt.Println("🚀 Código recebido do Facebook:", code)

	// Por enquanto, vamos avisar o Front que deu certo
	return c.Format("Opa! Facebook conectado. Agora o Go vai trocar esse código pelo seu Token de 60 dias.")
}
