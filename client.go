package maib

import (
	"net/http"
	"sync"
	"time"
)

const (
	defaultBaseURL = "https://api.maibmerchants.md/v1"
)

// Client is the main entry point for the Maib E-commerce API.
type Client struct {
	projectID     string
	projectSecret string
	signatureKey  string
	
	baseURL    string
	httpClient *http.Client

	// Token management
	tokenMu          sync.RWMutex
	accessToken      string
	refreshToken     string
	accessTokenExp   time.Time
	refreshTokenExp  time.Time
}

// NewClient creates a new Maib client with the given options.
// PROJECT_ID and PROJECT_SECRET are required for API calls.
// SIGNATURE_KEY is required for callback validation.
func NewClient(options ...Option) *Client {
	c := &Client{
		baseURL:    defaultBaseURL,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}

	for _, opt := range options {
		opt(c)
	}

	return c
}

// Option is a functional option for configuring the Client.
type Option func(*Client)

// WithProjectID sets the Project ID.
func WithProjectID(id string) Option {
	return func(c *Client) {
		c.projectID = id
	}
}

// WithProjectSecret sets the Project Secret.
func WithProjectSecret(secret string) Option {
	return func(c *Client) {
		c.projectSecret = secret
	}
}

// WithSignatureKey sets the Signature Key used for callback validation.
func WithSignatureKey(key string) Option {
	return func(c *Client) {
		c.signatureKey = key
	}
}

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(client *http.Client) Option {
	return func(c *Client) {
		c.httpClient = client
	}
}

// WithBaseURL sets a custom Base URL (e.g. for testing).
func WithBaseURL(url string) Option {
	return func(c *Client) {
		c.baseURL = url
	}
}
