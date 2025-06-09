package presentation

import (
	"api/app/domain/entities"
	apiv1 "api/gen/api/v1"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sort"
	"strings"
	"time"

	"connectrpc.com/connect"
)

type BalanceUsecase interface {
	UpdateBalance(ctx context.Context, userName string, endDate time.Time) (string, int64, []string, error)
}
type TinkoffNotificationHandler struct {
	balanceUsecase BalanceUsecase

	secretKey        string
	terminalKey      string
	telegramBotToken string
}

func NewTinkoffNotificationHandler(balanceUsecase BalanceUsecase, secretKey string, terminalKey string, telegramBotToken string) *TinkoffNotificationHandler {
	return &TinkoffNotificationHandler{
		balanceUsecase: balanceUsecase,

		secretKey:        secretKey,
		terminalKey:      terminalKey,
		telegramBotToken: telegramBotToken,
	}
}

// Method for processing payment request
func (h *TinkoffNotificationHandler) ProcessPaymentRequestHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	// Read request body
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Не удалось прочитать тело запроса", http.StatusBadRequest)
		return
	}

	// Unmarshal request body
	var requestData entities.Order

	err = json.Unmarshal(bodyBytes, &requestData)
	if err != nil {
		http.Error(w, "Некорректный JSON", http.StatusBadRequest)
		return
	}

	err = requestData.Validate()
	if err != nil {
		http.Error(w, fmt.Sprintf("Некорректные данные: %v", err), http.StatusBadRequest)
		return
	}

	// Generate payment link
	paymentURL, err := h.GeneratePaymentLink(requestData)
	if err != nil {
		http.Error(w, fmt.Sprintf("Не удалось сгенерировать платежную ссылку: %v", err), http.StatusInternalServerError)
		return
	}

	// return payment link
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(paymentURL))
}

// Method for generating payment link

func (h *TinkoffNotificationHandler) GeneratePaymentLink(order entities.Order) (string, error) {

	data := order.ToPaymentData(h.terminalKey, h.secretKey)

	// Копируем данные без поля Receipt для расчета Token
	dataForToken := make(map[string]interface{})
	for k, v := range data {
		if k != "Receipt" {
			dataForToken[k] = v
		}
	}

	// Сортируем ключи
	keys := make([]string, 0, len(dataForToken))
	for k := range dataForToken {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Конкатенируем значения
	var values []string
	for _, k := range keys {
		values = append(values, fmt.Sprintf("%v", dataForToken[k]))
	}
	concatenatedString := strings.Join(values, "")

	// Вычисляем хеш
	hash := sha256.Sum256([]byte(concatenatedString))
	hashedString := hex.EncodeToString(hash[:])

	// Добавляем Token в исходные данные
	data["Token"] = hashedString

	// Удаляем Password из данных перед отправкой
	delete(data, "Password")

	// Преобразуем данные в JSON
	requestBody, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("ошибка при преобразовании данных запроса в JSON: %v", err)
	}

	log.Printf("data: %s", data)
	// Отправляем запрос в Тинькофф
	url := "https://securepay.tinkoff.ru/v2/Init"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		return "", fmt.Errorf("ошибка при создании запроса: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("ошибка при отправке запроса: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API Тинькофф вернул статус %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// Читаем ответ
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("ошибка при чтении ответа: %v", err)
	}

	// Разбираем ответ
	var responseMap map[string]interface{}
	err = json.Unmarshal(bodyBytes, &responseMap)
	if err != nil {
		return "", fmt.Errorf("ошибка при разборе JSON ответа: %v", err)
	}

	// Проверяем успех и получаем ссылку на оплату
	if success, ok := responseMap["Success"].(bool); ok && success {
		if paymentURL, ok := responseMap["PaymentURL"].(string); ok {
			return paymentURL, nil
		} else {
			return "", fmt.Errorf("PaymentURL не найден в ответе")
		}
	} else {
		return "", fmt.Errorf("инициация платежа не удалась: %v", responseMap)
	}
}

func (h *TinkoffNotificationHandler) verifySignature(params map[string]string) bool {

	// Extract token from params
	receivedToken := params["Token"]
	if receivedToken == "" {
		log.Printf("Отсутствует Token в параметрах")
		return false
	}

	// 1. Собираем массив всех обязательных передаваемых параметров для конкретного метода в виде пар Ключ-Значение — кроме параметра Token
	signatureParams := make(map[string]string)
	for key, value := range params {
		if key != "Token" {
			signatureParams[key] = value
		}
	}

	// 2. Добавляем в массив пару Password
	signatureParams["Password"] = h.secretKey

	// 3. Отсортируем массив по ключам по алфавиту
	keys := make([]string, 0, len(signatureParams))
	for key := range signatureParams {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	// 4. Конкатенируем значения всех пар в порядке сортировки ключей
	var concatenatedValues strings.Builder
	for _, key := range keys {
		concatenatedValues.WriteString(signatureParams[key])
	}
	concatenatedString := concatenatedValues.String()

	// 5. Вычисляем SHA-256 хеш от полученной строки
	hash := sha256.Sum256([]byte(concatenatedString))
	calculatedToken := hex.EncodeToString(hash[:])

	// 6. Сравниваем вычисленный хеш с полученным значением Token (без учета регистра)
	return strings.EqualFold(calculatedToken, receivedToken)
}

func (h *TinkoffNotificationHandler) TinkoffNotificationHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		log.Printf("Метод не поддерживается")
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	// Получаем и логируем Content-Type
	contentType := r.Header.Get("Content-Type")

	var stringParams map[string]string

	if strings.Contains(contentType, "application/json") {
		// Обработка JSON
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			log.Printf("Ошибка чтения тела запроса: %v", err)
			http.Error(w, "Ошибка чтения тела запроса", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		log.Printf("Тело запроса: %s", string(bodyBytes))

		//  json.Decoder с UseNumber() for  unmarshal a number into an interface{} as a [Number] instead of as a float64.
		decoder := json.NewDecoder(bytes.NewReader(bodyBytes))
		decoder.UseNumber()

		var params map[string]interface{}
		err = decoder.Decode(&params)
		if err != nil {
			log.Printf("Ошибка разбора JSON: %v", err)
			http.Error(w, "Некорректный JSON", http.StatusBadRequest)
			return
		}

		// Transform interface{} to string taking into account json.Number
		stringParams = make(map[string]string)
		for key, value := range params {
			switch v := value.(type) {
			case json.Number:
				stringParams[key] = v.String()
			case string:
				stringParams[key] = v
			case bool:
				stringParams[key] = fmt.Sprintf("%t", v)
			default:
				stringParams[key] = fmt.Sprintf("%v", v)
			}
		}

	} else if strings.Contains(contentType, "application/x-www-form-urlencoded") {
		// Form-data
		err := r.ParseForm()
		if err != nil {
			log.Printf("Ошибка разбора формы: %v", err)
			http.Error(w, "Некорректный запрос", http.StatusBadRequest)
			return
		}

		stringParams = make(map[string]string)
		for key, values := range r.Form {
			if len(values) > 0 {
				stringParams[key] = values[0]
			}
		}

	} else {
		log.Printf("Некорректный Content-Type: %s", contentType)
		http.Error(w, "Некорректный Content-Type", http.StatusBadRequest)
		return
	}

	// Check signature
	validSignature := h.verifySignature(stringParams)
	if !validSignature {
		log.Printf("Неверная подпись")
		http.Error(w, "Неверная подпись", http.StatusBadRequest)
		return
	}

	// Check status and other parameters
	status := stringParams["Status"]
	paymentID := stringParams["PaymentId"]

	email, endDate, err := simpleDecrypt(stringParams["OrderId"])
	if err != nil {
		log.Printf("Ошибка расшифровки OrderId: %v", err)
	}

	endDateParsed, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		log.Printf("Ошибка парсинга даты: %v", err)
	}
	var message string
	switch status {
	case "AUTHORIZED":
		message = fmt.Sprintf("Платеж авторизован. PaymentId: %s, OrderId: %s endDate: %s", paymentID, email, endDate)
		h.balanceUsecase.UpdateBalance(r.Context(), email, endDateParsed)
	case "CONFIRMED":
		// if payment is confirmed, update subscription
		message = fmt.Sprintf("Платеж подтвержден. PaymentId: %s, OrderId: %s", paymentID, email)
	case "REJECTED":
		message = fmt.Sprintf("Платеж отклонен. PaymentId: %s, OrderId: %s", paymentID, email)
	case "REFUNDED":
		message = fmt.Sprintf("Произведен возврат. PaymentId: %s, OrderId: %s", paymentID, email)
	default:
		message = fmt.Sprintf("Статус платежа: %s. PaymentId: %s, OrderId: %s", status, paymentID, email)
	}

	log.Print(message)

	// Send response to Tinkoff
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func simpleDecrypt(order string) (string, string, error) {

	parts := strings.Split(order, ";")
	if len(parts) != 3 {
		return "", "", fmt.Errorf("неверный формат данных (ожидалось 'email;date;orderNumber')")
	}

	email := parts[0]
	date := parts[1]
	return email, date, nil
}

// CreatePaymentRequest implements PaymentService.CreatePaymentRequest
func (h *TinkoffNotificationHandler) CreatePaymentRequest(ctx context.Context, req *connect.Request[apiv1.PaymentRequestInput]) (*connect.Response[apiv1.PaymentResponse], error) {
	// Validate required fields
	if req.Msg.Amount <= 0 {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("некорректная сумма"))
	}
	if req.Msg.Email == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("не указан email"))
	}
	if req.Msg.OrderNumber <= 0 {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("некорректный номер заказа"))
	}

	// Convert protobuf to entities.Order for compatibility with existing code
	order := entities.Order{
		Amount:      int(req.Msg.Amount),
		Email:       req.Msg.Email,
		OrderNumber: int(req.Msg.OrderNumber),
		Description: req.Msg.Description,
		EndDate:     req.Msg.EndDate,
	}

	// Convert receipt if provided
	if req.Msg.Receipt != nil {
		order.Receipt = entities.Receipt{
			Email:    req.Msg.Receipt.Email,
			Taxation: req.Msg.Receipt.Taxation,
			Items:    make([]entities.ReceiptItem, len(req.Msg.Receipt.Items)),
		}
		for i, item := range req.Msg.Receipt.Items {
			order.Receipt.Items[i] = entities.ReceiptItem{
				Name:          item.Name,
				Price:         int(item.Price),
				Quantity:      item.Quantity,
				Amount:        int(item.Amount),
				Tax:           item.Tax,
				PaymentMethod: item.PaymentMethod,
				PaymentObject: item.PaymentObject,
			}
		}
	}

	// Validate order
	if err := order.Validate(); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("некорректные данные: %w", err))
	}

	// Generate payment link using existing logic
	paymentURL, err := h.GeneratePaymentLink(order)
	if err != nil {
		log.Printf("Ошибка генерации платежной ссылки: %v", err)
		return &connect.Response[apiv1.PaymentResponse]{
			Msg: &apiv1.PaymentResponse{
				Success:      false,
				ErrorMessage: err.Error(),
			},
		}, nil
	}

	// Return success response
	return &connect.Response[apiv1.PaymentResponse]{
		Msg: &apiv1.PaymentResponse{
			Success:    true,
			PaymentUrl: paymentURL,
		},
	}, nil
}

// ProcessTinkoffNotification implements PaymentService.ProcessTinkoffNotification
func (h *TinkoffNotificationHandler) ProcessTinkoffNotification(ctx context.Context, req *connect.Request[apiv1.TinkoffNotificationRequest]) (*connect.Response[apiv1.TinkoffNotificationResponse], error) {
	// Convert protobuf request to map[string]string for signature verification
	params := map[string]string{
		"TerminalKey": req.Msg.TerminalKey,
		"Amount":      fmt.Sprintf("%d", req.Msg.Amount),
		"OrderId":     fmt.Sprintf("%d", req.Msg.OrderId),
		"Success":     fmt.Sprintf("%t", req.Msg.Success),
		"Status":      req.Msg.Status,
		"PaymentId":   fmt.Sprintf("%d", req.Msg.PaymentId),
		"ErrorCode":   req.Msg.ErrorCode,
		"Message":     req.Msg.Message,
		"Details":     req.Msg.Details,
		"Token":       req.Msg.Token,
	}

	// Verify signature
	if !h.verifySignature(params) {
		log.Printf("Неверная подпись")
		return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("неверная подпись"))
	}

	// Process notification
	status := req.Msg.Status
	paymentID := fmt.Sprintf("%d", req.Msg.PaymentId)
	orderIdStr := fmt.Sprintf("%d", req.Msg.OrderId)

	email, endDate, err := simpleDecrypt(orderIdStr)
	if err != nil {
		log.Printf("Ошибка расшифровки OrderId: %v", err)
	}

	endDateParsed, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		log.Printf("Ошибка парсинга даты: %v", err)
	}

	var message string
	switch status {
	case "AUTHORIZED":
		message = fmt.Sprintf("Платеж авторизован. PaymentId: %s, OrderId: %s endDate: %s", paymentID, email, endDate)
		h.balanceUsecase.UpdateBalance(ctx, email, endDateParsed)
	case "CONFIRMED":
		message = fmt.Sprintf("Платеж подтвержден. PaymentId: %s, OrderId: %s", paymentID, email)
	case "REJECTED":
		message = fmt.Sprintf("Платеж отклонен. PaymentId: %s, OrderId: %s", paymentID, email)
	case "REFUNDED":
		message = fmt.Sprintf("Произведен возврат. PaymentId: %s, OrderId: %s", paymentID, email)
	default:
		message = fmt.Sprintf("Статус платежа: %s. PaymentId: %s, OrderId: %s", status, paymentID, email)
	}

	log.Print(message)

	return &connect.Response[apiv1.TinkoffNotificationResponse]{
		Msg: &apiv1.TinkoffNotificationResponse{
			Status: "OK",
		},
	}, nil
}
