package handlers

import (
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/xwiuu/datx-backend/internal/database"
	"github.com/xwiuu/datx-backend/internal/models"
	"github.com/xwiuu/datx-backend/internal/services"
)

// NormalizedConversion é a estrutura padrão do DATX.
// O Tradutor converte qualquer doideira de qualquer gateway para isso aqui.
type NormalizedConversion struct {
	Gateway       string
	TransactionID string
	RawStatus     string  // Status original que veio do Gateway (ex: "waiting_payment")
	Amount        float64 // Valor (Lucro Líquido ou Valor do Boleto)
	ClickID       string  // Nosso UUID que o Pixel injetou no link
}

// GatewayReceiver escuta TODOS os checkouts na rota /webhooks/:gateway
func GatewayReceiver(c *fiber.Ctx) error {
	gateway := strings.ToLower(c.Params("gateway"))
	var data NormalizedConversion

	// ---------------------------------------------------------
	// 🔌 TRADUTORES DE GATEWAYS (ADAPTERS)
	// ---------------------------------------------------------

	switch gateway {
	case "kiwify":
		var payload struct {
			OrderID     string `json:"order_id"`
			OrderStatus string `json:"order_status"`
			Commissions struct {
				Liquid float64 `json:"liquid"`
			} `json:"Commissions"`
			TrackingParameters struct {
				Src string `json:"src"`
				Sck string `json:"sck"`
			} `json:"tracking_parameters"`
		}
		if err := c.BodyParser(&payload); err != nil {
			return c.SendStatus(400)
		}
		data.Gateway = "Kiwify"
		data.TransactionID = payload.OrderID
		data.RawStatus = payload.OrderStatus
		data.Amount = payload.Commissions.Liquid
		data.ClickID = payload.TrackingParameters.Src
		if data.ClickID == "" {
			data.ClickID = payload.TrackingParameters.Sck
		}

	case "hotmart":
		var payload struct {
			Data struct {
				Purchase struct {
					Transaction string `json:"transaction"`
					Status      string `json:"status"`
					Price       struct {
						Value float64 `json:"value"`
					} `json:"price"`
				} `json:"purchase"`
				Tracking struct {
					Source string `json:"source"`
				} `json:"tracking"`
			} `json:"data"`
		}
		if err := c.BodyParser(&payload); err != nil {
			return c.SendStatus(400)
		}
		data.Gateway = "Hotmart"
		data.TransactionID = payload.Data.Purchase.Transaction
		data.RawStatus = payload.Data.Purchase.Status
		data.Amount = payload.Data.Purchase.Price.Value
		data.ClickID = payload.Data.Tracking.Source

	case "payevo":
		var payload struct {
			ID       string  `json:"id"`
			State    string  `json:"state"`
			NetValue float64 `json:"net_value"`
			Utm      struct {
				Src string `json:"src"`
			} `json:"utm"`
		}
		if err := c.BodyParser(&payload); err != nil {
			return c.SendStatus(400)
		}
		data.Gateway = "PayEvo"
		data.TransactionID = payload.ID
		data.RawStatus = payload.State
		data.Amount = payload.NetValue
		data.ClickID = payload.Utm.Src

	case "luna":
		var payload struct {
			TransactionID string  `json:"transaction_id"`
			PaymentStatus string  `json:"payment_status"`
			LiquidAmount  float64 `json:"liquid_amount"`
			AffiliateData struct {
				SubID string `json:"sub_id"`
			} `json:"affiliate_data"`
		}
		if err := c.BodyParser(&payload); err != nil {
			return c.SendStatus(400)
		}
		data.Gateway = "Luna"
		data.TransactionID = payload.TransactionID
		data.RawStatus = payload.PaymentStatus
		data.Amount = payload.LiquidAmount
		data.ClickID = payload.AffiliateData.SubID

	case "vegacheckout":
		var payload struct {
			Code      string  `json:"code"`
			Status    string  `json:"status"`
			Comission float64 `json:"comission"`
			Src       string  `json:"src"`
		}
		if err := c.BodyParser(&payload); err != nil {
			return c.SendStatus(400)
		}
		data.Gateway = "VegaCheckout"
		data.TransactionID = payload.Code
		data.RawStatus = payload.Status
		data.Amount = payload.Comission
		data.ClickID = payload.Src

	default:
		fmt.Println("Gateway não suportado:", gateway)
		return c.Status(404).SendString("Gateway não integrado no Datx")
	}

	// ---------------------------------------------------------
	// 🧠 MOTOR CENTRAL DE MATCHING E ATRIBUIÇÃO
	// ---------------------------------------------------------

	if data.ClickID == "" {
		fmt.Printf("[%s] Evento Orgânico (Sem Rastreio) | Status: %s\n", data.Gateway, data.RawStatus)
		return c.SendStatus(200)
	}

	var click models.ClickLog
	if err := database.DB.Where("id = ?", data.ClickID).First(&click).Error; err != nil {
		fmt.Println("❌ MATCH FALHOU: O Click ID não existe na nossa base:", data.ClickID)
		return c.SendStatus(200)
	}

	// 1. Padronizador de Status Master
	statusLower := strings.ToLower(data.RawStatus)
	finalStatus := "pending"

	if statusLower == "paid" || statusLower == "approved" || statusLower == "pagamento_aprovado" || statusLower == "complete" {
		finalStatus = "paid"
	} else if statusLower == "refunded" || statusLower == "chargeback" || statusLower == "devolvido" {
		finalStatus = "refunded"
	} else if statusLower == "waiting_payment" || statusLower == "pending" || statusLower == "boleto_impresso" || statusLower == "pix_generated" || statusLower == "checkout_initiated" {
		finalStatus = "pending"
	} else if statusLower == "refused" || statusLower == "canceled" || statusLower == "recusado" {
		finalStatus = "canceled"
	}

	// 2. Prepara os dados para salvar no banco
	updates := map[string]interface{}{
		"payment_status":   finalStatus,
		"transaction_id":   data.TransactionID,
		"reached_checkout": true, // Se chegou no webhook, passou pelo checkout
	}

	if finalStatus == "paid" {
		updates["converted"] = true
		updates["conversion_value"] = data.Amount
	} else if finalStatus == "refunded" {
		updates["converted"] = false
	}

	if click.LinkID != [16]byte{} { // Verifica se temos o Link vinculado
		var link models.Link
		database.DB.First(&link, "id = ?", click.LinkID)

		if link.FbPixelID != "" && link.FbAccessToken != "" {
			// Mapeia dados para o disparador
			clickMap := map[string]interface{}{
				"ip":          click.IPAddress,
				"ua":          click.UserAgent,
				"fbp":         click.Fbclid, // Usando o que temos no modelo
				"fbc":         click.Fbc,
				"external_id": click.ID.String(),
				"page_url":    link.PageURL,
			}

			// Disparo em background para não travar o Webhook (Go Routine)
			if finalStatus == "paid" {
				go services.PushToFacebook(link.FbPixelID, link.FbAccessToken, "Purchase", clickMap, data.Amount, "")
			} else if finalStatus == "pending" {
				go services.PushToFacebook(link.FbPixelID, link.FbAccessToken, "Lead", clickMap, data.Amount, "")
			}
		}
	}

	// Salva no banco
	database.DB.Model(&click).Updates(updates)

	fmt.Printf("🎯 [DATX FUNIL] %s | %s | R$ %.2f | Campanha: %s\n",
		data.Gateway, strings.ToUpper(finalStatus), data.Amount, click.UtmCampaign)

	/// ---------------------------------------------------------
	// 🚀 CAPI ENTRYPOINT (Buscando o Dono do Link)
	// ---------------------------------------------------------

	if click.LinkID != [16]byte{} { // Se temos um link atrelado ao clique...
		// 1. Acha o Link no banco
		var link models.Link
		database.DB.First(&link, "id = ?", click.LinkID)

		// 2. Acha o Usuário (Dono) daquele Link
		var user models.User
		database.DB.First(&user, "id = ?", link.UserID)

		// 3. Verifica se o usuário integrou o Facebook
		if user.FbPixelID != "" && user.FbAccessToken != "" {
			// Monta os dados com alta qualidade
			clickMap := map[string]interface{}{
				"ip":          click.IPAddress,
				"ua":          click.UserAgent,
				"fbp":         click.Fbp,
				"fbc":         click.Fbc,
				"external_id": click.ID.String(),
				"page_url":    link.PageURL,
			}

			// Dispara o evento
			if finalStatus == "paid" {
				go services.PushToFacebook(user.FbPixelID, user.FbAccessToken, "Purchase", clickMap, data.Amount, "")
			} else if finalStatus == "pending" {
				go services.PushToFacebook(user.FbPixelID, user.FbAccessToken, "Lead", clickMap, data.Amount, "")
			}
		} else {
			fmt.Println("⚠️ Link pertence a um usuário sem Facebook configurado. CAPI ignorado.")
		}
	}
	return c.SendStatus(200)
}
