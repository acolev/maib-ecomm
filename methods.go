package maib

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
    "time"
)

// GenerateToken obtains a new access token using Project ID and Project Secret.
func (c *Client) GenerateToken(ctx context.Context) (*TokenResponse, error) {
	reqBody := map[string]string{
		"projectId":     c.projectID,
		"projectSecret": c.projectSecret,
	}
	
	var res TokenResponse
	if err := c.doRequest(ctx, http.MethodPost, "/generate-token", reqBody, &res); err != nil {
		return nil, err
	}
	
	// Cache the token
	c.tokenMu.Lock()
	c.accessToken = res.AccessToken
	c.accessTokenExp = time.Now().Add(time.Duration(res.ExpiresIn) * time.Second)
	c.refreshToken = res.RefreshToken
	c.refreshTokenExp = time.Now().Add(time.Duration(res.RefreshExpiresIn) * time.Second)
	c.tokenMu.Unlock()
	
	return &res, nil
}

// Pay initiates a direct payment.
func (c *Client) Pay(ctx context.Context, req *PayRequest) (*PayResponse, error) {
    if err := c.ensureToken(ctx); err != nil {
        return nil, err
    }
    
    var res map[string]interface{} // Intermediate response wrapper
    if err := c.doRequest(ctx, http.MethodPost, "/pay", req, &res); err != nil {
        return nil, err
    }
    
    // Extract result
    result, ok := res["result"].(map[string]interface{})
    if !ok {
        return nil, fmt.Errorf("invalid response format: missing result")
    }
    
    // Marshal back to struct
    paramsBytes, _ := json.Marshal(result)
    var payRes PayResponse
    if err := json.Unmarshal(paramsBytes, &payRes); err != nil {
        return nil, err
    }
    
    return &payRes, nil
}

// RegisterCard initiates a recurring card registration.
func (c *Client) RegisterCard(ctx context.Context, req *RegisterCardRequest) (*RegisterCardResponse, error) {
    if err := c.ensureToken(ctx); err != nil {
        return nil, err
    }
    
    var res map[string]interface{}
    if err := c.doRequest(ctx, http.MethodPost, "/savecard-recurring", req, &res); err != nil {
        return nil, err
    }
    
    result, ok := res["result"].(map[string]interface{})
    if !ok {
        return nil, fmt.Errorf("invalid response format: missing result")
    }
    
    paramsBytes, _ := json.Marshal(result)
    var regRes RegisterCardResponse
    if err := json.Unmarshal(paramsBytes, &regRes); err != nil {
        return nil, err
    }
    
    return &regRes, nil
}

// ExecuteRecurring executes a recurring payment.
func (c *Client) ExecuteRecurring(ctx context.Context, req *ExecuteRecurringRequest) (*ExecuteRecurringResponse, error) {
     if err := c.ensureToken(ctx); err != nil {
        return nil, err
    }
    
    var res map[string]interface{}
    if err := c.doRequest(ctx, http.MethodPost, "/execute-recurring", req, &res); err != nil {
        return nil, err
    }
    
    result, ok := res["result"].(map[string]interface{})
    if !ok {
        return nil, fmt.Errorf("invalid response format: missing result")
    }
    
    paramsBytes, _ := json.Marshal(result)
    var execRes ExecuteRecurringResponse
    if err := json.Unmarshal(paramsBytes, &execRes); err != nil {
        return nil, err
    }
    
    return &execRes, nil
}

// Refund refunds a transaction.
func (c *Client) Refund(ctx context.Context, req *RefundRequest) (*RefundResponse, error) {
     if err := c.ensureToken(ctx); err != nil {
        return nil, err
    }
    
    var res map[string]interface{}
    if err := c.doRequest(ctx, http.MethodPost, "/refund", req, &res); err != nil {
        return nil, err
    }
    
    result, ok := res["result"].(map[string]interface{})
    if !ok {
        return nil, fmt.Errorf("invalid response format: missing result")
    }
    
    paramsBytes, _ := json.Marshal(result)
    var refundRes RefundResponse
    if err := json.Unmarshal(paramsBytes, &refundRes); err != nil {
        return nil, err
    }
    
    return &refundRes, nil
}

// DeleteCard deletes a saved card.
// Note: Documentation URL is https://api.maibmerchants.md/v1/delete-card but need to check params.
// The list early said "delete-card".
func (c *Client) DeleteCard(ctx context.Context, billerID string) error {
      if err := c.ensureToken(ctx); err != nil {
        return err
    }
    
    req := map[string]string{"billerId": billerID}
    return c.doRequest(ctx, http.MethodPost, "/delete-card", req, nil)
}

// Internal helper to ensure we have a valid token
func (c *Client) ensureToken(ctx context.Context) error {
    c.tokenMu.RLock()
    valid := c.accessToken != "" && time.Now().Before(c.accessTokenExp)
    c.tokenMu.RUnlock()
    
    if valid {
        return nil
    }
    
    // Token expired or not set, generate new one
    // TODO: Use RefreshToken if available? For simplicity, we just regenerate using credentials.
    // Doc says token lives for a while, but generating new one is safe.
    _, err := c.GenerateToken(ctx)
    return err
}

// doRequest performs the HTTP request
func (c *Client) doRequest(ctx context.Context, method, endpoint string, body interface{}, result interface{}) error {
	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+endpoint, bodyReader)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	
	c.tokenMu.RLock()
	token := c.accessToken
	c.tokenMu.RUnlock()
	
	// Don't add auth header for generate-token itself
	if token != "" && endpoint != "/generate-token" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

    // Check for API errors (JSON with "ok": false)
	var errRes ErrorResponse
    // We try to unmarshal into error response first.
    // Note: The API returns { "result": ..., "ok": true } OR { "errors": [...], "ok": false }
    if err := json.Unmarshal(respBody, &errRes); err == nil && !errRes.OK && len(errRes.Errors) > 0 {
        return &APIError{
            Errors: errRes.Errors,
        }
    }
    
    // If not an error, unmarshal into result
    if result != nil {
        if err := json.Unmarshal(respBody, result); err != nil {
             return fmt.Errorf("failed to unmarshal response: %w. Body: %s", err, string(respBody))
        }
    } else {
         // If result is nil, we just check if "ok" is true in strict mode, but here we assume if no error object, it's fine.
         // Actually delete-card might return just { "ok": true }?
    }

	return nil
}
