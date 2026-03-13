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
type NormalizedConversion struct {
	Gateway       string
	TransactionID string
	RawStatus     string
	Amount        float64
	ClickID       string
}

// GatewayReceiver escuta TODOS os checkouts na rota /webhooks/:gateway
func GatewayReceiver(c *fiber.Ctx) error {
	gateway := strings.ToLower(c.Params("gateway"))
	var data NormalizedConversion

	// 1. Pegamos o corpo bruto para o scanner vasculhar se as UTMs falharem
	var rawBody map[string]interface{}
	if err := c.BodyParser(&rawBody); err != nil {
		fmt.Println("DATX: Erro ao ler body bruto para o detetive")
	}

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
		c.BodyParser(&payload)
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
		c.BodyParser(&payload)
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
		c.BodyParser(&payload)
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
		c.BodyParser(&payload)
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
		c.BodyParser(&payload)
		data.Gateway = "VegaCheckout"
		data.TransactionID = payload.Code
		data.RawStatus = payload.Status
		data.Amount = payload.Comission
		data.ClickID = payload.Src

	default:
		data.Gateway = "Generic-" + gateway
	}

	// ---------------------------------------------------------
	// 🕵️ O PULO DO GATO: Se os campos oficiais falharam, scaneia o JSON
	// ---------------------------------------------------------
	if data.ClickID == "" {
		data.ClickID = findClickIDInMap(rawBody)
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
		fmt.Println("❌ MATCH FALHOU: ClickID não existe na base:", data.ClickID)
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

	// 2. Prepara os dados para salvar
	updates := map[string]interface{}{
		"payment_status":   finalStatus,
		"transaction_id":   data.TransactionID,
		"reached_checkout": true,
	}

	if finalStatus == "paid" {
		updates["converted"] = true
		updates["conversion_value"] = data.Amount
	} else if finalStatus == "refunded" {
		updates["converted"] = false
	}

	// Salva no banco
	database.DB.Model(&click).Updates(updates)

	fmt.Printf("🎯 [DATX FUNIL] %s | %s | R$ %.2f | Campanha: %s\n",
		data.Gateway, strings.ToUpper(finalStatus), data.Amount, click.UtmCampaign)

	// ---------------------------------------------------------
	// 🚀 CAPI ENTRYPOINT (Facebook Conversion API)
	// ---------------------------------------------------------
	if click.LinkID != [16]byte{} {
		var link models.Link
		database.DB.First(&link, "id = ?", click.LinkID)

		var user models.User
		database.DB.First(&user, "id = ?", link.UserID)

		if user.FbPixelID != "" && user.FbAccessToken != "" {
			clickMap := map[string]interface{}{
				"ip":          click.IPAddress,
				"ua":          click.UserAgent,
				"fbp":         click.Fbp,
				"fbc":         click.Fbc,
				"external_id": click.ID.String(),
				"page_url":    link.PageURL,
			}

			if finalStatus == "paid" {
				go services.PushToFacebook(user.FbPixelID, user.FbAccessToken, "Purchase", clickMap, data.Amount, "")
			} else if finalStatus == "pending" {
				go services.PushToFacebook(user.FbPixelID, user.FbAccessToken, "Lead", clickMap, data.Amount, "")
			}
		}
	}

	return c.SendStatus(200)
}

// 🛡️ Função Auxiliar para o Scanner (Deep Tracking)
func findClickIDInMap(data map[string]interface{}) string {
	searchFields := []string{"datx_id", "external_id", "custom_field", "sub_id", "tracking_id", "metadata"}
	for _, field := range searchFields {
		if val, ok := data[field]; ok {
			// Se o campo for um objeto aninhado (ex: metadata: { datx_id: "..." })
			if m, isMap := val.(map[string]interface{}); isMap {
				if id, exists := m["datx_id"]; exists {
					return fmt.Sprintf("%v", id)
				}
				if id, exists := m["external_id"]; exists {
					return fmt.Sprintf("%v", id)
				}
			}
			// Se for o campo direto
			return fmt.Sprintf("%v", val)
		}
	}
	return ""
}
