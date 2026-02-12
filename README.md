# Maib E-commerce Go Package

![Go Version](https://img.shields.io/github/go-mod/go-version/acolev/maib-ecomm)
[![GoDoc](https://godoc.org/github.com/acolev/maib-ecomm?status.svg)](https://godoc.org/github.com/acolev/maib-ecomm)

A robust, idiomatic Go client for the Maib E-commerce API. This package provides a simple interface to process payments, manage recurring billing, and handle callbacks with automatic signature verification.

[Читать на русском](README.ru.md)

---

## Features

*   **Complete Implementation**: Supports all API methods (Pay, Refund, Register Card, Execute Recurring, Delete Card).
*   **Automatic Signature Verification**: Securely validates callback notifications using the correct algorithm (reverse-engineered from official PHP SDK).
*   **Type Safety**: Strongly typed request and response structures.
*   **Context Support**: Full support for `context.Context` for timeouts and cancellation.
*   **Token Management**: Handles access token generation and caching automatically.

## Installation

```bash
go get github.com/acolev/maib-ecomm
```

## Usage

### 1. Initialize Client

You need your Project ID, Project Secret, and Signature Key (for callbacks) from the Maib merchant portal.

```go
import "github.com/acolev/maib-ecomm"

client := maib.NewClient(
    maib.WithProjectID("YOUR_PROJECT_ID"),
    maib.WithProjectSecret("YOUR_PROJECT_SECRET"),
    maib.WithSignatureKey("YOUR_SIGNATURE_KEY"), // Required for callback validation
)
```

### 2. Create a Payment (Direct)

```go
req := &maib.PayRequest{
    Amount:      100.00,
    Currency:    maib.CurrencyMDL,
    ClientIP:    "127.0.0.1",
    Language:    maib.LanguageRO,
    Description: "Order #123",
    CallbackUrl: "https://your-site.com/callback",
    OkUrl:       "https://your-site.com/success",
    FailUrl:     "https://your-site.com/fail",
}

resp, err := client.Pay(ctx, req)
if err != nil {
    log.Fatal(err)
}

// Redirect user to payment page
fmt.Printf("Redirect to: %s\n", resp.PayURL)
```

### 3. Handle Callback (Webhook)

This package automatically verifies the signature of the incoming callback request.

```go
func handleCallback(w http.ResponseWriter, r *http.Request) {
    // Parse verifies signature and decodes JSON
    data, err := client.ParseCallback(r)
    if err != nil {
        log.Printf("Invalid callback: %v", err)
        http.Error(w, "Invalid signature", http.StatusBadRequest)
        return
    }

    if data.Status == "OK" {
        fmt.Printf("Payment %s (Order %s) successful!\n", data.PayID, data.OrderID)
    } else {
        fmt.Printf("Payment failed: %s\n", data.StatusMessage)
    }

    w.WriteHeader(http.StatusOK)
}
```

### 4. Recurring Payments (Save Card)

**Step 1: Register the card**
This initiates a transaction (usually 0 or small amount) to save the card.

```go
req := &maib.RegisterCardRequest{
    BillerExpiry: "1225", // December 2025
    ClientIP:     "127.0.0.1",
    Currency:     maib.CurrencyMDL,
    Language:     maib.LanguageRO,
    Email:        "user@example.com", // Required for recurring
    CallbackUrl:  "https://your-site.com/callback",
    // ...
}

resp, err := client.RegisterCard(ctx, req)
// Redirect user to resp.PayURL
```

**Step 2: Obtain BillerID from Callback**
In the callback for the registration, you will receive a `billerId`. Save this!

```go
// Inside callback handler
if data.BillerID != "" {
    saveToDB(userID, data.BillerID)
}
```

**Step 3: Execute Recurring Payment**
Charge the saved card without user interaction.

```go
resp, err := client.ExecuteRecurring(ctx, &maib.ExecuteRecurringRequest{
    BillerID: "SAVED_BILLER_ID",
    Amount:   50.00,
    Currency: maib.CurrencyMDL,
    Description: "Monthly Subscription",
})

if resp.Status == "OK" {
    fmt.Println("Charged successfully!")
}
```

## References

*   [Official Documentation](https://docs.maibmerchants.md/e-commerce/maib-e-commerce-api)
*   [Reference Repository](https://github.com/acolev/maib-ecomm)
