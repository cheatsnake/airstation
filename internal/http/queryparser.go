package http

import (
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
