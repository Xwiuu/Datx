package database

import (
	"fmt"
	"log"

	"github.com/xwiuu/datx-backend/internal/models" // Ajuste o caminho se necessário
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func ConnectDB() {
	dsn := "host=localhost user=admin password=datx_secret_password dbname=datx_db port=5432 sslmode=disable"

	// 1. Conecta (usando uma variável temporária ou direto na global)
	database, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

	if err != nil {
		log.Fatal("Failed to connect to database: ", err)
	}

	// 2. ATRIBUIÇÃO CRUCIAL: Agora a global DB não é mais nula
	DB = database

	fmt.Println("🚀 Database Connected Successfully!")

	// 3. AGORA SIM MIGRAR: Como DB já tem valor, não vai dar panic
	err = DB.AutoMigrate(
		&models.User{},
		&models.Product{},
		&models.Link{},
		&models.ClickLog{},
		&models.LinkStats{},
	)

	if err != nil {
		log.Fatal("Failed to migrate database: ", err)
	}
}
