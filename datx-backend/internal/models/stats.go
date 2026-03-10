package models

type LinkStats struct {
	ID         uint   `gorm:"primaryKey"`
	Slug       string `gorm:"index" json:"slug"`
	Hour       string `json:"hour"`
	Date       string `gorm:"index" json:"date"`
	HumanCount int    `json:"humanos"`
	BotCount   int    `json:"bots"`
}
