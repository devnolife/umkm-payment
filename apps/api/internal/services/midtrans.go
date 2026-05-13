package services

import (
	"bytes"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/devnolife/umkm-api/internal/config"
)

type MidtransSnapResponse struct {
	Token       string `json:"token"`
	RedirectURL string `json:"redirect_url"`
}

type SnapItem struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Price    int    `json:"price"`
	Quantity int    `json:"quantity"`
}

type SnapCustomer struct {
	FirstName string `json:"first_name"`
	Email     string `json:"email,omitempty"`
	Phone     string `json:"phone,omitempty"`
}

type SnapPayload struct {
	TransactionDetails map[string]any `json:"transaction_details"`
	ItemDetails        []SnapItem     `json:"item_details"`
	CustomerDetails    SnapCustomer   `json:"customer_details"`
}

// CreateSnapTransaction calls Midtrans Snap API.
func CreateSnapTransaction(orderID string, grossAmount int, items []SnapItem, customer SnapCustomer) (*MidtransSnapResponse, error) {
	cfg := config.Get()
	if cfg.MidtransServerKey == "" {
		return nil, errors.New("MIDTRANS_SERVER_KEY not configured")
	}

	endpoint := "https://app.sandbox.midtrans.com/snap/v1/transactions"
	if cfg.MidtransIsProduction {
		endpoint = "https://app.midtrans.com/snap/v1/transactions"
	}

	payload := SnapPayload{
		TransactionDetails: map[string]any{
			"order_id":     orderID,
			"gross_amount": grossAmount,
		},
		ItemDetails:     items,
		CustomerDetails: customer,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	auth := base64.StdEncoding.EncodeToString([]byte(cfg.MidtransServerKey + ":"))
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Basic "+auth)

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("midtrans %d: %s", resp.StatusCode, string(raw))
	}

	var out MidtransSnapResponse
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// VerifyNotificationSignature checks the SHA512 signature_key from Midtrans webhook.
func VerifyNotificationSignature(orderID, statusCode, grossAmount, signatureKey string) bool {
	cfg := config.Get()
	if cfg.MidtransServerKey == "" {
		return false
	}
	raw := orderID + statusCode + grossAmount + cfg.MidtransServerKey
	h := sha512.Sum512([]byte(raw))
	return hex.EncodeToString(h[:]) == signatureKey
}
