package services

import (
	"math"
	"strconv"
	"strings"
)

// fractionMap pairs decimal fraction values with unicode glyphs.
// Order matters only for human reading; lookup uses tolerance match.
var fractionMap = []struct {
	decimal float64
	glyph   string
}{
	{0.125, "⅛"},  // ⅛
	{0.25, "¼"},   // ¼
	{1.0 / 3.0, "⅓"}, // ⅓
	{0.375, "⅜"},  // ⅜
	{0.5, "½"},    // ½
	{0.625, "⅝"},  // ⅝
	{2.0 / 3.0, "⅔"}, // ⅔
	{0.75, "¾"},   // ¾
	{0.875, "⅞"},  // ⅞
}

const fractionTolerance = 0.02

// ParseQuantity parses "1/2", "1 1/2", "0.5", "1.5", "2" into a float.
// Returns 0 if no leading numeric prefix.
func ParseQuantity(s string) float64 {
	parts := strings.Fields(s)
	var total float64
	consumed := 0
	for _, p := range parts {
		v, ok := parseNumericToken(p)
		if !ok {
			break
		}
		total += v
		consumed++
	}
	if consumed == 0 {
		return 0
	}
	return total
}

func parseNumericToken(p string) (float64, bool) {
	if strings.Contains(p, "/") {
		nd := strings.SplitN(p, "/", 2)
		n, errN := strconv.ParseFloat(nd[0], 64)
		d, errD := strconv.ParseFloat(nd[1], 64)
		if errN != nil || errD != nil || d == 0 {
			return 0, false
		}
		return n / d, true
	}
	v, err := strconv.ParseFloat(p, 64)
	if err != nil {
		return 0, false
	}
	return v, true
}

// NormalizeIngredient converts any leading mixed-fraction prefix to a decimal.
// "1/2 cup flour" -> "0.5 cup flour"; "1 1/2 cups sugar" -> "1.5 cups sugar".
func NormalizeIngredient(s string) string {
	parts := strings.Fields(s)
	var total float64
	consumed := 0
	for _, p := range parts {
		v, ok := parseNumericToken(p)
		if !ok {
			break
		}
		total += v
		consumed++
	}
	if consumed == 0 {
		return s
	}
	rest := strings.Join(parts[consumed:], " ")
	qtyStr := FormatDecimal(total)
	if rest == "" {
		return qtyStr
	}
	return qtyStr + " " + rest
}

// FormatDecimal renders a float without trailing zeros; integer-valued returns "2" not "2.0".
func FormatDecimal(d float64) string {
	if d == math.Trunc(d) {
		return strconv.FormatFloat(d, 'f', 0, 64)
	}
	return strconv.FormatFloat(d, 'f', -1, 64)
}

// FormatFraction renders a decimal as a mixed unicode fraction when close to a common value.
// 0.5 -> "½"; 1.5 -> "1½"; 0.6 -> "0.6" (no close match).
func FormatFraction(d float64) string {
	if d == 0 {
		return "0"
	}
	neg := d < 0
	if neg {
		d = -d
	}
	whole := math.Floor(d)
	frac := d - whole
	glyph := ""
	for _, f := range fractionMap {
		if math.Abs(frac-f.decimal) < fractionTolerance {
			glyph = f.glyph
			break
		}
	}
	sign := ""
	if neg {
		sign = "-"
	}
	switch {
	case whole == 0 && glyph == "":
		return sign + FormatDecimal(d)
	case whole == 0:
		return sign + glyph
	case glyph == "":
		return sign + FormatDecimal(whole)
	default:
		return sign + FormatDecimal(whole) + glyph
	}
}

// FormatIngredient swaps a leading decimal/fraction quantity in an ingredient string
// for a pretty unicode fraction. Non-numeric strings pass through untouched.
func FormatIngredient(s string) string {
	parts := strings.Fields(s)
	if len(parts) == 0 {
		return s
	}
	var total float64
	consumed := 0
	for _, p := range parts {
		v, ok := parseNumericToken(p)
		if !ok {
			break
		}
		total += v
		consumed++
	}
	if consumed == 0 {
		return s
	}
	rest := strings.Join(parts[consumed:], " ")
	if rest == "" {
		return FormatFraction(total)
	}
	return FormatFraction(total) + " " + rest
}
