# Maib E-commerce Client на Golang

![Go Version](https://img.shields.io/github/go-mod/go-version/acolev/maib-ecomm)
[![GoDoc](https://godoc.org/github.com/acolev/maib-ecomm?status.svg)](https://godoc.org/github.com/acolev/maib-ecomm)

Надежный и идиоматичный клиент на Go для API электронной коммерции Maib. Этот пакет предоставляет простой интерфейс для обработки платежей, управления регулярными списаниями и обработки колбэков с автоматической проверкой подписи.

[Read this in English](README.md)

---

## Возможности

*   **Полная реализация**: Поддержка всех методов API (Оплата, Возврат, Регистрация карты, Рекуррентный платеж, Удаление карты).
*   **Автоматическая проверка подписи**: Безопасная валидация уведомлений (callback) с использованием корректного алгоритма (восстановлен из официального PHP SDK).
*   **Строгая типизация**: Структуры запросов и ответов строго типизированы.
*   **Поддержка Context**: Полная поддержка `context.Context` для таймаутов и отмены запросов.
*   **Управление токенами**: Автоматическое получение и кэширование Access Token.

## Установка

```bash
go get github.com/acolev/maib-ecomm
```

## Использование

### 1. Инициализация клиента

Вам понадобятся Project ID, Project Secret и Signature Key (для колбэков) из личного кабинета мерчанта Maib.

```go
import "github.com/acolev/maib-ecomm"

client := maib.NewClient(
    maib.WithProjectID("ВАШ_PROJECT_ID"),
    maib.WithProjectSecret("ВАШ_PROJECT_SECRET"),
    maib.WithSignatureKey("ВАШ_SIGNATURE_KEY"), // Обязательно для проверки колбэков
)
```

### 2. Создание платежа (Direct Payment)

```go
req := &maib.PayRequest{
    Amount:      100.00,
    Currency:    maib.CurrencyMDL,
    ClientIP:    "127.0.0.1",
    Language:    maib.LanguageRO,
    Description: "Заказ #123",
    CallbackUrl: "https://your-site.com/callback",
    OkUrl:       "https://your-site.com/success",
    FailUrl:     "https://your-site.com/fail",
}

resp, err := client.Pay(ctx, req)
if err != nil {
    log.Fatal(err)
}

// Перенаправьте пользователя на страницу оплаты
fmt.Printf("Redirect to: %s\n", resp.PayURL)
```

### 3. Обработка колбэка (Webhook)

Пакет автоматически проверяет подпись входящего запроса.

```go
func handleCallback(w http.ResponseWriter, r *http.Request) {
    // ParseCallback проверяет подпись и декодирует JSON
    data, err := client.ParseCallback(r)
    if err != nil {
        log.Printf("Invalid callback: %v", err)
        http.Error(w, "Invalid signature", http.StatusBadRequest)
        return
    }

    if data.Status == "OK" {
        fmt.Printf("Платеж %s (Заказ %s) успешен!\n", data.PayID, data.OrderID)
    } else {
        fmt.Printf("Платеж не прошел: %s\n", data.StatusMessage)
    }

    w.WriteHeader(http.StatusOK)
}
```

### 4. Рекуррентные платежи (Привязка карты)

**Шаг 1: Регистрация карты**
Инициирует транзакцию (обычно на 0 или небольшую сумму) для сохранения карты.

```go
req := &maib.RegisterCardRequest{
    BillerExpiry: "1225", // Декабрь 2025
    ClientIP:     "127.0.0.1",
    Currency:     maib.CurrencyMDL,
    Language:     maib.LanguageRO,
    Email:        "user@example.com", // Обязательно для рекуррентных платежей
    CallbackUrl:  "https://your-site.com/callback",
    // ...
}

resp, err := client.RegisterCard(ctx, req)
// Перенаправьте пользователя на resp.PayURL
```

**Шаг 2: Получение BillerID из колбэка**
В колбэке регистрации вы получите `billerId`. Сохраните его в базе данных!

```go
// Внутри обработчика колбэка
if data.BillerID != "" {
    saveToDB(userID, data.BillerID)
}
```

**Шаг 3: Выполнение рекуррентного платежа**
Списание с сохраненной карты без участия пользователя.

```go
resp, err := client.ExecuteRecurring(ctx, &maib.ExecuteRecurringRequest{
    BillerID: "СОХРАНЕННЫЙ_BILLER_ID",
    Amount:   50.00,
    Currency: maib.CurrencyMDL,
    Description: "Ежемесячная подписка",
})

if resp.Status == "OK" {
    fmt.Println("Списано успешно!")
}
```

### 5. Возврат средств (Refund)

```go
resp, err := client.Refund(ctx, &maib.RefundRequest{
    PayID:        "ID_ТРАНЗАКЦИИ",
    RefundAmount: 50.00, // Опционально. Если не указать, вернется вся сумма.
})

if err != nil {
    log.Fatal(err)
}
fmt.Printf("Возврат успешен: %s\n", resp.Status)
```

### 6. Удаление сохраненной карты

```go
err := client.DeleteCard(ctx, "BILLER_ID_ДЛЯ_УДАЛЕНИЯ")
if err != nil {
    log.Printf("Ошибка удаления карты: %v", err)
} else {
    fmt.Println("Карта успешно удалена")
}
```

### 7. Helper методы

*   `GenerateToken(ctx)`: Обычно вызывается автоматически, но можно вызвать вручную для получения токена.
*   `ParseCallback(r)`: Проверяет подпись и возвращает типизированную структуру `CallbackData`.

## Ссылки

*   [Официальная документация](https://docs.maibmerchants.md/e-commerce/maib-e-commerce-api)
*   [Эталонный репозиторий](https://github.com/acolev/maib-ecomm)
