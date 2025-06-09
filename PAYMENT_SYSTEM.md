# Payment System Documentation

## Overview

This API now includes a complete payment system integration with Tinkoff Bank and balance protection middleware.

## Architecture Components

### 1. Payment Endpoints

#### POST `/payment/request`
Creates a payment request and returns a Tinkoff payment URL.

**Request Body:**
```json
{
  "amount": 100000,
  "email": "user@example.com", 
  "orderNumber": 12345,
  "description": "API Credits Purchase",
  "endDate": "2025-12-31",
  "receipt": {
    "Email": "user@example.com",
    "Taxation": "usn_income",
    "Items": [
      {
        "Name": "API Credits",
        "Price": 100000,
        "Quantity": 1.0,
        "Amount": 100000,
        "Tax": "none",
        "PaymentMethod": "full_payment",
        "PaymentObject": "service"
      }
    ]
  }
}
```

**Response:** Payment URL from Tinkoff

#### POST `/payment/notification`
Webhook endpoint for Tinkoff payment notifications. Automatically updates user balance when payment is authorized.

### 2. Balance System

#### GET `/balance`
Returns user balance (requires Authorization header).

**Headers:**
```
Authorization: Bearer your-api-key
```

**Response:**
```json
{
  "balance": 1000
}
```

### 3. Balance Protection Middleware

Automatically protects all API endpoints (except balance and payment endpoints) by requiring minimum balance of 10 credits.

**Protected Endpoints:**
- All `/api.v1.CreateProductCardService/*` endpoints
- Any other future endpoints

**Excluded Endpoints:**
- `/balance*` - Balance checking endpoints
- `/payment/*` - Payment endpoints  
- `/metrics` - Prometheus metrics

## Environment Variables

Add these to your environment:

```bash
# Tinkoff Configuration
TINKOFF_SECRET_KEY=your-secret-key
TINKOFF_TERMINAL_KEY=your-terminal-key
TELEGRAM_BOT_TOKEN=your-bot-token  # Optional
```

## Database Schema

The system uses existing tables:

```sql
-- User balances
CREATE TABLE IF NOT EXISTS api_key_balances (
    api_key TEXT PRIMARY KEY,
    balance INT NOT NULL
);

-- Token costs for billing
CREATE TABLE IF NOT EXISTS token_costs (
    token_type TEXT PRIMARY KEY,
    cost INT NOT NULL
);
```

## Usage Flow

1. **User makes payment request** → GET payment URL
2. **User pays via Tinkoff** → Webhook updates balance 
3. **User makes API calls** → Balance checked automatically
4. **Insufficient balance** → Request blocked with 402 Payment Required

## Testing

Use the provided test script:

```bash
./test_payment.sh
```

This tests:
- Payment request creation
- Balance checking
- Balance protection middleware

## Error Codes

- `401 Unauthorized` - Missing/invalid API key
- `402 Payment Required` - Insufficient balance  
- `400 Bad Request` - Invalid payment data
- `500 Internal Server Error` - System error

## Payment Flow

1. **Payment Request**: Client calls `/payment/request` with order details
2. **Tinkoff Processing**: System creates payment in Tinkoff, returns URL
3. **User Payment**: User completes payment via Tinkoff URL
4. **Webhook Notification**: Tinkoff sends notification to `/payment/notification`
5. **Balance Update**: System automatically adds credits to user balance
6. **API Access**: User can now make API calls (if balance ≥ 10)

## Security Features

- Signature verification for Tinkoff webhooks
- API key authentication for all protected endpoints
- Balance validation before processing requests
- Secure payment data handling (receipts, amounts) 