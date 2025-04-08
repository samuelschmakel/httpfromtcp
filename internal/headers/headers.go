package headers

import (
	"bytes"
	"fmt"
	"strings"
	"unicode"
)

const crlf = "\r\n"
const specialChars = "!#$%&'*+-.^_`|~"

type Headers map[string]string

func NewHeaders() Headers {
	return map[string]string{}
}

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	idx := bytes.Index(data, []byte(crlf))

	// Assume not enough data
	if idx == -1 {
		return 0, false, nil
	}

	// End of headers
	if idx == 0 {
		return len(crlf), true, nil
	}

	headersText := string(data[:idx])

	// Removes trailing and leading whitespace
	headersText = strings.TrimSpace(headersText)

	// Split on first colon
	colonIdx := strings.Index(headersText, ":")
	if colonIdx == -1 {
		return 0, false, fmt.Errorf("invalid header: missing colon")
	}

	fieldName, fieldValue := headersText[:colonIdx], headersText[colonIdx+len(":"):]

	if unicode.IsSpace(rune(fieldName[len(fieldName)-1])) {
		return 0, false, fmt.Errorf("invalid header: whitespace between field name and colon")
	}

	fieldName = strings.TrimSpace(fieldName)
	fieldValue = strings.TrimSpace(fieldValue)

	// Check for invalid characters in fieldName
	if !validTokens(fieldName, specialChars) {
		return 0, false, fmt.Errorf("invalid character found in fieldName")
	}

	fieldName = strings.ToLower(fieldName)

	// Check if map key exists
	value, ok := h[fieldName]
	if ok {
		h[fieldName] = value + ", " + fieldValue
	} else {
		// Add it to the map
		h[fieldName] = fieldValue
	}
	
	return idx + len(crlf), false, nil
}

func (h Headers) Get(key string) (string, bool) {
	key = strings.ToLower(key)
	v, ok := h[key]
	return v, ok
}

func validTokens(s string, specialChars string) bool {
	for _, r := range s {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && !strings.ContainsRune(specialChars, r) {
			return false
		}
	}
	return true
}

func (h Headers) Override(key string, value string) {
	h[key] = value
}