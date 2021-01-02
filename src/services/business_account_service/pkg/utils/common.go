package utils

import (
	"bytes"
	"encoding/json"
)

// CreateRequestBody converts any type to bytes
func CreateRequestBody(body interface{}) (*bytes.Reader, error) {
	b, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	return bytes.NewReader(b), nil
}
