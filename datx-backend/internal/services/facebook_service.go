package services

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// HashSHA256 padroniza os dados para o Facebook (E-mail, Telefone, etc)
func HashSHA256(data string) string {
	if data == "" {
		return ""
	}
	h := sha256.New()
	h.Write([]byte(strings.ToLower(strings.TrimSpace(data))))
	return hex.EncodeToString(h.Sum(nil))
}

type FBEvent struct {
	EventName      string `json:"event_name"`
	EventTime      int64  `json:"event_time"`
	ActionSource   string `json:"action_source"`
	EventSourceURL string `json:"event_source_url"`
	UserData       struct {
		ClientIPAddress string `json:"client_ip_address"`
		ClientUserAgent string `json:"client_user_agent"`
		Fbp             string `json:"fbp,omitempty"`
		Fbc             string `json:"fbc,omitempty"`
		ExternalID      string `json:"external_id,omitempty"`
		Email           string `json:"em,omitempty"` // Hashed
		Phone           string `json:"ph,omitempty"` // Hashed
	} `json:"user_data"`
	CustomData struct {
		Currency string  `json:"currency,omitempty"`
		Value    float64 `json:"value,omitempty"`
	} `json:"custom_data,omitempty"`
}

func PushToFacebook(pixelID, token, eventName string, click map[string]interface{}, value float64, email string) {
	url := fmt.Sprintf("https://graph.facebook.com/v18.0/%s/events?access_token=%s", pixelID, token)

	event := FBEvent{
		EventName:      eventName,
		EventTime:      time.Now().Unix(),
		ActionSource:   "website",
		EventSourceURL: fmt.Sprintf("%v", click["page_url"]),
	}

	// Dados de rastreio (O MATCH PERFEITO)
	event.UserData.ClientIPAddress = fmt.Sprintf("%v", click["ip"])
	event.UserData.ClientUserAgent = fmt.Sprintf("%v", click["ua"])
	event.UserData.Fbp = fmt.Sprintf("%v", click["fbp"])
	event.UserData.Fbc = fmt.Sprintf("%v", click["fbc"])
	event.UserData.ExternalID = fmt.Sprintf("%v", click["external_id"])

	// Se tivermos o email do Webhook, fazemos o hash pra subir o Match Score
	if email != "" {
		event.UserData.Email = HashSHA256(email)
	}

	if value > 0 {
		event.CustomData.Currency = "BRL"
		event.CustomData.Value = value
	}

	payload := map[string]interface{}{"data": []FBEvent{event}}
	jsonData, _ := json.Marshal(payload)

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println("❌ Erro CAPI:", err)
		return
	}
	defer resp.Body.Close()
	fmt.Printf("🚀 [CAPI] Evento %s enviado! Pixel: %s | Status: %s\n", eventName, pixelID, resp.Status)
}
