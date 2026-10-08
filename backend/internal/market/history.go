package market

import (
	"strings"
	"time"
)

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

// Candle is one candlestick: the first, highest, lowest and last price in a
// time slot, and the ADA traded in it.
type Candle struct {
	Time   int64   `json:"time"`
	Open   float64 `json:"open"`
	High   float64 `json:"high"`
	Low    float64 `json:"low"`
	Close  float64 `json:"close"`
	Volume float64 `json:"volume"`
}

// Range is a chart time range with the spacing of its line points and candles.
type Range struct {
	Key            string
	Span           time.Duration
	Interval       time.Duration // line chart
	CandleInterval time.Duration // candlestick chart
}

// RangeList is every chart range, shortest first. A range can be shown once
// that much history has been loaded.
var RangeList = []Range{
	{Key: "24H", Span: 24 * time.Hour, Interval: 5 * time.Minute, CandleInterval: 15 * time.Minute},
	{Key: "7D", Span: 7 * 24 * time.Hour, Interval: time.Hour, CandleInterval: time.Hour},
	{Key: "30D", Span: 30 * 24 * time.Hour, Interval: 4 * time.Hour, CandleInterval: 4 * time.Hour},
}

// RangeByKey finds a range by its key ("24H", "7D", "30D"), ignoring case.
func RangeByKey(key string) (Range, bool) {
	for _, r := range RangeList {
		if strings.EqualFold(r.Key, key) {
			return r, true
		}
	}
	return Range{}, false
}

// MaxHistoryDays is the longest history that can be loaded.
const MaxHistoryDays = 30

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

// candles groups swaps into candlesticks between from and to. A slot with no
// swaps is a flat candle at the previous close, and each candle opens at the
// previous close so the candles join up.
func candles(points []Point, from, to time.Time, interval time.Duration) []Candle {
	step := int64(interval / time.Second)
	start := from.Unix() / step * step
	end := to.Unix()
	if start > end {
		return nil
	}
	n := int((end-start)/step) + 1
	out := make([]Candle, n)
	price := priceAt(points, start-1)

	i := 0
	for i < len(points) && points[i].Time < start {
		i++
	}
	for b := 0; b < n; b++ {
		bStart := start + int64(b)*step
		bEnd := bStart + step
		c := Candle{Time: bStart, Open: price, High: price, Low: price, Close: price}
		for i < len(points) && points[i].Time < bEnd {
			p := points[i].Price
			c.High = max(c.High, p)
			c.Low = min(c.Low, p)
			c.Close = p
			c.Volume += points[i].VolumeADA
			i++
		}
		price = c.Close
		out[b] = c
	}
	return out
}
