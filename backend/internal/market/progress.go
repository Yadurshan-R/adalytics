package market

import (
	"time"

	"cardano-analytics/internal/koios"
)

// bootProgress tracks the four steps of the first load, until the list opens.
type bootProgress struct {
	step       int // 0 listing trades, 1 reading the newest hour, 2 other tokens, 3 token details
	found      int // transactions listed (then: in the newest hour)
	read       int // transactions of the newest hour read
	quietDone  int // quiet tokens checked
	quietTotal int
}

// RangeStatus says how much of a chart range's history is loaded.
type RangeStatus struct {
	Key      string `json:"key"`
	Progress int    `json:"progress"` // 0 to 100
	Ready    bool   `json:"ready"`
}

// LoadStep is one step of the first load, for the loading screen.
type LoadStep struct {
	Key   string `json:"key"`   // trades_found, reading, quiet, details
	State string `json:"state"` // done, active or todo
	Done  int    `json:"done"`
	Total int    `json:"total,omitempty"`
}

// Loading is the progress of the first load, shown until the list is ready.
type Loading struct {
	Percent        int        `json:"percent"`
	Steps          []LoadStep `json:"steps"`
	ElapsedSeconds int        `json:"elapsed_seconds"`        // since the first load started
	LeftSeconds    int        `json:"left_seconds,omitempty"` // estimate; 0 while unknown
	Requests       int64      `json:"requests"`               // Koios requests sent so far
	Activity       []Activity `json:"activity"`               // what the backend did, newest first
}

// History says how much history is loaded.
type History struct {
	Loading     *Loading      `json:"loading,omitempty"` // only during the first load
	Days        int           `json:"days"`              // days that will be loaded (HISTORY_DAYS)
	LoadedHours float64       `json:"loaded_hours"`      // hours loaded so far
	Done        bool          `json:"done"`
	Ranges      []RangeStatus `json:"ranges"` // only ranges within Days
}

// History returns how much history is loaded and which chart ranges are ready.
func (t *Tracker) History() History {
	t.mu.RLock()
	defer t.mu.RUnlock()
	h := History{Days: int(t.span / (24 * time.Hour)), Done: t.historyDone}
	if !t.loaded {
		h.Loading = t.bootLoading()
	}
	if t.loaded {
		h.LoadedHours = float64(int(t.now().Sub(t.coveredFrom).Hours()*10)) / 10
	}
	for _, r := range RangeList {
		if r.Span > t.span {
			break
		}
		st := RangeStatus{Key: r.Key, Ready: t.rangeReady(r)}
		if st.Ready {
			st.Progress = 100
		} else if t.loaded {
			st.Progress = min(99, int(t.now().Sub(t.coveredFrom)*100/r.Span))
		}
		h.Ranges = append(h.Ranges, st)
	}
	return h
}

// setBoot changes the first load's progress safely while other code reads it.
func (t *Tracker) setBoot(f func(b *bootProgress)) {
	t.mu.Lock()
	f(&t.boot)
	t.mu.Unlock()
}

// bootLoading turns the first load's progress into steps and one percentage:
// listing 0-20%, the newest hour 20-35%, other tokens' pools 35-90%,
// details 90-100%.
// Call with t.mu held.
func (t *Tracker) bootLoading() *Loading {
	b := t.boot
	state := func(step int) string {
		switch {
		case b.step > step:
			return "done"
		case b.step == step:
			return "active"
		}
		return "todo"
	}
	frac := func(done, total int) float64 {
		if total <= 0 {
			return 0
		}
		return min(1, float64(done)/float64(total))
	}
	var pct float64
	switch b.step {
	case 0:
		pct = 19 * frac(b.found, 3*koios.AddressTxsPageSize) // about 3 pages a day
	case 1:
		pct = 20 + 15*frac(b.read, b.found)
	case 2:
		pct = 35 + 55*frac(b.quietDone, b.quietTotal)
	default:
		pct = 97
	}
	l := &Loading{
		Percent:  int(pct),
		Activity: t.recentActivity(),
		Steps: []LoadStep{
			{Key: "trades_found", State: state(0), Done: b.found},
			{Key: "reading", State: state(1), Done: b.read, Total: b.found},
			{Key: "quiet", State: state(2), Done: b.quietDone, Total: b.quietTotal},
			{Key: "details", State: state(3)},
		},
	}
	if !t.bootStart.IsZero() {
		elapsed := t.now().Sub(t.bootStart)
		l.ElapsedSeconds = int(elapsed / time.Second)
		l.LeftSeconds = timeLeft(elapsed, l.Percent)
	}
	if t.client != nil {
		l.Requests = t.client.Requests()
	}
	return l
}

// rangeReady reports whether the whole range is loaded. Call with t.mu held.
func (t *Tracker) rangeReady(r Range) bool {
	if !t.loaded || r.Span > t.span {
		return false
	}
	return t.historyDone || t.now().Sub(t.coveredFrom) >= r.Span
}
