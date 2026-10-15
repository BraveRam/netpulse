package enricher

import (
	"strings"
	"unicode"
)

// CountryToFlag converts an ISO 3166-1 alpha-2 country code to a Unicode flag emoji.
// For example: "US" -> "🇺🇸", "DE" -> "🇩🇪", "LOCAL" -> "📍".
func CountryToFlag(countryCode string) string {
	code := strings.TrimSpace(strings.ToUpper(countryCode))
	switch code {
	case "", "??", "UNKNOWN":
		return "🌐"
	case "LOCAL", "LAN", "LOOPBACK":
		return "📍"
	}

	if len(code) != 2 {
		return "🌐"
	}

	r0 := rune(code[0])
	r1 := rune(code[1])

	if !unicode.IsUpper(r0) || !unicode.IsUpper(r1) {
		return "🌐"
	}

	// Regional Indicator Symbol Letter A is 0x1F1E6
	const regionalIndicatorBase = 0x1F1E6
	firstRune := regionalIndicatorBase + (r0 - 'A')
	secondRune := regionalIndicatorBase + (r1 - 'A')

	return string([]rune{firstRune, secondRune})
}
