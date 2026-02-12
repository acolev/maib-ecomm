package maib

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"sort"
	"strings"
	"encoding/json"
	"errors"
	"io"
	"net/http"
    "strconv"
)

// ParseCallback verifies the signature and decodes the callback data.
// It reads the body from the request.
func (c *Client) ParseCallback(r *http.Request) (*CallbackData, error) {
	if c.signatureKey == "" {
		return nil, errors.New("signature key is not set")
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read body: %w", err)
	}
	defer r.Body.Close()

	var raw map[string]interface{}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}
	
	resultMap, ok := raw["result"].(map[string]interface{})
	if !ok {
		return nil, errors.New("missing result object in callback")
	}
	
	signature, ok := raw["signature"].(string)
	if !ok {
		return nil, errors.New("missing signature in callback")
	}
	
	if !c.VerifySignature(resultMap, signature) {
		return nil, errors.New("invalid signature")
	}
	
	// Decode into struct
	// We can reuse the body or marshal the map
	// Re-marshalling is safer as it ensures typed structure
	resultBytes, _ := json.Marshal(resultMap)
	var data CallbackData
	if err := json.Unmarshal(resultBytes, &data); err != nil {
		return nil, fmt.Errorf("failed to decode callback data: %w", err)
	}
	
	return &data, nil
}

// VerifySignature checks if the signature matches the data using the client's signature key.
func (c *Client) VerifySignature(data map[string]interface{}, receivedSignature string) bool {
    // 1. Sort data by keys recursively
    // 2. Append signature key
    // 3. Implode with ':'
    // 4. Base64(SHA256(string))
    
    // We clone the map to avoid modifying the original if passed by reference (though here it's map so it is ref)
    // But we are not actually appending to the map in the PHP sense, 
    // PHP does `$sortedDataByKeys[] = SIGNATURE_KEY;` which treats it as a list addition?
    // Wait. In PHP:
    // $sortedDataByKeys = sortByKeyRecursive($data_result);
    // $sortedDataByKeys[] = SIGNATURE_KEY;
    // If $sortedDataByKeys is `['a'=>1]`, doing `[] = key` makes it `['a'=>1, 0=>key]`.
    // Then `implodeRecursive` iterates.
    // Iteration order in PHP for mixed array: keys are preserved.
    // The numeric key `0` (or whatever) comes after the string keys if added later?
    // Actually PHP arrays are ordered maps. Insertion order matters for iteration unless sorted.
    // `ksort` implementation:
    // If keys are mixed string/int, it sorts them.
    // `ksort` sorts by key.
    // So if I have `{"a": 1}` represented as `['a'=>1]`.
    // And I add `['0' => KEY]`.
    // `ksort` will put `0` before `a`?
    // Wait, the PHP code sorts *BEFORE* adding the signature key.
    // `$sortedDataByKeys = sortByKeyRecursive($data_result);` -> this returns a sorted array.
    // `$sortedDataByKeys[] = SIGNATURE_KEY;` -> appends to the end.
    // The keys are NOT re-sorted after adding the signature key.
    // So the signature key is effectively the LAST element.
    // So we just need to implode the sorted data, AND THEN append `:<SIGNATURE_KEY>`?
    // Let's verify `implodeRecursive`.
    
    /*
    $sortedDataByKeys[] = SIGNATURE_KEY;
    $signString = implodeRecursive(':', $sortedDataByKeys);
    
    function implodeRecursive($separator, $array) {
        $result = '';
        foreach ($array as $item) {
            $result .= (is_array($item) ? implodeRecursive($separator, $item) : (string)$item) . $separator;
        }
        return substr($result, 0, -1);
    }
    */
    
    // So yes, it iterates the array. Since `[]=` appends to the end of the array order, 
    // and `sortByKeyRecursive` was called *before* this append,
    // The signature key is indeed the last item.
    // So `signString` = `implodeRecursive(sortedMap) + ":" + SIGNATURE_KEY`.
    
    serialized := implodeRecursive(data)
    toSign := serialized + ":" + c.signatureKey
    
    hash := sha256.Sum256([]byte(toSign))
    computed := base64.StdEncoding.EncodeToString(hash[:])
    
    return computed == receivedSignature
}

func implodeRecursive(data interface{}) string {
    switch v := data.(type) {
    case map[string]interface{}:
        // Sort keys
        keys := make([]string, 0, len(v))
        for k := range v {
            keys = append(keys, k)
        }
        sort.Strings(keys)
        
        var parts []string
        for _, k := range keys {
            parts = append(parts, implodeRecursive(v[k]))
        }
        return strings.Join(parts, ":")
        
    case []interface{}:
        // Array/List. In JSON, this is a list.
        // PHP `json_decode(..., true)` makes it an indexed array.
        // `sortByKeyRecursive` uses `ksort`. For `[0=>'a', 1=>'b']`, keys are 0,1.
        // They are already sorted.
        // So we just iterate.
        var parts []string
        for _, item := range v {
            parts = append(parts, implodeRecursive(item))
        }
        return strings.Join(parts, ":")
        
    default:
        return toString(v)
    }
}

func toString(v interface{}) string {
    switch val := v.(type) {
    case string:
        return val
    case float64:
        // JSON numbers are float64 in Go interface{}.
        // PHP `(string)10.25` is "10.25".
        // PHP `(string)10` is "10".
        // We need to format carefully. `strconv.FormatFloat` with 'f', -1, 64 usually works but trims/expands.
        // However, if the JSON had "10.25", Go might have parsed it such that printing it is consistent.
        // But what if it's integer 10? `(string)10` -> "10". `(string)10.0` -> "10".
        return strconv.FormatFloat(val, 'f', -1, 64)
    case bool:
        if val {
            return "1"
        }
        return "" // PHP false is empty string
    case nil:
        return ""
    default:
        return fmt.Sprintf("%v", val)
    }
}
