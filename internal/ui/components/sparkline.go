package components

import (
	"strings"
)

var sparklineRunes = []rune{' ', '▂', '▃', '▄', '▅', '▆', '▇', '█'}

// RenderSparkline generates a unicode sparkline string for a series of numeric values.
func RenderSparkline(values []float64, maxBars int) string {
	if len(values) == 0 {
		return ""
	}

	data := values
	if maxBars > 0 && len(data) > maxBars {
		data = data[len(data)-maxBars:]
	}

	// Determine min and max
	var minVal, maxVal float64
	minVal = data[0]
	maxVal = data[0]

	for _, v := range data {
		if v < minVal {
			minVal = v
		}
		if v > maxVal {
			maxVal = v
		}
	}

	diff := maxVal - minVal
	var sb strings.Builder

	for _, v := range data {
		if diff == 0 || maxVal == 0 {
			if v > 0 {
				sb.WriteRune(sparklineRunes[3])
			} else {
				sb.WriteRune(sparklineRunes[0])
			}
			continue
		}

		normalized := (v - minVal) / diff
		idx := int(normalized * float64(len(sparklineRunes)-1))
		if idx < 0 {
			idx = 0
		}
		if idx >= len(sparklineRunes) {
			idx = len(sparklineRunes) - 1
		}
		sb.WriteRune(sparklineRunes[idx])
	}

	return sb.String()
}
