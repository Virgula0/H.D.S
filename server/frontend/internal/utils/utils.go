package utils

import (
	"encoding/base64"
	"regexp"
	"strings"
)

// base64urlRegex ensures only valid base64url characters are allowed
var base64urlRegex = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)

// IsJWT validates whether the given string is a valid JWT token format
func IsJWT(token string) bool {
	// JWT must have exactly 3 parts (header.payload.signature)
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return false
	}

	// Validate each part for base64url encoding
	for _, part := range parts {
		if part == "" || !base64urlRegex.MatchString(part) {
			return false
		}
	}

	return true
}

// isBase64URLEncoded checks if a string is Base64 URL encoded.
func isBase64URLEncoded(s string) bool {
	_, err := base64.RawURLEncoding.DecodeString(s)
	return err == nil
}

func Base64StringToString(input string) (string, error) {
	decoded, err := base64.StdEncoding.DecodeString(input)

	if err != nil {
		return "", err
	}

	return string(decoded), nil
}

// IsValidUUIDv4 checks if a string is a valid UUID v4.
func IsValidUUIDv4(uuid string) bool {
	re := regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-4[0-9a-fA-F]{3}-[89ABab][0-9a-fA-F]{3}-[0-9a-fA-F]{12}$`)
	return re.MatchString(uuid)
}
