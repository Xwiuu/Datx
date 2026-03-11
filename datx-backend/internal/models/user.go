package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID            uuid.UUID `gorm:"type:uuid;primary_key;" json:"id"`
	Email         string    `gorm:"uniqueIndex;not null" json:"email"`
	Password      string    `gorm:"not null" json:"-"` // O json:"-" esconde a senha nas respostas da API
	Name          string    `json:"name"`
	Plan          string    `gorm:"default:'free'" json:"plan"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	FbPixelID     string    `json:"fb_pixel_id"`
	FbAccessToken string    `json:"fb_access_token"`
}

// Hook do GORM para gerar o UUID antes de criar no banco
func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	u.ID = uuid.New()
	return
}
