package utils

import (
	"strings"
)

func IsSystemRole(role string) bool {
	switch strings.ToLower(strings.TrimSpace(role)) {
	case "admin", "owner":
		return true
	default:
		return false
	}
}
