package common

import (
	"encoding/base64"
	"fmt"
)

// DecodeConnectionString decodes a base64 encoded connection string.
// TODO: Implement proper error handling and potentially decryption if needed.
func DecodeConnectionString(encoded string) (string, error) {
	if encoded == "" {
		return "", fmt.Errorf("encoded connection string is empty")
	}
	dcodedBytes, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64 connection string: %w", err)
	}
	return string(dcodedBytes), nil
}
