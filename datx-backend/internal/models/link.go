package models

import (
	"time"

	"github.com/google/uuid"
)

type Link struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;" json:"id"`
	UserID      uuid.UUID `gorm:"type:uuid;not null" json:"user_id"`
	Name        string    `json:"name"`
	Slug        string    `gorm:"uniqueIndex;not null" json:"slug"`
	RedirectURL string    `gorm:"not null" json:"redirect_url"`
	PageURL     string    `gorm:"not null" json:"page_url"`

	// --- FILTROS DE DISPOSITIVO ---
	AllowMobile bool `gorm:"default:true" json:"allow_mobile"`
	AllowPC     bool `gorm:"default:true" json:"allow_pc"`

	// --- MÓDULOS DE SEGURANÇA (O ARSENAL) ---
	BlockBots           bool `gorm:"default:true" json:"block_bots"`
	HardwareFingerprint bool `gorm:"default:true" json:"hardware_fingerprint"`
	BehaviorCheck       bool `gorm:"default:true" json:"behavior_check"`
	AntiInspect         bool `gorm:"default:true" json:"anti_inspect"`
	AntiClone           bool `gorm:"default:true" json:"anti_clone"`
	IframeBuster        bool `gorm:"default:true" json:"iframe_buster"` // ADICIONE ESTA LINHA 🛡️

	// --- CONFIGURAÇÃO DE CHECKOUT ---
	HasCheckout bool   `gorm:"default:false" json:"has_checkout"`
	CheckoutURL string `json:"checkout_url"`

	// --- ANALYTICS (BIG NUMBERS) ---
	TotalVisits  int `gorm:"default:0" json:"total_visits"`
	HumanVisits  int `gorm:"default:0" json:"human_visits"`
	BotBlocked   int `gorm:"default:0" json:"bot_blocked"`
	MobileVisits int `gorm:"default:0" json:"mobile_visits"`
	PCVisits     int `gorm:"default:0" json:"pc_visits"`

	FbPixelID     string `json:"fb_pixel_id"`
	FbAccessToken string `json:"fb_access_token"`

	Status    string    `gorm:"default:'active'" json:"status"`
	CreatedAt time.Time `json:"created_at"`
}
