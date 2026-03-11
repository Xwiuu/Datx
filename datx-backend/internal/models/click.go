package models

import (
	"time"

	"github.com/google/uuid"
)

// ClickLog é o registro único de cada humano que passou pelo funil
type ClickLog struct {
	ID     uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primary_key" json:"click_id"` // O famoso click_id
	LinkID uuid.UUID `gorm:"type:uuid;not null;index" json:"link_id"`                         // De qual campanha veio

	// --- IDENTIFICAÇÃO BÁSICA ---
	IPAddress string `json:"ip_address"`
	UserAgent string `json:"user_agent"`
	Device    string `json:"device"` // mobile ou pc

	// --- RASTREIO DE ADS (UTMs) ---
	UtmSource   string `json:"utm_source"`
	UtmMedium   string `json:"utm_medium"`
	UtmCampaign string `json:"utm_campaign"`
	UtmContent  string `json:"utm_content"`
	UtmTerm     string `json:"utm_term"`

	// --- FACEBOOK CAPI (O OURO) ---
	Fbclid     string `gorm:"index" json:"fbclid"`
	Fbp        string `json:"fbp"`
	Fbc        string `json:"fbc"`
	ExternalID string `gorm:"index" json:"external_id"` // ID único que enviaremos pro Face depois

	// --- O MATCHING (Conversão) ---
	Converted       bool    `gorm:"default:false;index" json:"converted"`       // Virou false no clique. Vai virar true no Webhook!
	ConversionValue float64 `gorm:"default:0.0" json:"conversion_value"`        // Valor da venda (ex: 197.00)
	TransactionID   string  `json:"transaction_id"`                             // ID da transação da Hotmart/Kiwify
	PaymentStatus   string  `gorm:"default:'none';index" json:"payment_status"` // none, paid, pending, refunded, canceled
	Gateway         string  `json:"gateway"`
	ReachedCheckout bool    `gorm:"default:false;index" json:"reached_checkout"` // kiwify, eduzz, hotmart...

	CreatedAt time.Time `gorm:"index" json:"created_at"`
}
