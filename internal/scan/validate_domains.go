package scan

import (
	"log/slog"
	"regexp"
	"strings"
)

func ValidateDomains(domains []string) []string {
	domainRegex := regexp.MustCompile(`^([a-zA-Z0-9-]+\.)+[a-zA-Z]{2,}$`)

	var valid []string
	for _, d := range domains {
		d = strings.TrimSpace(strings.ToLower(d))

		if len(d) == 0 || len(d) > 253 {
			slog.Warn("Dominio ignorado: longitud inválida", "domain", d)
			continue
		}

		if !domainRegex.MatchString(d) {
			slog.Warn("Dominio ignorado: formato inválido", "domain", d)
			continue
		}

		valid = append(valid, d)
	}
	return valid
}
