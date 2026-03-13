package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Product struct {
	ID             uuid.UUID `gorm:"type:uuid;primary_key;" json:"id"`
	UserID         uuid.UUID `gorm:"type:uuid;index;not null" json:"user_id"`
	Name           string    `gorm:"not null" json:"name"`
	GlobalSafePage string    `json:"global_safe_page"` // URL padrão para bots deste produto

	// Centralização do CAPI no Produto
	FbPixelID     string `json:"fb_pixel_id"`
	FbAccessToken string `json:"fb_access_token"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (p *Product) BeforeCreate(tx *gorm.DB) (err error) {
	p.ID = uuid.New()
	return
}
