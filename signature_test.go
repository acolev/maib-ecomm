package maib

import (
	"encoding/json"
	"testing"
)

func TestImplodeRecursive(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected string
	}{
		{
			name:     "simple map",
			input:    map[string]interface{}{"a": 1, "b": "2"},
			expected: "1:2",
		},
		{
			name:     "nested map",
			input:    map[string]interface{}{"a": 1, "b": map[string]interface{}{"c": 2, "d": 3}},
			expected: "1:2:3",
		},
		{
			name:     "mixed types",
			input:    map[string]interface{}{"a": true, "b": nil, "c": 10.5},
			expected: "1::10.5",
		},
		{
			name:     "list",
			input:    []interface{}{"a", "b"},
			expected: "a:b",
		},
		{
			name:     "mixed list map",
			input:    map[string]interface{}{"x": []interface{}{"a", "b"}, "y": "c"},
			expected: "a:b:c",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := implodeRecursive(tt.input)
			if got != tt.expected {
				t.Errorf("implodeRecursive() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestVerifySignature(t *testing.T) {
	client := NewClient(WithSignatureKey("SECRET"))

	// {"a": 1, "b": 2} -> "1:2:SECRET"
	// echo -n "1:2:SECRET" | openssl dgst -sha256 -binary | openssl base64
	// SHA256("1:2:SECRET") =
	// 3a6eb0790f39ac87c94f3856b2dd2c5d110e6811602261a9a923d3bb23adc8b7
	// Base64 = Om6weQ85rIfJTzhWst0sXREoaBFgImGjqSPTuyOtyLc=

	// We need to ensure 1.0 is treated as "1".
	// My implementation uses strconv.FormatFloat(val, 'f', -1, 64).
	// 1.0 -> "1". Correct.

	// Wait, if input is map[string]interface{}{"a": 1}, 1 is int.
	// But coming from JSON unmarshal it is float64.
	// So let's simulate JSON unmarshal.

	jsonData := `{"a": 1, "b": "2"}`
	var raw map[string]interface{}
	json.Unmarshal([]byte(jsonData), &raw)

	validSig := "JHvWuPNVVIkPbFqdAxP56oKpolcxcUtPyrAYGsYMOms="

	if !client.VerifySignature(raw, validSig) {
		t.Errorf("VerifySignature failed for valid signature. Parsed data: %v", raw)
	}
}
