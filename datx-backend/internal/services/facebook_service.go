package services

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io" // Importação nova para ler erro
	"net/http"
	"strings"
	"time"
)

// Helper para evitar o envio de "<nil>" ou strings vazias
func getString(m map[string]interface{}, key string) string {
	val, ok := m[key]
	if !ok || val == nil {
		return ""
	}
	return fmt.Sprintf("%v", val)
}

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
		Email           string `json:"em,omitempty"`
	} `json:"user_data"`
	CustomData struct {
		Currency string  `json:"currency,omitempty"`
		Value    float64 `json:"value,omitempty"`
	} `json:"custom_data,omitempty"`
}

func PushToFacebook(pixelID, token, eventName string, click map[string]interface{}, value float64, email string) {
	// API v18.0 está ótima, mas a v19.0+ já é padrão. Pode manter a 18 se quiser.
	url := fmt.Sprintf("https://graph.facebook.com/v19.0/%s/events?access_token=%s", pixelID, token)

	event := FBEvent{
		EventName:      eventName,
		EventTime:      time.Now().Unix(),
		ActionSource:   "website",
		EventSourceURL: getString(click, "page_url"),
	}

	// 🛡️ Blindagem contra valores <nil>
	event.UserData.ClientIPAddress = getString(click, "ip")
	event.UserData.ClientUserAgent = getString(click, "ua")
	event.UserData.Fbp = getString(click, "fbp")
	event.UserData.Fbc = getString(click, "fbc")
	event.UserData.ExternalID = getString(click, "external_id")

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
		fmt.Println("❌ Erro Conexão CAPI:", err)
		return
	}
	defer resp.Body.Close()

	// 🔍 DEBUG: Se falhar, vamos ler o porquê (O Facebook manda o erro no body)
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("⚠️ [CAPI] Falha no Pixel %s: %s\n", pixelID, string(body))
	} else {
		fmt.Printf("🚀 [CAPI] Evento %s enviado com Match Score alto! | Status: %s\n", eventName, resp.Status)
	}
}
