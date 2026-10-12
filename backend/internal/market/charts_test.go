// Tests for chart points, candles and price formatting.
package market

import (
	"testing"
	"time"
)

// TestBucketCarriesPriceForward checks that chart steps with no trade keep the
// last price, and that volume is added up per step.
func TestBucketCarriesPriceForward(t *testing.T) {
	base := time.Unix(1_000_200, 0) // a multiple of 300 seconds
	pts := []Point{
		{Time: base.Unix() - 1000, Price: 1},
		{Time: base.Unix() + 310, Price: 2, VolumeADA: 5},
		{Time: base.Unix() + 320, Price: 3, VolumeADA: 7},
	}
	got := bucket(pts, base, base.Add(15*time.Minute), 5*time.Minute)
	want := []ChartPoint{
		{Time: base.Unix(), Price: 1},
		{Time: base.Unix() + 300, Price: 3, Volume: 12},
		{Time: base.Unix() + 600, Price: 3},
		{Time: base.Unix() + 900, Price: 3},
	}
	if len(got) != len(want) {
		t.Fatalf("got %v", got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("point %d: got %v, want %v", i, got[i], want[i])
		}
	}
}

// TestFormatADA checks prices are shown with 4 significant digits.
func TestFormatADA(t *testing.T) {
	cases := map[float64]string{0.002657024: "₳0.002657", 0.1934: "₳0.1934", 1.5: "₳1.5000", 0.00000123456: "₳0.000001235"}
	for in, want := range cases {
		if got := FormatADA(in); got != want {
			t.Errorf("FormatADA(%v) = %s, want %s", in, got, want)
		}
	}
}

// TestCandles checks open, high, low and close for each candle.
func TestCandles(t *testing.T) {
	base := time.Unix(1_000_800, 0) // a multiple of 900 seconds
	pts := []Point{
		{Time: base.Unix() - 100, Price: 10},
		{Time: base.Unix() + 60, Price: 12, VolumeADA: 5},
		{Time: base.Unix() + 120, Price: 8, VolumeADA: 3},
		{Time: base.Unix() + 600, Price: 9, VolumeADA: 1},
	}
	got := candles(pts, base, base.Add(30*time.Minute), 15*time.Minute)
	want := []Candle{
		{Time: base.Unix(), Open: 10, High: 12, Low: 8, Close: 9, Volume: 9},
		{Time: base.Unix() + 900, Open: 9, High: 9, Low: 9, Close: 9},
		{Time: base.Unix() + 1800, Open: 9, High: 9, Low: 9, Close: 9},
	}
	if len(got) != len(want) {
		t.Fatalf("got %+v", got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("candle %d: got %+v, want %+v", i, got[i], want[i])
		}
	}
}
