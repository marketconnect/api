package entities

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
)

type Order struct {
	Amount      int    `json:"amount"`
	Email       string `json:"email"`
	OrderNumber int    `json:"orderNumber"`
	Description string `json:"description"`
	// SubscriptionID   int     `json:"subscriptionId"`
	// SubscriptionType string  `json:"subscriptionType"`
	// StartDate        string  `json:"startDate"`
	EndDate string  `json:"endDate"`
	Receipt Receipt `json:"receipt"`
}

func (o *Order) Validate() error {
	if o.Amount <= 0 {
		return fmt.Errorf("invalid amount")
	}
	if o.Email == "" {
		return fmt.Errorf("invalid chatId")
	}
	if o.OrderNumber == 0 {
		return fmt.Errorf("invalid orderNumber")
	}

	if o.EndDate == "" {
		return fmt.Errorf("invalid endDate")
	}
	if err := o.Receipt.Validate(); err != nil {
		return err
	}

	return nil
}

func (o *Order) GenerateOrderID() string {

	return fmt.Sprintf("%s;%s;%d", o.Email, o.EndDate, o.OrderNumber)
}

func (o *Order) ToPaymentData(terminalKey, secretKey string) map[string]interface{} {
	data := map[string]interface{}{
		"Amount":      o.Amount,
		"OrderId":     o.GenerateOrderID(),
		"Description": o.Description,
		"TerminalKey": terminalKey,
		"Password":    secretKey,
		"Receipt":     o.Receipt, // Добавляем Receipt в данные запроса
	}
	return data
}

func GenerateSignature(data map[string]interface{}) string {
	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var values []string
	for _, k := range keys {
		values = append(values, fmt.Sprintf("%v", data[k]))
	}
	concatenatedString := strings.Join(values, "")

	hash := sha256.Sum256([]byte(concatenatedString))
	return hex.EncodeToString(hash[:])
}

type ReceiptItem struct {
	Name          string  `json:"Name"`
	Price         int     `json:"Price"`
	Quantity      float64 `json:"Quantity"`
	Amount        int     `json:"Amount"`
	Tax           string  `json:"Tax"`
	PaymentMethod string  `json:"PaymentMethod"`
	PaymentObject string  `json:"PaymentObject"`
}

type Receipt struct {
	Email    string        `json:"Email,omitempty"`
	Phone    string        `json:"Phone,omitempty"`
	Taxation string        `json:"Taxation"`
	Items    []ReceiptItem `json:"Items"`
}

func (r *Receipt) Validate() error {
	if r.Taxation == "" {
		return fmt.Errorf("Receipt Taxation is required")
	}
	if len(r.Items) == 0 {
		return fmt.Errorf("Receipt must have at least one item")
	}
	for _, item := range r.Items {
		if err := item.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func (item *ReceiptItem) Validate() error {
	if item.Name == "" {
		return fmt.Errorf("item Name is required")
	}
	if item.Price <= 0 {
		return fmt.Errorf("item Price must be greater than zero")
	}
	if item.Quantity <= 0 {
		return fmt.Errorf("item Quantity must be greater than zero")
	}
	if item.Amount <= 0 {
		return fmt.Errorf("item Amount must be greater than zero")
	}
	if item.Tax == "" {
		return fmt.Errorf("item Tax is required")
	}
	return nil
}
