package parsers

import (
	"strconv"
	"strings"
)

// extract value after specific prefix in a field
func extractFieldValue(field, prefix string) (int, error) {
	if strings.HasPrefix(field, prefix) {
		valueStr := strings.TrimPrefix(field, prefix)
		valueStr = strings.ReplaceAll(valueStr, ",", "")
		return strconv.Atoi(valueStr)
	}
	return 0, nil
}
