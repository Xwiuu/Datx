package models

import (
	"time"

	"github.com/google/uuid"
)

type ShieldLog struct {
	ID     uuid.UUID `gorm:"type:uuid;primary_key;" json:"id"`
	LinkID uuid.UUID `gorm:"type:uuid;index" json:"link_id"` // ID da Campanha
	UserID uuid.UUID `gorm:"type:uuid;index" json:"user_id"` // Dono da campanha

	// Dados do Visitante
	IP        string `json:"ip"`
	UserAgent string `json:"user_agent"`
	Country   string `json:"country"` // Podemos integrar GeoIP depois
	IsBot     bool   `json:"is_bot"`

	// O Motivo do Bloqueio (O que o Sentinel vai ler)
	// Ex: "GPU_FAIL", "USER_AGENT_BOT", "INSPECT_TRIGGER", "IFRAME_DETECT"
	Reason string `json:"reason"`

	// Destino final: "MONEY" (Passou) ou "SAFE" (Barrado)
	FinalDest string `json:"final_dest"`

	CreatedAt time.Time `json:"created_at"`
}
