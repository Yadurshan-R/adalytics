// Tests for calendar candles (1W, 1M): weeks start on Monday and months on
// the 1st, at 00:00 UTC, and a week or month that began before the history
// is left out instead of shown with its first days missing.
package market

import (
	"testing"
	"time"
)

// utc is a time on the given day and hour, in UTC.
func utc(y int, m time.Month, d, h int) time.Time { return time.Date(y, m, d, h, 0, 0, 0, time.UTC) }

func TestWeekAndMonthStart(t *testing.T) {
	// Sunday 8 Nov 2026 belongs to the week of Monday 2 Nov; Monday 9 Nov starts its own.
	if got := weekStart(utc(2026, 11, 8, 23)); !got.Equal(utc(2026, 11, 2, 0)) {
		t.Errorf("week of Sun 8 Nov starts %v, want Mon 2 Nov", got)
	}
	if got := weekStart(utc(2026, 11, 9, 0)); !got.Equal(utc(2026, 11, 9, 0)) {
		t.Errorf("week of Mon 9 Nov starts %v, want Mon 9 Nov", got)
	}
	if got := monthStart(utc(2026, 11, 30, 12)); !got.Equal(utc(2026, 11, 1, 0)) {
		t.Errorf("month of 30 Nov starts %v, want 1 Nov", got)
	}
}

func TestCandleStartsCalendar(t *testing.T) {
	from, to := utc(2026, 10, 10, 6), utc(2026, 11, 9, 2) // 30 days of history
	week, _ := CandleSizeByKey("1W")
	month, _ := CandleSizeByKey("1M")

	weeks := candleStarts(week, from, to)
	want := []time.Time{utc(2026, 10, 12, 0), utc(2026, 10, 19, 0), utc(2026, 10, 26, 0), utc(2026, 11, 2, 0), utc(2026, 11, 9, 0)}
	if len(weeks) != len(want) {
		t.Fatalf("got %d weeks, want %d (the week of 5 Oct began before the history)", len(weeks), len(want))
	}
	for i, w := range want {
		if weeks[i] != w.Unix() {
			t.Errorf("week %d starts %v, want %v", i, time.Unix(weeks[i], 0).UTC(), w)
		}
	}

	months := candleStarts(month, from, to)
	if len(months) != 1 || months[0] != utc(2026, 11, 1, 0).Unix() {
		t.Errorf("months %v, want only November (October began before the history)", months)
	}

	// History starting exactly on the 1st keeps that month.
	if got := candleStarts(month, utc(2026, 10, 1, 0), to); len(got) != 2 {
		t.Errorf("from 1 Oct: %d months, want October and November", len(got))
	}
}

func TestWeekCandleCountsItsWholeWeek(t *testing.T) {
	// Swaps on Monday 2 Nov, Sunday 8 Nov (both week 2 Nov) and Monday 9 Nov.
	points := []Point{
		{Time: utc(2026, 11, 2, 1).Unix(), Price: 1.0, VolumeADA: 10},
		{Time: utc(2026, 11, 8, 23).Unix(), Price: 1.5, VolumeADA: 20},
		{Time: utc(2026, 11, 9, 1).Unix(), Price: 1.2, VolumeADA: 5},
	}
	week, _ := CandleSizeByKey("1W")
	to := utc(2026, 11, 9, 2)
	got := candlesAt(points, candleStarts(week, utc(2026, 11, 1, 0), to), to)
	if len(got) != 2 {
		t.Fatalf("got %d candles, want 2", len(got))
	}
	if c := got[0]; c.Close != 1.5 || c.High != 1.5 || c.Volume != 30 {
		t.Errorf("week of 2 Nov: %+v, want close 1.5, high 1.5, volume 30", c)
	}
	if c := got[1]; c.Open != 1.5 || c.Close != 1.2 || c.Volume != 5 {
		t.Errorf("week of 9 Nov: %+v, want open 1.5 (last close), close 1.2, volume 5", c)
	}
}
