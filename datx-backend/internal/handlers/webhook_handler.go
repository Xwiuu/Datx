package handlers

import (
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/xwiuu/datx-backend/internal/database"
	"github.com/xwiuu/datx-backend/internal/models"
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
		return c.SendStatus(200) // Retorna 200 pro gateway não ficar tentando reenviar
	}

	var click models.ClickLog
	if err := database.DB.Where("id = ?", data.ClickID).First(&click).Error; err != nil {
		fmt.Println("❌ MATCH FALHOU: O Click ID não existe na nossa base:", data.ClickID)
		return c.SendStatus(200)
	}

	// 1. Padronizador de Status Master
	statusLower := strings.ToLower(data.RawStatus)
	finalStatus := "pending" // Padrão

	if statusLower == "paid" || statusLower == "approved" || statusLower == "pagamento_aprovado" || statusLower == "complete" {
		finalStatus = "paid"
	} else if statusLower == "refunded" || statusLower == "chargeback" || statusLower == "devolvido" {
		finalStatus = "refunded"
	} else if statusLower == "waiting_payment" || statusLower == "pending" || statusLower == "boleto_impresso" || statusLower == "pix_generated" {
		finalStatus = "pending"
	} else if statusLower == "refused" || statusLower == "canceled" || statusLower == "recusado" {
		finalStatus = "canceled"
	}

	// 2. Prepara os dados para salvar no banco
	updates := map[string]interface{}{
		"payment_status": finalStatus,
		"transaction_id": data.TransactionID,
	}

	// Só consolida a conversão de ROI se for dinheiro no bolso
	if finalStatus == "paid" {
		updates["converted"] = true
		updates["conversion_value"] = data.Amount
	} else if finalStatus == "refunded" {
		// Se o cara pediu reembolso, a gente tira o "converted" pra não mascarar o ROI!
		updates["converted"] = false
	}

	// Salva no banco
	database.DB.Model(&click).Updates(updates)

	fmt.Printf("🎯 [DATX FUNIL] %s | %s | R$ %.2f | Campanha: %s\n",
		data.Gateway, strings.ToUpper(finalStatus), data.Amount, click.UtmCampaign)

	// 🚀 CAPI ENTRYPOINT:
	// if finalStatus == "paid" { Dispara evento Purchase }
	// if finalStatus == "pending" { Dispara evento GenerateLead ou AddToCart }

	return c.SendStatus(200)
}
