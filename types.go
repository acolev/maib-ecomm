package maib

// Currency represents the currency of the transaction.
type Currency string

const (
	CurrencyMDL Currency = "MDL"
	CurrencyUSD Currency = "USD"
	CurrencyEUR Currency = "EUR"
)

// Language represents the language of the checkout page.
type Language string

const (
	LanguageRO Language = "ro"
	LanguageRU Language = "ru"
	LanguageEN Language = "en"
)

// Common response structures

// ErrorResponse represents the API error structure.
type ErrorResponse struct {
	OK     bool        `json:"ok"`
	Errors []ErrorItem `json:"errors,omitempty"`
}

type ErrorItem struct {
	ErrorCode    string                 `json:"errorCode"`
	ErrorMessage string                 `json:"errorMessage"`
	ErrorArgs    map[string]interface{} `json:"errorArgs,omitempty"`
}

// TokenResponse represents the response from /generate-token.
type TokenResponse struct {
	AccessToken      string `json:"accessToken"`
	ExpiresIn        int    `json:"expiresIn"`
	RefreshToken     string `json:"refreshToken"`
	RefreshExpiresIn int    `json:"refreshExpiresIn"`
	TokenType        string `json:"tokenType"`
}

// PayRequest represents parameters for /pay (Direct Payment).
type PayRequest struct {
	Amount      float64  `json:"amount"`
	Currency    Currency `json:"currency"`
	ClientIP    string   `json:"clientIp"`
	Language    Language `json:"language"`
	Description string   `json:"description,omitempty"`
	ClientName  string   `json:"clientName,omitempty"`
	Email       string   `json:"email,omitempty"`
	Phone       string   `json:"phone,omitempty"`
	OrderID     string   `json:"orderId,omitempty"`
	Delivery    float64  `json:"delivery,omitempty"`
	Items       []Item   `json:"items,omitempty"`
	CallbackUrl string   `json:"callbackUrl,omitempty"`
	OkUrl       string   `json:"okUrl,omitempty"`
	FailUrl     string   `json:"failUrl,omitempty"`
}

type Item struct {
	ID       string  `json:"id,omitempty"`
	Name     string  `json:"name,omitempty"`
	Price    float64 `json:"price,omitempty"`
	Quantity int     `json:"quantity,omitempty"`
}

// PayResponse represents the intermediate response from /pay.
type PayResponse struct {
	PayID   string `json:"payId"`
	OrderID string `json:"orderId"`
	PayURL  string `json:"payUrl"`
}

// CallbackData represents the final transaction data received on Callback URL.
type CallbackData struct {
	PayID         string  `json:"payId"`
	OrderID       string  `json:"orderId"`
	Status        string  `json:"status"`
	StatusCode    string  `json:"statusCode"`
	StatusMessage string  `json:"statusMessage"`
	ThreeDs       string  `json:"threeDs"`
	RRN           string  `json:"rrn"`
	Approval      string  `json:"approval"`
	CardNumber    string  `json:"cardNumber"`
	Amount        float64 `json:"amount,string"` // API returns string, but we can try to parse it. Wait, doc says "10.25". JSON number or string?
	// Doc says "Amount: String" for callback, but "number(decimal)" for request. PHP SDK might treat as string.
	// Let's use string/float depending on context or interface{} to be safe in signature verification.
	// Actually for the struct we return to user, we want strong types.
	// But for signature verification we need raw values.
	Currency string `json:"currency"`
}

// Used for parsing raw callback JSON to verify signature
type rawCallbackContainer struct {
	Result    map[string]interface{} `json:"result"`
	Signature string                 `json:"signature"`
}

// RegisterCardRequest represents parameters for /savecard-recurring.
type RegisterCardRequest struct {
	BillerExpiry string   `json:"billerExpiry"` // MMYY
	ClientIP     string   `json:"clientIp"`
	Currency     Currency `json:"currency"`
	Language     Language `json:"language"`
	Email        string   `json:"email"` // Required for recurring
	Amount       float64  `json:"amount,omitempty"`
	Description  string   `json:"description,omitempty"`
	ClientName   string   `json:"clientName,omitempty"`
	Phone        string   `json:"phone,omitempty"`
	OrderID      string   `json:"orderId,omitempty"`
	Delivery     float64  `json:"delivery,omitempty"`
	Items        []Item   `json:"items,omitempty"`
	CallbackUrl  string   `json:"callbackUrl,omitempty"`
	OkUrl        string   `json:"okUrl,omitempty"`
	FailUrl      string   `json:"failUrl,omitempty"`
}

// RegisterCardResponse is the intermediate response.
type RegisterCardResponse struct {
	PayID   string `json:"payId"`
	OrderID string `json:"orderId"`
	PayURL  string `json:"payUrl"`
}

// ExecuteRecurringRequest represents parameters for /execute-recurring.
type ExecuteRecurringRequest struct {
	BillerID    string   `json:"billerId"`
	Amount      float64  `json:"amount"`
	Currency    Currency `json:"currency"`
	Description string   `json:"description,omitempty"`
	OrderID     string   `json:"orderId,omitempty"`
	Delivery    float64  `json:"delivery,omitempty"`
	Items       []Item   `json:"items,omitempty"`
}

// ExecuteRecurringResponse represents response from /execute-recurring.
// Note: This one returns "result" directly with status, or intermediate?
// Doc says "Response parameters: result Object ... status OK". So it's immediate if successful?
// Or does it require 3DS?
// Doc for Execute Recurring says response has "status", "rrn", "approval".
// It seems it's a direct server-to-server call usually, but let's check if there's a payUrl.
// Doc doesn't list payUrl in response for execute-recurring.
type ExecuteRecurringResponse struct {
	PayID         string  `json:"payId"`
	BillerID      string  `json:"billerId"`
	OrderID       string  `json:"orderId"`
	Status        string  `json:"status"`
	StatusCode    string  `json:"statusCode"`
	StatusMessage string  `json:"statusMessage"`
	RRN           string  `json:"rrn"`
	Approval      string  `json:"approval"`
	CardNumber    string  `json:"cardNumber"`
	Amount        float64 `json:"amount"`
	Currency      string  `json:"currency"`
}

// RefundRequest represents parameters for /refund.
type RefundRequest struct {
	PayID        string  `json:"payId"`
	RefundAmount float64 `json:"refundAmount,omitempty"`
}

// RefundResponse represents response from /refund.
type RefundResponse struct {
	PayID         string  `json:"payId"`
	OrderID       string  `json:"orderId"`
	Status        string  `json:"status"`
	StatusCode    string  `json:"statusCode"`
	StatusMessage string  `json:"statusMessage"`
	RefundAmount  float64 `json:"refundAmount"`
}

// DeleteCardRequest represents parameters for /delete-card.
type DeleteCardRequest struct {
	BillerID string `json:"billerId"`
}
