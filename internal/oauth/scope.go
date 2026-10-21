package oauth

import (
	"sort"
	"strings"
)

const (
	ScopeOpenID  = "openid"
	ScopeProfile = "profile"
)

func NormalizeScope(value string) (string, bool) {
	requested := strings.Fields(value)
	accepted := make(map[string]struct{})

	for _, scope := range requested {
		switch scope {
		case ScopeOpenID, ScopeProfile:
			accepted[scope] = struct{}{}
		default:
			return "", false
		}
	}

	if _, exists := accepted[ScopeOpenID]; !exists {
		return "", false
	}

	result := make([]string, 0, len(accepted))
	for scope := range accepted {
		result = append(result, scope)
	}

	sort.Strings(result)

	return strings.Join(result, " "), true
}

func ScopeContains(scope string, expected string) bool {
	for _, value := range strings.Fields(scope) {
		if value == expected {
			return true
		}
	}

	return false
}
