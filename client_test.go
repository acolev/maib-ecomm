package maib

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClient_GenerateToken(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/generate-token" {
			t.Errorf("Expected path /generate-token, got %s", r.URL.Path)
		}
		if r.Method != "POST" {
			t.Errorf("Expected method POST, got %s", r.Method)
		}
		
		var req map[string]string
		json.NewDecoder(r.Body).Decode(&req)
		if req["projectId"] != "pid" {
			t.Errorf("Expected projectId pid, got %s", req["projectId"])
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(TokenResponse{
			AccessToken: "access",
			ExpiresIn:   3600,
		})
	}))
	defer ts.Close()
	
	c := NewClient(WithProjectID("pid"), WithProjectSecret("sec"), WithBaseURL(ts.URL))
	token, err := c.GenerateToken(context.Background())
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}
	
	if token.AccessToken != "access" {
		t.Errorf("Expected access token 'access', got %s", token.AccessToken)
	}
}

func TestClient_Pay(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/generate-token" {
			json.NewEncoder(w).Encode(TokenResponse{AccessToken: "token", ExpiresIn: 3600})
			return
		}
		
		if r.URL.Path != "/pay" {
			t.Errorf("Expected path /pay, got %s", r.URL.Path)
		}
		
		auth := r.Header.Get("Authorization")
		if auth != "Bearer token" {
			t.Errorf("Expected Authorization Bearer token, got %s", auth)
		}
		
		w.Header().Set("Content-Type", "application/json")
		// Response is wrapped in "result"
		response := map[string]interface{}{
			"result": PayResponse{
				PayID:   "pay123",
				PayURL:  "http://pay.url",
				OrderID: "ord123",
			},
			"ok": true,
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer ts.Close()
	
	c := NewClient(WithProjectID("pid"), WithProjectSecret("sec"), WithBaseURL(ts.URL))
	
	// Ensure token is fetched
	res, err := c.Pay(context.Background(), &PayRequest{
		Amount: 100.0,
		Currency: CurrencyMDL,
	})
	if err != nil {
		t.Fatalf("Pay failed: %v", err)
	}
	
	if res.PayID != "pay123" {
		t.Errorf("Expected PayID pay123, got %s", res.PayID)
	}
}

func TestClient_Refund_Error(t *testing.T) {
    ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if r.URL.Path == "/generate-token" {
            json.NewEncoder(w).Encode(TokenResponse{AccessToken: "token", ExpiresIn: 3600})
            return
        }
        
        w.Header().Set("Content-Type", "application/json")
        response := ErrorResponse{
            OK: false,
            Errors: []ErrorItem{
                {ErrorCode: "ERR01", ErrorMessage: "Bad Refund"},
            },
        }
        json.NewEncoder(w).Encode(response)
    }))
    defer ts.Close()
    
    c := NewClient(WithProjectID("pid"), WithProjectSecret("sec"), WithBaseURL(ts.URL))
    
    _, err := c.Refund(context.Background(), &RefundRequest{PayID: "pid"})
    if err == nil {
        t.Fatal("Expected error, got nil")
    }
    
    apiErr, ok := err.(*APIError)
    if !ok {
        t.Fatalf("Expected *APIError, got %T", err)
    }
    
    if len(apiErr.Errors) != 1 || apiErr.Errors[0].ErrorCode != "ERR01" {
        t.Errorf("Unexpected error content: %v", apiErr)
    }
}
