package kubernetes

import "unicode"

func int32Ptr(i int32) *int32 { return &i }

func kubernetesAcceptedNamespace(s string) string {
	rs := make([]rune, 0, len(s))
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsNumber(r) || r == '.' || r == '-' {
			rs = append(rs, unicode.ToLower(r))
		}
	}
	return string(rs)
}
