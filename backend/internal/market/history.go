package market

import "time"

// Point is the pool price right after one swap.
type Point struct {
	Time      int64   // unix seconds
	Price     float64 // ADA per token
	VolumeADA float64 // ADA traded in this transaction (0 for deposits/withdrawals)
}

// ChartPoint is one evenly spaced point on a chart.
type ChartPoint struct {
	Time   int64   `json:"time"`
	Price  float64 `json:"price"`
	Volume float64 `json:"volume"`
}

// Range is a chart time range and its point spacing.
type Range struct {
	Key      string
	Span     time.Duration
	Interval time.Duration
}

// Ranges the chart endpoint serves. 7D is added once 7 days are loaded.
var Ranges = map[string]Range{
	"24H": {Key: "24H", Span: 24 * time.Hour, Interval: 5 * time.Minute},
}

// HistorySpan is how much history is loaded at startup and kept in memory.
const HistorySpan = 24 * time.Hour

// priceAt returns the price at time t: the last swap at or before t. If every
// swap is after t, the first known price is used.
func priceAt(points []Point, t int64) float64 {
	if len(points) == 0 {
		return 0
	}
	p := points[0].Price
	for _, pt := range points {
		if pt.Time > t {
			break
		}
		p = pt.Price
	}
	return p
}

// bucket turns swaps into evenly spaced chart points between from and to.
// Each point carries the last price in its interval (or the previous price if
// nothing traded) and the ADA volume traded in it.
func bucket(points []Point, from, to time.Time, interval time.Duration) []ChartPoint {
	step := int64(interval / time.Second)
	start := from.Unix() / step * step
	end := to.Unix()
	if start > end {
		return nil
	}
	n := int((end-start)/step) + 1
	out := make([]ChartPoint, n)
	price := priceAt(points, start-1)

	i := 0
	for i < len(points) && points[i].Time < start {
		i++
	}
	for b := 0; b < n; b++ {
		bStart := start + int64(b)*step
		bEnd := bStart + step
		vol := 0.0
		for i < len(points) && points[i].Time < bEnd {
			price = points[i].Price
			vol += points[i].VolumeADA
			i++
		}
		out[b] = ChartPoint{Time: bStart, Price: price, Volume: vol}
	}
	// The last point shows the current price at the current time.
	if len(points) > 0 {
		out[n-1].Price = points[len(points)-1].Price
	}
	return out
}

// changePct is the percentage change from the price at time t to now.
func changePct(points []Point, t int64) float64 {
	if len(points) == 0 {
		return 0
	}
	then := priceAt(points, t)
	now := points[len(points)-1].Price
	if then == 0 {
		return 0
	}
	return (now/then - 1) * 100
}

func volumeSince(points []Point, t int64) float64 {
	v := 0.0
	for i := len(points) - 1; i >= 0 && points[i].Time > t; i-- {
		v += points[i].VolumeADA
	}
	return v
}

// trim drops swaps older than the kept history, but keeps the last one before
// the cut so the price at the start of the window stays known.
func trim(points []Point, now time.Time) []Point {
	cut := now.Add(-HistorySpan).Unix()
	i := 0
	for i+1 < len(points) && points[i+1].Time <= cut {
		i++
	}
	return points[i:]
}
