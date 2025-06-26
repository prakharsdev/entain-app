package utils

import "strings"

var allowedSourceTypes = map[string]struct{}{
	"game":    {},
	"server":  {},
	"payment": {},
}

// IsValidSourceType checks if the Source-Type header is valid (case-insensitive)
func IsValidSourceType(source string) bool {
	_, ok := allowedSourceTypes[strings.ToLower(source)]
	return ok
}

// IsValidState checks if the state is either "win" or "lose"
func IsValidState(state string) bool {
	return state == "win" || state == "lose"
}
