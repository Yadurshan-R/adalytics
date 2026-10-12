package market

import (
	"fmt"
	"math"
	"math/big"
	"strings"
)

// scaled turns an amount in the smallest unit into whole units, e.g.
// 1,500,000 lovelace with 6 decimals is 1.5 ADA.
func scaled(n *big.Int, decimals int) float64 {
	d := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil)
	f, _ := new(big.Rat).SetFrac(n, d).Float64()
	return f
}

// plural returns "s" unless n is 1, for log messages like "2 swaps".
func plural(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}

// FormatADA shows a price with 4 significant digits, e.g. ₳0.002657.
func FormatADA(v float64) string {
	if v <= 0 {
		return "₳0"
	}
	if v >= 1 {
		return fmt.Sprintf("₳%.4f", v)
	}
	decimals := 3 - int(math.Floor(math.Log10(v)))
	return fmt.Sprintf("₳%.*f", decimals, v)
}

// change describes a price move for the log, e.g. "▲ 2.10%".
func change(prev, now float64) string {
	switch {
	case prev == 0:
		return ""
	case now > prev:
		return fmt.Sprintf("▲ %.2f%%", (now/prev-1)*100)
	case now < prev:
		return fmt.Sprintf("▼ %.2f%%", (1-now/prev)*100)
	default:
		return "= 0.00%"
	}
}

// group formats a number as 1,977,518.
func group(v float64) string {
	s := fmt.Sprintf("%.0f", v)
	var b strings.Builder
	for i, c := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			b.WriteByte(',')
		}
		b.WriteRune(c)
	}
	return b.String()
}
