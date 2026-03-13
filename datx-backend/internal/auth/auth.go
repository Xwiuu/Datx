package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var jwtSecret = []byte("SUA_CHAVE_SUPER_SECRETA_DATX_2026") // Isso deve ir para o .env depois

// HashPassword transforma a senha em uma hash segura
func HashPassword(password string) (string, error) {
	// Custo 10: Rápido, seguro e não trava o servidor!
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	return string(bytes), err
}

// CheckPasswordHash compara a senha digitada com a do banco
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// GenerateToken cria o JWT que o Front-end vai guardar
func GenerateToken(email string) (string, error) {
	claims := jwt.MapClaims{
		"email": email,
		"exp":   time.Now().Add(time.Hour * 72).Unix(), // Token dura 3 dias
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}
