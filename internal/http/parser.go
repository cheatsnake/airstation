package http

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
)

func parseIntQuery(queries url.Values, key string, defaultValue int) int {
	queryValue := queries.Get(key)
	parsed, err := strconv.Atoi(queryValue)
	if err != nil {
		parsed = defaultValue
	}

	return parsed
}

func parseJSONBody[T any](r *http.Request) (*T, error) {
	rawBytes, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	if len(rawBytes) == 0 {
		return nil, fmt.Errorf("request body is empty")
	}

	var jsonData T
	if err := json.Unmarshal(rawBytes, &jsonData); err != nil {
		return nil, err
	}

	return &jsonData, nil
}
